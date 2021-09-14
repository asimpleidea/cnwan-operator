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

// TODO: write documentation
package settings

import (
	"fmt"
	"strings"
)

type NamespaceAction string

const (
	EnabledNamespaceAction  NamespaceAction = "enabled"
	DisabledNamespaceAction NamespaceAction = "disabled"
)

type ServiceRegistryToUse string

const (
	EtcdServiceRegistry             ServiceRegistryToUse = "etcd"
	ServiceDirectoryServiceRegistry ServiceRegistryToUse = "servicedirectory"
)

type Operator struct {
	DefaultNamespaceAction NamespaceAction      `yaml:"defaultNamespaceAction,omitempty"`
	ServiceFilters         ServiceFilters       `yaml:",inline"`
	CloudMetadata          CloudMetadata        `yaml:"cloudMetadata,omitempty"`
	ServiceRegistry        ServiceRegistryToUse `yaml:"serviceRegistry"`
}

type ServiceFilters struct {
	Labels      []string `yaml:"serviceLabels,omitempty"`
	Annotations []string `yaml:"serviceAnnotations,omitempty"`
}

// CloudMetadata contains data and configuration about the cloud provider
// that is hosting the cluster, if any.
type CloudMetadata struct {
	// Network name
	Network *string `yaml:"network,omitempty"`
	// SubNetwork name
	SubNetwork *string `yaml:"subNetwork,omitempty"`
}

func ValidateOperatorSettings(opSets *Operator) error {
	defNsAction := strings.ToLower(string(opSets.DefaultNamespaceAction))

	switch defNsAction {
	case string(EnabledNamespaceAction):
		opSets.DefaultNamespaceAction = EnabledNamespaceAction
	case string(DisabledNamespaceAction):
		opSets.DefaultNamespaceAction = DisabledNamespaceAction
	default:
		return fmt.Errorf("unrecognized")
	}

	// TODO: others...

	return nil
}

// func ValidateAndResetOperatorSettings(opSets *Operator) error {
// 	defNsAction := strings.ToLower(string(opSets.DefaultNamespaceAction))
// 	if defNsAction != string(EnabledNamespaceAction) && defNsAction != string(DisabledNamespaceAction) {
// 		// TODO: log that the value is not valid and reverting to disabled
// 		opSets.DefaultNamespaceAction = DisabledNamespaceAction
// 	}

// 	if len(opSets.ServiceFilters.Annotations) > 0 {
// 		opSets.ServiceFilters.AnnotationsSet = setFromKeys(opSets.ServiceFilters.Annotations)
// 		opSets.ServiceFilters.Annotations = keysFromMap(opSets.ServiceFilters.AnnotationsSet)
// 	} else {
// 		opSets.ServiceFilters.AnnotationsSet = map[string]bool{}
// 		opSets.ServiceFilters.Annotations = []string{}
// 	}

// 	if len(opSets.ServiceFilters.Labels) > 0 {
// 		opSets.ServiceFilters.LabelsSet = setFromKeys(opSets.ServiceFilters.Labels)
// 		opSets.ServiceFilters.Labels = keysFromMap(opSets.ServiceFilters.LabelsSet)
// 	} else {
// 		opSets.ServiceFilters.LabelsSet = map[string]bool{}
// 		opSets.ServiceFilters.Labels = []string{}
// 	}

// 	if opSets.CloudMetadata != nil {
// 		if opSets.CloudMetadata.Network == nil && opSets.CloudMetadata.SubNetwork == nil {
// 			opSets.CloudMetadata = nil
// 		} else {
// 			op
// 		}
// 	}
// }

// func keysFromMap(m map[string]bool) (keys []string) {
// 	for key := range m {
// 		keys = append(keys, key)
// 	}
// 	return
// }

// func setFromKeys(arr []string) (set map[string]bool) {
// 	for _, val := range arr {
// 		set[val] = true
// 	}
// 	return
// }
