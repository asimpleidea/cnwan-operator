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
	"os"
	"time"

	gcpmetadata "cloud.google.com/go/compute/metadata"
	"github.com/CloudNativeSDWAN/cnwan-operator/pkg/cli/option"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	gccontainer "google.golang.org/api/container/v1"
	gcoption "google.golang.org/api/option"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	log zerolog.Logger
)

// NewRootCommand defines the root command, including its flags definition
// and subcommands, and returns it so that it could be used.
func NewRootCommand() *cobra.Command {
	opSettings := &option.Operator{}
	namespace := ""
	settingsPath := ""
	settingsCfgMap := ""
	verbosity := 1

	servAccPath := ""
	serviceAccountSecret := ""

	// -------------------------------
	// Define root command
	// -------------------------------

	// TODO: write this
	cmd := &cobra.Command{
		Use: `cnwan-operator etcd|service-directory [OPTIONS] [--help, -h]`,

		Short: `TODO root`,

		Long: `TODO root`,

		Example: "cnwan-operator etcd --prefix my-service-registry",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logLevels := []zerolog.Level{zerolog.DebugLevel, zerolog.InfoLevel, zerolog.ErrorLevel}
			log = zerolog.New(os.Stderr).Level(logLevels[1]).With().Timestamp().Logger()

			if verbosity < 0 && verbosity > len(logLevels) {
				log.Error().
					Int("verbosity-level", verbosity).
					Int("default-verbosity", 1).
					Msg("invalid verbosity level provided, using default...")
				verbosity = 1
			}

			log = log.Level(logLevels[verbosity]).With().Logger()

			if namespace == "" {
				kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
					clientcmd.NewDefaultClientConfigLoadingRules(),
					&clientcmd.ConfigOverrides{},
				)

				ns, _, err := kubeconfig.Namespace()
				if err != nil {
					ns = "cnwan-operator-system"
				}

				namespace = ns
			}
			log.Info().Str("namespace", namespace).Msg("retrieved namespace")

			// ---------------------------
			// Get options from file/secret, if provided
			// ---------------------------

			var fromFile *option.Operator
			dataBytes, err := loadExistingSettings(settingsPath, settingsCfgMap, defaultOperatorConfigMapName, namespace)
			if err != nil {
				return err
			}

			if len(dataBytes) > 0 {
				var _fromFile option.Operator
				if err := yaml.Unmarshal(dataBytes, &_fromFile); err != nil {
					return fmt.Errorf("error while trying to decode settings file: %s", err.Error())
				}

				fromFile = &_fromFile
				if !cmd.Flag("watch-all-namespaces").Changed {
					opSettings.WatchAllNamespaces = fromFile.WatchAllNamespaces
				}
			}

			opSettings, err = validateAndMergeOperatorSettings(opSettings, fromFile)
			if err != nil {
				return err
			}

			// ---------------------------
			// Get network data from cloud platform, if enabled
			// ---------------------------

			if opSettings.CloudMetadata.Network != "auto" && opSettings.CloudMetadata.SubNetwork != "auto" {
				return nil
			}

			// We only support GCE for the moment, and in order to get network
			// data we need a service account OR scopes injected into the
			// VM that is hosting us.

			serviceAccount, err := getGCPServiceAccount(servAccPath, serviceAccountSecret, namespace)
			if err != nil {
				return err
			}

			networkName, subNetworkName, err := getNetworkNamesFromGCE(serviceAccount, &gcpMetadata{})
			if err != nil {
				return err
			}

			if opSettings.CloudMetadata.Network == "auto" {
				opSettings.CloudMetadata.Network = networkName
			}

			if opSettings.CloudMetadata.SubNetwork == "auto" {
				opSettings.CloudMetadata.SubNetwork = subNetworkName
			}

			return nil
		},

		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("%+v\n", opSettings)
			return fmt.Errorf("please provide a service registry")
		},
	}

	// -------------------------------
	// Define global persistent flags
	// -------------------------------

	pf := cmd.PersistentFlags()
	pf.StringVar(&settingsPath, "operator-settings-path", settingsPath, "the path to the yaml file containing settings for the operator.")
	pf.StringVar(&settingsCfgMap, "operator-settings-configmap", settingsCfgMap, "the name of the configmap with the settings for the operator. Default: cnwan-operator-settings.")

	pf.BoolVar(&opSettings.WatchAllNamespaces, "watch-all-namespaces", false, "if set, the operator will watch all namespaces, unless they are explitly disabled.")
	pf.StringSliceVar(&opSettings.ServiceFilters.Labels, "service-labels", []string{}, "the service labels to watch and register.")
	pf.StringSliceVar(&opSettings.ServiceFilters.Annotations, "service-annotations", []string{}, "the service annotations to watch and register.")
	pf.StringVar(&opSettings.CloudMetadata.Network, "network-name", "", "the name of the network to where the services are deployed to. Write 'auto' to detect this automatically.")
	pf.StringVar(&opSettings.CloudMetadata.SubNetwork, "subnetwork-name", "", "the name of the sub-network to where the services are deployed to. Write 'auto' to detect this automatically.")

	pf.StringVar(&namespace, "namespace", "", "the namespace where the operator is running in or where the ConfigMaps or Secrets should be deployed to.")
	pf.IntVarP(&verbosity, "verbosity", "v", verbosity, "log verbosity level. Provide a value between 0 and 3, where 0 is the most verbose and 3 only logs fatal errors.")

	pf.StringVar(&servAccPath, "gcp-service-account-path", "", "the path to the google service account, this is used in case you need network and/or subnetwork names automatically retrieved from GCP.")
	pf.StringVar(&serviceAccountSecret, "gcp-service-account-secret", "", "the name of the secret holding Google service account.")

	// This will be enabled on next versions
	pf.MarkHidden("service-labels")

	// -------------------------------
	// Define subcommands
	// -------------------------------

	cmd.AddCommand(newEtcdCommand(opSettings))
	cmd.AddCommand(newServiceDirectoryCommand(opSettings))

	return cmd
}

