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

package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/yaml"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// log is for logging in this package.
var configlog = logf.Log.WithName("config-resource")

func (r *Config) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-vector-oam-dev-v1alpha1-config,mutating=false,failurePolicy=fail,sideEffects=None,groups=vector.oam.dev,resources=configs,verbs=create;update,versions=v1alpha1,name=vconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Config{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Config) ValidateCreate() error {
	configlog.Info("validate create", "name", r.Name)
	return validate(*r)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Config) ValidateUpdate(old runtime.Object) error {
	configlog.Info("validate update", "name", r.Name)
	return validate(*r)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Config) ValidateDelete() error {
	configlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func validate(config Config) error {
	if config.Spec.Role == SidecarRoleType {
		configStr, err := yaml.Marshal(config.Spec.VectorConfig)
		if err != nil {
			return err
		}
		stdout, err := runValidate(strings.Join([]string{config.Name, config.Namespace}, "-"), configStr)
		if err != nil {
			klog.Error(err.Error())
			return fmt.Errorf("vectorConfig error: %s", string(stdout))
		}
	}
	return nil
}

func runValidate(namePrefix string, config []byte) ([]byte, error) {
	suffix := randStringRunes(8)
	fileName := strings.Join([]string{namePrefix, suffix}, "-")
	filePath := "/tmp/" + fileName + ".yaml"
	err := ioutil.WriteFile(filePath, config, 0666)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := os.Remove(filePath)
		if err != nil {
			klog.Errorf("remove tmp file meet error", err)
		}
	}()
	cmd := exec.Command("/vector", "validate", "--no-environment", filePath)
	return cmd.Output()
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
