// Copyright 2021 Flant JSC
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

package infrastructure

import (
	"github.com/deckhouse/deckhouse/dhctl/pkg/config"
	"github.com/deckhouse/deckhouse/dhctl/pkg/state"
	"github.com/deckhouse/deckhouse/dhctl/pkg/terraform"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/tomb"
)

type BaseInfraTerraformController struct {
	metaConfig *config.MetaConfig
	stateCache state.Cache
}

func NewBaseInfraController(metaConfig *config.MetaConfig, stateCache state.Cache) *BaseInfraTerraformController {
	return &BaseInfraTerraformController{
		metaConfig: metaConfig,
		stateCache: stateCache,
	}
}

func (r *BaseInfraTerraformController) Destroy(clusterState []byte, sanityCheck bool) error {
	if err := saveInCacheIfNotExists(r.stateCache, "base-infrastructure.tfstate", clusterState); err != nil {
		return err
	}

	baseRunner := terraform.NewRunnerFromConfig(r.metaConfig, "base-infrastructure").
		WithVariables(r.metaConfig.MarshalConfig()).
		WithCache(r.stateCache).
		WithAllowedCachedState(true).
		WithAutoApprove(sanityCheck)
	tomb.RegisterOnShutdown("base-infrastructure", baseRunner.Stop)

	return terraform.DestroyPipeline(baseRunner, "Kubernetes cluster")
}
