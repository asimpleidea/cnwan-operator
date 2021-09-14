// Copyright © 2021 Cisco
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// All rights reserved.

package command

import (
	"context"
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/CloudNativeSDWAN/cnwan-operator/pkg/cli/option"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newEtcdCommand(opcli *option.Operator) *cobra.Command {
	etcdOpts := &option.Etcd{}
	settingsPath := ""
	settingsCfgMap := ""
	credentialsSecret := ""
	inputPassword := false
	username := ""
	password := ""

	// -------------------------------
	// Define root command
	// -------------------------------

	// TODO: write this
	cmd := &cobra.Command{
		Use: `etcd [--help, -h]`,

		Short: `TODO etcd`,

		Long: `TODO etcd`,

		Example: "etcd --prefix my-service-registry",

		PreRunE: func(cmd *cobra.Command, args []string) error {
			namespace, _ := cmd.Flags().GetString("namespace")
			var err error
			username, password, err = getEtcdUsernameAndPassword(inputPassword, username, credentialsSecret, namespace, &stdinPasswordReader{})
			if err != nil {
				return err
			}

			dataBytes, err := loadExistingSettings(settingsPath, settingsCfgMap, defaultEtcdConfigMapName, namespace)
			if err != nil {
				return err
			}

			var fromFile option.Etcd
			if len(dataBytes) > 0 {
				if err := yaml.Unmarshal(dataBytes, &fromFile); err != nil {
					return fmt.Errorf("error while trying to decode settings file: %s", err.Error())
				}
			}

			etcdOpts = mergeEtcdSettings(etcdOpts, &fromFile)
			return nil
		},

		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("%+v %s %s\n", etcdOpts, username, password)
			return nil
		},
	}

	// -------------------------------
	// Define global persistent flags
	// -------------------------------

	f := cmd.Flags()
	f.StringVar(&settingsPath, "settings-path", "", "path to etcd settings file.")
	f.StringVar(&settingsCfgMap, "settings-configmap", "", "the name of the configmap containing etcd settings.")
	f.StringVar(&credentialsSecret, "credentials-secret", "", "the name of the Kubernetes secret containing etcd credentials.")
	f.StringVar(&etcdOpts.Prefix, "prefix", defaultEtcdPrefix, "the prefix where to put all keys.")
	f.StringArrayVar(&etcdOpts.Endpoints, "endpoints", []string{}, "a list of etcd endpoints.")
	f.StringVarP(&username, "username", "u", "", "the username to login as.")
	f.BoolVarP(&inputPassword, "password", "p", false, "the password to use for the provided user.")

	return cmd
}

func mergeEtcdSettings(fromCli, fromFile *option.Etcd) *option.Etcd {
	prefix := func() string {
		if fromCli.Prefix != "" {
			return fromCli.Prefix
		}

		if fromFile != nil && fromFile.Prefix != "" {
			return fromFile.Prefix
		}

		return ""
	}()
	switch prefix {
	case "":
		prefix = defaultEtcdPrefix
	case "/":
		prefix = "/"
	default:
		prefix = fmt.Sprintf("/%s/", strings.Trim(prefix, "/"))
	}

	endpoints := func() []string {
		eps := []string{}
		if len(fromCli.Endpoints) > 0 {
			eps = fromCli.Endpoints
		} else {
			if fromFile != nil {
				eps = fromFile.Endpoints
			}
		}

		// At this point just return a default one and make peace with it.
		// This most definitely won't work if we are inside Kubernetes, but
		// there's need to throw an error here as it will just fail to connect
		// anyways.
		if len(eps) == 0 {
			eps = []string{"localhost:2379"}
		}

		endpSet := map[string]bool{}
		for _, ep := range eps {
			if _, exists := endpSet[ep]; !exists {
				endpSet[ep] = true
			}
		}

		eps = []string{}
		for key := range endpSet {
			eps = append(eps, key)
		}
		return eps
	}()

	return &option.Etcd{
		Prefix:    prefix,
		Endpoints: endpoints,
	}
}

type passwordReader interface {
	readPassword() (string, error)
}

type stdinPasswordReader struct{}

func (s *stdinPasswordReader) readPassword() (string, error) {
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	return string(bytePassword), nil
}

func getEtcdUsernameAndPassword(inputPassword bool, username, credSecret, namespace string, passReader passwordReader) (string, string, error) {
	password := ""
	if inputPassword {
		if runningInK8s() {
			return "", "", fmt.Errorf("please use Kubernetes Secrets to define credentials")
		}

		if username == "" {
			// TODO: log.Warning and use root
			username = "root"
		}

		for password == "" {
			fmt.Printf(`please insert password for %s: `, username)
			pass, err := passReader.readPassword()
			if err != nil {
				return "", "", err
			}
			password = pass
			fmt.Println()
		}
	}

	if username != "" {
		return username, password, nil
	}

	// Try to get credentials from somewhere else in case they were not
	// provided via CLI.
	if credSecret != "" {
		return loadEtcdUsernameAndPasswordFromSecret(namespace, credSecret)
	}

	if runningInK8s() {
		// This step is done in case we are running inside k8s but
		// user did not provide a name for the credentials secret.
		// We will try to retrieve it anyway with the default
		// secret name and if we do find it that's alright, but if
		// we don't then it's not an error: maybe the user is
		// allowing unauthenticated users to access etcd after all.
		if username, password, err := loadEtcdUsernameAndPasswordFromSecret(namespace, defaultEtcdCredentialsSecretName); err == nil {
			return username, password, nil
		}
	}

	return "", "", nil
}

func loadEtcdUsernameAndPasswordFromSecret(namespace, name string) (string, string, error) {
	kcli, err := getKubernetesClientSet()
	if err != nil {
		return "", "", err
	}

	ctx, canc := context.WithTimeout(context.Background(), time.Minute)
	defer canc()

	secret, err := kcli.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}

	username, exists := secret.Data["username"]
	if !exists {
		return "", "", fmt.Errorf(`secret %s/%s doesn't have a username`, namespace, name)
	}

	password, exists := secret.Data["password"]
	if !exists {
		// This might be intentional.
		// TODO: log.warn about this
		_ = exists
	}

	return string(username), string(password), nil
}
