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

package option

type (
	// NamespaceAction is a type that defines the action that the operator
	// needs to take when it can't find the label "operator.cnwan.io/enabled".
	// Read the documentation to learn how this value affects the behavior of
	// the operator.
	NamespaceAction string
)

const (
	// EnabledNamespaceAction instructs the operator to pretend that all
	// namespaces have the label "operator.cnwan.io/enabled" with value 'yes',
	// unless they explicitly have the value 'no', in which case the operator
	// will ignore it.
	EnabledNamespaceAction NamespaceAction = "enabled"
	// DisabledNamespaceAction instructs the operator to pretend that all
	// namespaces have the label "operator.cnwan.io/enabled" with value 'no',
	// which means that they will be ignored by the operator, unless they
	// explicitly have the value 'yes', in which case the operator watch it.
	DisabledNamespaceAction NamespaceAction = "disabled"
	// DefaultConfigmapName is the default name of the ConfigMap containing
	// settings for the operator. This can be overidden by the CLI.
	DefaultConfigmapName string = "cnwan-operator-settings"
	// DefaultNamespace is the default name of the namespace where the
	// operator is running in Kubernetes. This can be overidden by the CLI.
	DefaultNamespace string = "cnwan-operator-system"
	// DefaultNamespaceAction is the namespace action to use in case no value
	// is provided.
	DefaultNamespaceAction NamespaceAction = "enabled"
)

// Operator contains settings for the operator.
type Operator struct {
	WatchAllNamespaces bool           `yaml:"watchAllNamespaces"`
	ServiceFilters     ServiceFilters `yaml:",inline"`
	CloudMetadata      CloudMetadata  `yaml:"cloudMetadata,omitempty"`
}

type ServiceFilters struct {
	Labels      []string `yaml:"serviceLabels,omitempty"`
	Annotations []string `yaml:"serviceAnnotations,omitempty"`
}

// CloudMetadata contains data and configuration about the cloud provider
// that is hosting the cluster, if any.
type CloudMetadata struct {
	// Network name.
	Network string `yaml:"network,omitempty"`
	// SubNetwork name.
	SubNetwork string `yaml:"subNetwork,omitempty"`
}
