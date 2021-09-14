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
	"errors"
	"fmt"
	"testing"

	"github.com/CloudNativeSDWAN/cnwan-operator/pkg/cli/option"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

type fakeGcpMetadataRetriever struct {
	region  string
	project string
	inGCP   bool
}

func (f *fakeGcpMetadataRetriever) Zone() (string, error) {
	if !f.inGCP {
		return "", fmt.Errorf("cannot retrieve region: not running in GCP")
	}
	if f.region == "error" {
		return "", fmt.Errorf("error")
	}

	return f.region, nil
}

func (f *fakeGcpMetadataRetriever) ProjectID() (string, error) {
	if !f.inGCP {
		return "", fmt.Errorf("cannot retrieve project id: not running in GCP")
	}
	if f.project == "error" {
		return "", fmt.Errorf("error")
	}

	return f.project, nil
}

func TestMergeServiceDirectorySettings(t *testing.T) {
	cases := []struct {
		fromCli           *option.ServiceDirectory
		fromFile          *option.ServiceDirectory
		expRes            *option.ServiceDirectory
		metadataRetriever gcpMetadataRetriever
		expErr            error
	}{
		{
			fromCli: &option.ServiceDirectory{
				ProjectID:     "project",
				DefaultRegion: "region",
			},
			fromFile: &option.ServiceDirectory{
				ProjectID:     "file-project",
				DefaultRegion: "file-region",
			},
			expRes: &option.ServiceDirectory{
				ProjectID:     "project",
				DefaultRegion: "region",
			},
		},
		{
			fromCli: &option.ServiceDirectory{},
			fromFile: &option.ServiceDirectory{
				ProjectID:     "file-project",
				DefaultRegion: "file-region",
			},
			expRes: &option.ServiceDirectory{
				ProjectID:     "file-project",
				DefaultRegion: "file-region",
			},
		},
		{
			fromCli:           &option.ServiceDirectory{},
			metadataRetriever: &fakeGcpMetadataRetriever{inGCP: false},
			expErr:            fmt.Errorf("cannot retrieve project id: not running in GCP"),
		},
		{
			fromCli:           &option.ServiceDirectory{},
			metadataRetriever: &fakeGcpMetadataRetriever{project: "project", region: "error", inGCP: true},
			expErr:            errors.New("error"),
		},
		{
			fromCli:           &option.ServiceDirectory{},
			metadataRetriever: &fakeGcpMetadataRetriever{project: "error", region: "us-east1-c", inGCP: true},
			expErr:            errors.New("error"),
		},
		{
			fromCli:           &option.ServiceDirectory{},
			metadataRetriever: &fakeGcpMetadataRetriever{project: "project", region: "us-east1-c", inGCP: true},
			expRes:            &option.ServiceDirectory{ProjectID: "project", DefaultRegion: "us-east1"},
		},
	}

	a := assert.New(t)
	for _, currCase := range cases {
		res, err := validateAndMergeServiceDirectorySettings(currCase.fromCli, currCase.fromFile, currCase.metadataRetriever)

		a.Equal(currCase.expRes, res)
		if currCase.expErr == nil {
			a.NoError(err)
		} else {
			if currCase.expErr.Error() == "error" {
				a.Error(err)
			} else {
				a.Equal(currCase.expErr, err)
			}
		}
	}
}

func TestLoadGoogleServiceAccountSecret(t *testing.T) {
	cases := []struct {
		k8scli    kubernetes.Interface
		namespace string
		name      string
		expRes    []byte
		expErr    error
		caseName  string
	}{
		{
			k8scli: func() kubernetes.Interface {
				sec := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "ns",
					},
					Data: map[string][]byte{
						"whatever": []byte("whatever"),
					},
				}
				return fake.NewSimpleClientset(sec)
			}(),
			namespace: "different-ns",
			name:      "secret",
			expErr:    errors.New("any"),
		},
		{
			k8scli: func() kubernetes.Interface {
				sec := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "ns",
					},
				}
				return fake.NewSimpleClientset(sec)
			}(),
			namespace: "ns",
			name:      "secret",
			expErr:    errors.New("secret ns/secret has no data"),
		},
		{
			k8scli: func() kubernetes.Interface {
				sec := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "ns",
					},
					Data: map[string][]byte{
						"whatever-1": []byte("test-1"),
						"whatever-2": []byte("test-2"),
					},
				}
				return fake.NewSimpleClientset(sec)
			}(),
			namespace: "ns",
			name:      "secret",
			expRes:    []byte("test-1"),
			caseName:  "get-first-found",
		},
	}

	a := assert.New(t)
	for _, currCase := range cases {
		k8scli = currCase.k8scli
		res, err := loadGoogleServiceAccountSecret(currCase.namespace, currCase.name)

		if currCase.caseName != "get-first-found" {
			a.Equal(currCase.expRes, res)
		} else {
			a.NotNil(currCase.expRes)
			a.Contains([][]byte{[]byte("test-1"), []byte("test-2")}, res)
		}

		if currCase.expErr == nil {
			a.NoError(err)
		} else {
			if currCase.expErr.Error() == errors.New("any").Error() {
				a.Error(err)
			} else {
				a.Equal(currCase.expErr, err)
			}
		}
	}
}
