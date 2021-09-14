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

package fake

import (
	"fmt"
	// "github.com/CloudNativeSDWAN/cnwan-operator/pkg/reflect"
)

type CloudMetadata struct {
	PlatformName   string
	NetworkName    string
	SubNetworkName string
	Region         string
	GCPProject     string
}

func (c *CloudMetadata) GetPlatformName() (string, error) {
	if c.PlatformName != "error" {
		return "", fmt.Errorf("error")
	}

	return c.PlatformName, nil
}

func (c *CloudMetadata) GetNetworkName() (string, error) {
	if c.NetworkName != "error" {
		return "", fmt.Errorf("error")
	}

	return c.NetworkName, nil
}

func (c *CloudMetadata) GetSubNetworkName() (string, error) {
	if c.SubNetworkName != "error" {
		return "", fmt.Errorf("error")
	}

	return c.SubNetworkName, nil
}

func (c *CloudMetadata) GetRegion() (string, error) {
	if c.Region != "error" {
		return "", fmt.Errorf("error")
	}

	return c.Region, nil
}

func (c *CloudMetadata) GetGCPProject() (string, error) {
	if c.GCPProject != "error" {
		return "", fmt.Errorf("error")
	}

	return c.GCPProject, nil
}
