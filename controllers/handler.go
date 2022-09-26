/*
Copyright 2022 The KubeVela Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	v1alpha1 "github.com/oam-dev/vector-controller/api/v1alpha1"
)

const (
	uniqueConfigmapKey        = "vector.yaml"
	defaultConfigmapNamespace = "o11y-system"
	defaultConfigmapName      = "vector"
)

func createOrUpdateUniqueConfigmap(ctx context.Context, k8sClient client.Client, configObject v1alpha1.Config) error {
	configMap := v1.ConfigMap{}
	var err error
	if err = k8sClient.Get(ctx, types.NamespacedName{Namespace: configObject.Namespace, Name: configObject.Name}, &configMap); err != nil {
		if errors.IsNotFound(err) {
			if err = generateUniqueConfigmap(configObject, &configMap); err != nil {
				return err
			}
			if err = k8sClient.Create(ctx, &configMap); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	if err = generateUniqueConfigmap(configObject, &configMap); err != nil {
		return err
	}
	if err = k8sClient.Update(ctx, &configMap); err != nil {
		return err
	}
	return nil
}

func generateUniqueConfigmap(vectorConfig v1alpha1.Config, configmap *v1.ConfigMap) error {
	configStr, err := yaml.Marshal(vectorConfig.Spec.VectorConfig)
	if err != nil {
		return err
	}
	configmap.Data = map[string]string{uniqueConfigmapKey: string(configStr)}
	configmap.Name = vectorConfig.Name
	configmap.Namespace = vectorConfig.Namespace
	return nil
}

func createOrMergeConfigmap(ctx context.Context, k8sClient client.Client, configObject *v1alpha1.Config) error {
	cm, err := fetchTargetConfigmap(ctx, k8sClient, configObject)
	if errors.IsNotFound(err) {
		targetCMNamespace, targetCMName := targetConfigMapNamespaceAndName(configObject)
		cm := v1.ConfigMap{ObjectMeta: v12.ObjectMeta{Namespace: targetCMNamespace, Name: targetCMName}}
		if err = generateMergeConfigMap(configObject, &cm); err != nil {
			return err
		}
		if err = k8sClient.Create(ctx, &cm); err != nil {
			return err
		}
		return nil
	}
	if err = generateMergeConfigMap(configObject, cm); err != nil {
		return err
	}
	if err = k8sClient.Update(ctx, cm); err != nil {
		return err
	}
	return nil
}

func generateMergeConfigMap(vectorConfig *v1alpha1.Config, configmap *v1.ConfigMap) error {
	configStr, err := yaml.Marshal(vectorConfig.Spec.VectorConfig)
	if err != nil {
		return err
	}
	if configmap.Data == nil {
		configmap.Data = map[string]string{fmt.Sprintf("%s-%s-vector.yaml", vectorConfig.GetName(), vectorConfig.GetNamespace()): string(configStr)}
	} else {
		configmap.Data[fmt.Sprintf("%s-%s-vector.yaml", vectorConfig.GetName(), vectorConfig.GetNamespace())] = string(configStr)
	}
	return nil
}

func fetchTargetConfigmap(ctx context.Context, k8sClient client.Client, configObject *v1alpha1.Config) (*v1.ConfigMap, error) {
	configMap := v1.ConfigMap{}
	var err error

	targetCMNamespace, targetCMName := targetConfigMapNamespaceAndName(configObject)

	if err = k8sClient.Get(ctx, types.NamespacedName{Namespace: targetCMNamespace, Name: targetCMName}, &configMap); err != nil {
		return nil, err
	}
	return &configMap, nil
}

func targetConfigMapNamespaceAndName(config *v1alpha1.Config) (string, string) {
	var targetCMNamespace, targetCMName string
	if len(config.Spec.TargetConfigMap.Namespace) != 0 {
		targetCMNamespace = config.Spec.TargetConfigMap.Namespace
	} else {
		targetCMNamespace = defaultConfigmapNamespace
	}

	if len(config.Spec.TargetConfigMap.Name) != 0 {
		targetCMName = config.Spec.TargetConfigMap.Name
	} else {
		targetCMName = defaultConfigmapName
	}
	return targetCMNamespace, targetCMName
}

func deleteConfigFromConfigmap(ctx context.Context, k8sClient client.Client, configObject *v1alpha1.Config) error {
	cm, err := fetchTargetConfigmap(ctx, k8sClient, configObject)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	delete(cm.Data, fmt.Sprintf("%s-%s-vector.yaml", configObject.GetName(), configObject.GetNamespace()))
	if err = k8sClient.Update(ctx, cm); err != nil {
		return err
	}
	return nil
}
