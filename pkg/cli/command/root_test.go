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
	"fmt"
	"testing"

	"github.com/CloudNativeSDWAN/cnwan-operator/pkg/cli/option"
	"github.com/stretchr/testify/assert"
)

func TestValidateAndMergeOperatorSettings(t *testing.T) {
	cases := []struct {
		fromCli  *option.Operator
		fromFile *option.Operator
		expRes   *option.Operator
		expErr   error
	}{
		{
			fromCli: &option.Operator{},
			expErr:  fmt.Errorf("no annotations provided"),
		},
		{
			fromCli: &option.Operator{
				ServiceFilters: option.ServiceFilters{
					Annotations: []string{"test"},
					Labels:      []string{},
				},
				CloudMetadata: option.CloudMetadata{
					Network:    "network",
					SubNetwork: "subnetwork",
				},
			},
			expRes: &option.Operator{
				ServiceFilters: option.ServiceFilters{
					Annotations: []string{"test"},
					Labels:      []string{},
				},
				CloudMetadata: option.CloudMetadata{
					Network:    "network",
					SubNetwork: "subnetwork",
				},
			},
		},
		{
			fromCli: &option.Operator{},
			fromFile: &option.Operator{
				ServiceFilters: option.ServiceFilters{
					Annotations: []string{"test"},
				},
				CloudMetadata: option.CloudMetadata{
					Network:    "network",
					SubNetwork: "subnetwork",
				},
			},
			expRes: &option.Operator{
				ServiceFilters: option.ServiceFilters{
					Annotations: []string{"test"},
					Labels:      []string{},
				},
				CloudMetadata: option.CloudMetadata{
					Network:    "network",
					SubNetwork: "subnetwork",
				},
			},
		},
		{
			fromCli: &option.Operator{
				ServiceFilters: option.ServiceFilters{
					Annotations: []string{"annotation"},
				},
			},
			expRes: &option.Operator{
				ServiceFilters: option.ServiceFilters{
					Annotations: []string{"annotation"},
					Labels:      []string{},
				},
			},
		},
	}

	a := assert.New(t)
	for _, currCase := range cases {
		s, err := validateAndMergeOperatorSettings(currCase.fromCli, currCase.fromFile)

		a.Equal(currCase.expRes, s)
		a.Equal(currCase.expErr, err)
	}
}
