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

package containeranalysis

import (
	"context"
	"fmt"

	ca "cloud.google.com/go/containeranalysis/apiv1beta1"
	"github.com/tektoncd/chains/pkg/signing/formats"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	versioned "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/devtools/containeranalysis/v1beta1/attestation"
	"google.golang.org/genproto/googleapis/devtools/containeranalysis/v1beta1/common"
	grafeaspb "google.golang.org/genproto/googleapis/devtools/containeranalysis/v1beta1/grafeas"
)

const (
	StorageBackendContainerAnalysis = "containeranalysis"
)

// Backend is a storage backend
type Backend struct {
	pipelienclientset versioned.Interface
	logger            *zap.SugaredLogger
	tr                *v1beta1.TaskRun
	client            *ca.GrafeasV1Beta1Client
}

// NewStorageBackend returns a new Tekton StorageBackend that stores signatures on a TaskRun
func NewStorageBackend(ps versioned.Interface, logger *zap.SugaredLogger, tr *v1beta1.TaskRun) (*Backend, error) {
	ctx := context.Background()
	client, err := ca.NewGrafeasV1Beta1Client(ctx)
	if err != nil {
		return nil, err
	}
	return &Backend{
		client:            client,
		pipelienclientset: ps,
		logger:            logger,
		tr:                tr,
	}, nil
}

// StorePayload implements the Payloader interface
func (b *Backend) StorePayload(signed []byte, signature string, payloadType formats.PayloadType) error {
	b.logger.Infof("Storing payload type %s on TaskRun %s/%s", payloadType, b.tr.Namespace, b.tr.Name)

	req := &grafeaspb.CreateOccurrenceRequest{
		Parent: "projects/dlorenc-vmtest2",
		Occurrence: &grafeaspb.Occurrence{
			Resource: &grafeaspb.Resource{
				Name: string(b.tr.UID),
				Uri:  "tekton://chains.tekton.dev/taskruns/" + string(b.tr.UID),
			},
			NoteName: "projects/dlorenc-vmtest2/notes/tekton-chains",
			Kind:     common.NoteKind_ATTESTATION,
			Details: &grafeaspb.Occurrence_Attestation{
				Attestation: &attestation.Details{
					Attestation: &attestation.Attestation{
						Signature: &attestation.Attestation_GenericSignedAttestation{
							GenericSignedAttestation: &attestation.GenericSignedAttestation{
								SerializedPayload: signed,
								Signatures: []*common.Signature{
									{
										Signature: []byte(signature),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	occ, err := b.client.CreateOccurrence(context.Background(), req)
	fmt.Println(occ)
	return err
}

func (b *Backend) Type() string {
	return StorageBackendContainerAnalysis
}
