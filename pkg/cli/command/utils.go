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
	"io/ioutil"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	k8scli kubernetes.Interface
	inK8s  *bool
)

func runningInK8s() bool {
	if inK8s != nil {
		return *inK8s
	}

	_, err := rest.InClusterConfig()
	isRunning := err == nil
	inK8s = &isRunning

	return *inK8s
}

func loadExistingSettings(filepath, cfgMap, defaultCfgMap, namespace string) ([]byte, error) {
	if filepath != "" {
		log.Debug().Str("path", filepath).Msg("loading settings from file...")
		return ioutil.ReadFile(filepath)
	}

	if cfgMap != "" {
		return loadSettingsFromConfigMap(namespace, cfgMap)
	}

	if runningInK8s() {
		// This step is performed because we are running inside
		// Kubernetes and the user has not provided any configmap
		// name, so we just try to see if it is there or not.
		// Since the user did not provide the name explictly,
		// we don't consider this an error (the user might have
		// provided the settings via CLI in the deployment).
		if settings, err := loadSettingsFromConfigMap(namespace, defaultCfgMap); err == nil {
			log.Debug().
				Str("configmap-name", defaultCfgMap).
				Str("namespace", namespace).
				Msg("retrieved existing configmap")
			return settings, nil
		}
	}

	return nil, nil
}

func getKubernetesClientSet() (kubernetes.Interface, error) {
	if k8scli != nil {
		return k8scli, nil
	}

	k8sconf, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}

	_k8scli, err := kubernetes.NewForConfig(k8sconf)
	if err != nil {
		return nil, err
	}

	k8scli = _k8scli
	return k8scli, nil
}

func loadSettingsFromConfigMap(namespace, name string) ([]byte, error) {
	cli, err := getKubernetesClientSet()
	if err != nil {
		return nil, fmt.Errorf(`error while getting Kubernetes clientset: %s`, err)
	}

	ctx, canc := context.WithTimeout(context.Background(), time.Minute)
	defer canc()

	confMap, err := cli.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf(`error while getting ConfigMap "%s.%s": %s`, name, namespace, err)
	}

	// Get the first data in it.
	// This is for configmaps that are deployed as --from-file.
	var data []byte
	for _, d := range confMap.Data {
		data = []byte(d)
		break
	}

	log.Debug().
		Str("configmap-name", name).
		Str("namespace", namespace).
		Msg("configmap found")

	return data, nil
}

func getGCPServiceAccount(fromPath, fromSecret, namespace string) ([]byte, error) {
	if fromPath != "" {
		return ioutil.ReadFile(fromPath)
	}

	if fromSecret != "" {
		return loadGoogleServiceAccountSecret(namespace, fromSecret)
	}

	if runningInK8s() {
		// If we are running inside Kubernetes but user did not provide a
		// Secret name, we're going to do one last attempt with the default
		// Secret name.
		// We do not throw an error here because there was no explicit
		// intention from the user to get the Secret, as he/she did not use
		// the CLI flag to do so: we do this as a sort of "last resort" or
		// "just in case".
		if sa, err := loadGoogleServiceAccountSecret(namespace, defaultGCPServiceAccountSecretName); err == nil {
			return sa, nil
		}
	}

	// We could not find a service account anywhere! We'll just try to
	// authenticate with API scopes, I guess... 🤷‍♂️
	return nil, nil
}
