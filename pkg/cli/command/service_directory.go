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
	"time"

	gcpmetadata "cloud.google.com/go/compute/metadata"
	"github.com/CloudNativeSDWAN/cnwan-operator/pkg/cli/option"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newServiceDirectoryCommand(opcli *option.Operator) *cobra.Command {
	sdOpts := &option.ServiceDirectory{}
	settingsPath := ""
	settingsCfgMap := ""
	servAccPath := ""
	credentialsSecret := ""

	// -------------------------------
	// Define root command
	// -------------------------------

	// TODO: write this
	cmd := &cobra.Command{
		Use: `service-directory [--help, -h]`,

		Short: `TODO`,

		Long: `TODO`,

		Aliases: []string{"sd"},

		Example: "service-directory --service-account-path ./path/to/the/service-account.json",

		PreRunE: func(cmd *cobra.Command, args []string) error {
			var fromFile *option.ServiceDirectory
			namespace, _ := cmd.Flags().GetString("namespace")

			dataBytes, err := loadExistingSettings(settingsPath, settingsCfgMap, defaultServiceDirectoryConfigMapName, namespace)
			if err != nil {
				return err
			}

			if len(dataBytes) > 0 {
				var _fromFile option.ServiceDirectory
				if err := yaml.Unmarshal(dataBytes, &_fromFile); err != nil {
					return fmt.Errorf("error while trying to decode settings file: %s", err.Error())
				}

				fromFile = &_fromFile
			}

			sdOpts, err = validateAndMergeServiceDirectorySettings(sdOpts, fromFile, &gcpMetadata{})
			if err != nil {
				return err
			}

			servAccPath, _ = cmd.Flags().GetString("gcp-service-account-path")
			credentialsSecret, _ = cmd.Flags().GetString("gcp-service-account-secret")
			serviceAccount, err := getGCPServiceAccount(servAccPath, credentialsSecret, namespace)
			if err != nil {
				return err
			}
			log.Info().Msg("found service account")

			_ = serviceAccount
			return nil
		},

		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("%+v\n", sdOpts)
			return nil
		},
	}

	// -------------------------------
	// Define global persistent flags
	// -------------------------------

	f := cmd.Flags()
	f.StringVar(&settingsPath, "settings-path", "", "path to Service Directory settings file.")
	f.StringVar(&settingsCfgMap, "settings-configmap", "", "the name of the configmap containing Service Directory settings.")
	f.StringVar(&servAccPath, "service-account-path", "", "the path to the service account.")
	f.StringVar(&credentialsSecret, "credentials-secret", "", "the name of the Kubernetes secret containing Service Directory credentials.")
	f.StringVar(&sdOpts.ProjectID, "project-id", "", "the GCP project ID where Service Directory should be running.")
	f.StringVar(&sdOpts.DefaultRegion, "default-region", "", "the default region where service will be registered to.")

	return cmd
}

func validateAndMergeServiceDirectorySettings(fromCli, fromFile *option.ServiceDirectory, metadataRetriever gcpMetadataRetriever) (*option.ServiceDirectory, error) {
	projectID, err := func() (string, error) {
		if fromCli.ProjectID != "" {
			return fromCli.ProjectID, nil
		}

		if fromFile != nil {
			return fromFile.ProjectID, nil
		}

		return metadataRetriever.ProjectID()
	}()
	if err != nil {
		return nil, err
	}

	defaultRegion, err := func() (string, error) {
		if fromCli.DefaultRegion != "" {
			return fromCli.DefaultRegion, nil
		}

		if fromFile != nil {
			return fromFile.DefaultRegion, nil
		}

		zone, err := metadataRetriever.Zone()
		if err != nil {
			return "", err
		}

		i := strings.LastIndex(zone, "-")
		if i == -1 {
			return "", fmt.Errorf("unexpected zone found: %s", zone)
		}

		return zone[:i], err
	}()
	if err != nil {
		return nil, err
	}

	return &option.ServiceDirectory{
		ProjectID:     projectID,
		DefaultRegion: defaultRegion,
	}, nil
}

type gcpMetadataRetriever interface {
	ProjectID() (string, error)
	Zone() (string, error)
	// Networks() (string, string, error)
}

type gcpMetadata struct{}

func (g *gcpMetadata) Zone() (string, error) {
	if !gcpmetadata.OnGCE() {
		return "", fmt.Errorf("cannot retrieve region: not running in GCP")
	}

	return gcpmetadata.Zone()
}

func (g *gcpMetadata) ProjectID() (string, error) {
	if !gcpmetadata.OnGCE() {
		return "", fmt.Errorf("cannot retrieve project id: not running in GCP")
	}

	return gcpmetadata.ProjectID()
}

func loadGoogleServiceAccountSecret(namespace, name string) ([]byte, error) {
	kcli, err := getKubernetesClientSet()
	if err != nil {
		return nil, err
	}

	ctx, canc := context.WithTimeout(context.Background(), time.Minute)
	defer canc()

	secret, err := kcli.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	for _, d := range secret.Data {
		return d, nil
	}

	return nil, fmt.Errorf(`secret %s/%s has no data`, namespace, name)
}
