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

type mockPasswordReader struct {
	whatToReturn string
}

func (m *mockPasswordReader) readPassword() (string, error) {
	if m.whatToReturn == "error" {
		return "", fmt.Errorf("any")
	}

	return m.whatToReturn, nil
}

func TestMergeEtcdSettings(t *testing.T) {
	cases := []struct {
		fromCli  *option.Etcd
		fromFile *option.Etcd
		expRes   *option.Etcd
		expErr   error
	}{
		{
			fromCli: &option.Etcd{
				Prefix: "//test-prefix/another///",
				Endpoints: []string{
					"test:2379",
					"test:2379",
					"test-1:2379",
				},
			},
			expRes: &option.Etcd{
				Prefix: "/test-prefix/another/",
				Endpoints: []string{
					"test:2379",
					"test-1:2379",
				},
			},
		},
		{
			fromCli: &option.Etcd{},
			fromFile: &option.Etcd{
				Prefix: "test",
				Endpoints: []string{
					"test:2379",
					"test:2379",
					"test-1:2379",
				},
			},
			expRes: &option.Etcd{
				Prefix: "/test/",
				Endpoints: []string{
					"test:2379",
					"test-1:2379",
				},
			},
		},
		{
			fromCli: &option.Etcd{
				Prefix: "/",
			},
			expRes: &option.Etcd{
				Prefix: "/",
				Endpoints: []string{
					"localhost:2379",
				},
			},
		},
		{
			fromCli: &option.Etcd{},
			expRes: &option.Etcd{
				Prefix: fmt.Sprintf("/%s/", option.DefaultEtcdPrefix),
				Endpoints: []string{
					"localhost:2379",
				},
			},
		},
	}

	a := assert.New(t)
	for _, currCase := range cases {
		s := mergeEtcdSettings(currCase.fromCli, currCase.fromFile)

		a.Equal(currCase.expRes.Prefix, s.Prefix)
		a.ElementsMatch(currCase.expRes.Endpoints, s.Endpoints)
	}
}

func TestLoadEtcdUsernameAndPasswordFromSecret(t *testing.T) {
	cases := []struct {
		k8scli    kubernetes.Interface
		namespace string
		name      string
		expUser   string
		expPass   string
		expErr    error
	}{
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
			namespace: "different",
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
			expErr:    fmt.Errorf("secret ns/secret doesn't have a username"),
		},
		{
			k8scli: func() kubernetes.Interface {
				sec := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "ns",
					},
					Data: map[string][]byte{
						"username": []byte("test"),
						"password": []byte("test-1"),
					},
				}
				return fake.NewSimpleClientset(sec)
			}(),
			namespace: "ns",
			name:      "secret",
			expUser:   "test",
			expPass:   "test-1",
		},
	}

	a := assert.New(t)
	for _, currCase := range cases {
		k8scli = currCase.k8scli
		u, p, err := loadEtcdUsernameAndPasswordFromSecret(currCase.namespace, currCase.name)

		a.Equal(currCase.expUser, u)
		a.Equal(currCase.expPass, p)
		if currCase.expErr == nil {
			a.Nil(err)
		} else {
			if currCase.expErr.Error() == errors.New("any").Error() {
				a.Error(err)
			} else {
				a.Equal(currCase.expErr, err)
			}
		}
	}
}

func TestGetEtcdUsernameAndPassword(t *testing.T) {
	cases := []struct {
		before     func()
		after      func()
		inputPass  bool
		username   string
		credSecret string
		namespace  string
		passReader passwordReader
		expUser    string
		expPass    string
		expErr     error
	}{
		{
			expUser: "",
			expPass: "",
		},
		{
			before: func() {
				ik := true
				inK8s = &ik
			},
			after: func() {
				ik := false
				inK8s = &ik
			},
			inputPass: true,
			expErr:    fmt.Errorf("please use Kubernetes Secrets to define credentials"),
		},
		{
			inputPass:  true,
			expUser:    "root",
			passReader: &mockPasswordReader{whatToReturn: "pwd"},
			expPass:    "pwd",
			credSecret: "should-not-be-considered",
		},
		{
			inputPass:  true,
			passReader: &mockPasswordReader{whatToReturn: "error"},
			expErr:     errors.New("any"),
		},
		{
			username: "non-password-user",
			expUser:  "non-password-user",
		},
		{
			credSecret: "secret",
			namespace:  "ns",
			before: func() {
				sec := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "ns",
					},
					Data: map[string][]byte{
						"username": []byte("test"),
						"password": []byte("test-1"),
					},
				}
				k8scli = fake.NewSimpleClientset(sec)
			},
			expUser: "test",
			expPass: "test-1",
		},
		{
			namespace: "ns",
			before: func() {
				in := true
				inK8s = &in
				sec := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      option.DefaultEtcdCredentialsSecretName,
						Namespace: "ns",
					},
					Data: map[string][]byte{
						"username": []byte("user-from-k8s"),
						"password": []byte("pass-from-k8s"),
					},
				}
				k8scli = fake.NewSimpleClientset(sec)
			},
			expUser: "user-from-k8s",
			expPass: "pass-from-k8s",
		},
	}

	a := assert.New(t)
	for i, currCase := range cases {
		if currCase.before != nil {
			currCase.before()
		}
		u, p, err := getEtcdUsernameAndPassword(currCase.inputPass, currCase.username, currCase.credSecret, currCase.namespace, currCase.passReader)

		userRes, passRes := a.Equal(currCase.expUser, u), a.Equal(currCase.expPass, p)
		var errRes bool

		if currCase.expErr == nil {
			errRes = a.NoError(err)
		} else {
			if currCase.expErr.Error() == errors.New("any").Error() {
				errRes = a.Error(err)
			} else {
				errRes = a.Equal(currCase.expErr, err)
			}
		}
		if currCase.after != nil {
			currCase.after()
		}

		if !userRes || !passRes || !errRes {
			a.FailNow("failed", i)
		}
	}
}
