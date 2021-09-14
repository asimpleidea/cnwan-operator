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

package reflect

type CloudMetadataRetriever interface {
	GetPlatformName() (string, error)
	GetNetworkName() (string, error)
	GetSubNetworkName() (string, error)
	GetRegion() (string, error)
	GetGCPProject() (string, error)
}

type CloudMetadata struct {
}

func NewCloudMetadataRetriever() (*CloudMetadata, error) {
	return nil, nil
}

func (c *CloudMetadata) GetRegion() (string, error) {
	return "", nil
}

func (c *CloudMetadata) GetGCPProject() (string, error) {
	return "", nil
}

func (c *CloudMetadata) GetNetworkName() (string, error) {
	return "", nil
}

func (c *CloudMetadata) GetSubNetworkName() (string, error) {
	return "", nil
}
