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
package option

const (
	DefaultServiceAccountSecretName string = "service-directory-service-account"
)

type ServiceDirectory struct {
	ProjectID     string `yaml:"projectID,omitempty"`
	DefaultRegion string `yaml:"defaultRegion,omitempty"`
}

type ServiceDirectoryCli struct {
	ProjectID     string
	DefaultRegion string
}