// validateAndMergeOperatorSettings validates the settings for the operator by
// merging the options provided from CLI and those -- optionally -- provided
// from file.
//
// TODO: parse the annotations and remove the ones that are irrelevant
// i.e. if this/* is provided, then remove this/that.
// This is not critical, so it will be done on another version.
//
// TODO: merge service labels as well when they are going to be enabled.
func validateAndMergeOperatorSettings(fromCli, fromFile *option.Operator) (*option.Operator, error) {
	serviceAnnotations := func() []string {
		if len(fromCli.ServiceFilters.Annotations) > 0 {
			return fromCli.ServiceFilters.Annotations
		}

		if fromFile != nil && len(fromFile.ServiceFilters.Annotations) > 0 {
			return fromFile.ServiceFilters.Annotations
		}

		return []string{}
	}()
	if len(serviceAnnotations) == 0 {
		return nil, fmt.Errorf("no annotations provided")
	}

	networkName := func() string {
		if fromCli.CloudMetadata.Network != "" {
			return fromCli.CloudMetadata.Network
		}

		if fromFile != nil && fromFile.CloudMetadata.Network != "" {
			return fromFile.CloudMetadata.Network
		}

		return ""
	}()

	subnetworkName := func() string {
		if fromCli.CloudMetadata.SubNetwork != "" {
			return fromCli.CloudMetadata.SubNetwork
		}

		if fromFile != nil && fromFile.CloudMetadata.SubNetwork != "" {
			return fromFile.CloudMetadata.SubNetwork
		}

		return ""
	}()

	return &option.Operator{
		WatchAllNamespaces: fromCli.WatchAllNamespaces,
		ServiceFilters: option.ServiceFilters{
			Annotations: serviceAnnotations,
			Labels:      []string{},
		},
		CloudMetadata: option.CloudMetadata{
			Network:    networkName,
			SubNetwork: subnetworkName,
		},
	}, nil
}

func getNetworkNamesFromGCE(sa []byte, dataRetriever gcpMetadataRetriever) (string, string, error) {
	projectID, err := dataRetriever.ProjectID()
	if err != nil {
		return "", "", err
	}

	zone, err := dataRetriever.Zone()
	if err != nil {
		return "", "", err
	}

	clusterName, err := gcpmetadata.InstanceAttributeValue("cluster-name")
	if err != nil {
		return "", "", err
	}

	ctx, canc := context.WithTimeout(context.Background(), time.Minute)
	defer canc()

	opts := []gcoption.ClientOption{gcoption.WithScopes(gccontainer.CloudPlatformScope)}
	if len(sa) > 0 {
		log.Info().Msg("sa is there")
		opts = append(opts, gcoption.WithCredentialsJSON(sa))
	}

	cli, err := gccontainer.NewService(ctx, opts...)
	if err != nil {
		return "", "", err
	}

	cluster, err := cli.Projects.Zones.Clusters.Get(projectID, zone, clusterName).Do()
	if err != nil {
		return "", "", err
	}

	return cluster.Network, cluster.Subnetwork, nil
}
