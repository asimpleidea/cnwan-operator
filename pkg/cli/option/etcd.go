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
	DefaultEtcdPrefix                string = "service-registry"
	DefaultEtcdConfigMapName         string = "etcd-settings"
	DefaultEtcdCredentialsSecretName string = "etcd-credentials"
)

type Etcd struct {
	Prefix    string   `yaml:"prefix,omitempty"`
	Endpoints []string `yaml:"endpoints,omitempty"`
}

type EtcdCli struct {
	Prefix    string
	Endpoints []string
	Username  string
	Password  bool

	generatedSettings *Etcd
}

func (e *EtcdCli) Settings(fromFile *Etcd) (*Etcd, error) {
	if e.generatedSettings != nil {
		return e.generatedSettings, nil
	}

	prefix := func() string {
		if len(e.Prefix) > 0 {
			return e.Prefix
		}

		if fromFile != nil {
			return fromFile.Prefix
		}

		return ""
	}()

	endpoints := func() []string {
		if len(e.Endpoints) > 1 || (len(e.Endpoints) == 1 && e.Endpoints[0] != "localhost:2379") {
			return e.Endpoints
		}

		if fromFile != nil {
			return fromFile.Endpoints
		}

		// At this point just return the default one and make peace with it.
		return []string{"localhost:2379"}
	}()

	e.generatedSettings = &Etcd{
		Prefix:    prefix,
		Endpoints: endpoints,
	}

	return e.generatedSettings, nil
}
