/*
Copyright 2020 The Tekton Authors
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

package artifacts

import (
	"github.com/tektoncd/chains/pkg/config"
	"github.com/tektoncd/chains/pkg/signing/formats"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"go.uber.org/zap"
)

type Signable interface {
	// I don't like this, it should instead take a single thing to payload. Maybe an interface{}?
	// And an "extractFromTr" that returns a list if interfaces?
	GeneratePayloads(tr *v1beta1.TaskRun, enabledFormats map[string]struct{}) map[formats.PayloadType]interface{}
	EnabledFormats(cfg config.Config) map[string]struct{}
	EnabledStorageBackends(cfg config.Config) map[string]struct{}
}

type TaskRunArtifact struct {
	Logger *zap.SugaredLogger
}

func (ta *TaskRunArtifact) GeneratePayloads(tr *v1beta1.TaskRun, enabledFormats map[string]struct{}) map[formats.PayloadType]interface{} {
	payloads := map[formats.PayloadType]interface{}{}
	for _, payloader := range formats.AllPayloadTypes {
		if _, ok := enabledFormats[string(payloader.Type())]; !ok {
			ta.Logger.Debugf("skipping format %s for taskrun %s/%s", payloader, tr.Namespace, tr.Name)
		}
		payload := payloader.CreatePayload(tr)
		payloads[payloader.Type()] = payload
	}
	return payloads
}

func (ta *TaskRunArtifact) EnabledFormats(cfg config.Config) map[string]struct{} {
	return cfg.Artifacts.TaskRuns.Formats.EnabledFormats
}

func (ta *TaskRunArtifact) EnabledStorageBackends(cfg config.Config) map[string]struct{} {
	return cfg.Artifacts.TaskRuns.StorageBackends.EnabledBackends
}

type OCIImageArtifact struct {
	Logger *zap.SugaredLogger
}

type image struct {
	url    string
	digest string
}

func (oa *OCIImageArtifact) GeneratePayloads(tr *v1beta1.TaskRun, enabledFormats map[string]struct{}) map[formats.PayloadType]interface{} {
	// Look to see if we have any container images

	// Every image has a digest and a URL in the ResourcesResult list
	images := map[string]image{}
	for _, output := range tr.Status.TaskSpec.Resources.Outputs {
		if output.Type != v1beta1.PipelineResourceTypeImage {
			continue
		}

		image := image{}
		for _, rr := range tr.Status.ResourcesResult {
			if rr.ResourceRef.Name == output.Name {
				if rr.Key == "url" {
					image.url = rr.Value
				} else if rr.Key == "digest" {
					image.digest = rr.Value
				}
			}
		}
		images[output.Name] = image
	}

	payloads := map[formats.PayloadType]interface{}{}
	for _, payloader := range formats.AllPayloadTypes {
		imagePayloads := []interface{}{}
		for _, image := range images {
			if _, ok := enabledFormats[string(payloader.Type())]; !ok {
				oa.Logger.Debugf("skipping format %s for taskrun %s/%s", payloader, tr.Namespace, tr.Name)
			}
			imagePayloads = append(imagePayloads, payloader.CreatePayload(image))
		}
		payloads[payloader.Type()] = imagePayloads
	}
	return payloads
}

func (oa *OCIImageArtifact) EnabledFormats(cfg config.Config) map[string]struct{} {
	return cfg.Artifacts.TaskRuns.Formats.EnabledFormats
}

func (oa *OCIImageArtifact) EnabledStorageBackends(cfg config.Config) map[string]struct{} {
	return cfg.Artifacts.TaskRuns.StorageBackends.EnabledBackends
}
