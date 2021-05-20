/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"gotest.tools/assert"
	"testing"
)

func TestLoadBBSimServices(t *testing.T) {
	services, err := loadBBSimServices("../../configs/att-services.yaml")

	assert.NilError(t, err)

	assert.Equal(t, services[0].Name, "hsia")
	assert.Equal(t, services[0].CTag, 900)
	assert.Equal(t, services[0].STag, 900)
	assert.Equal(t, services[0].CTagAllocation, TagAllocationUnique.String())
	assert.Equal(t, services[0].STagAllocation, TagAllocationShared.String())
	assert.Equal(t, services[0].NeedsEapol, true)
	assert.Equal(t, services[0].NeedsDhcp, true)
	assert.Equal(t, services[0].NeedsIgmp, false)
	assert.Equal(t, services[0].TechnologyProfileID, 64)
}
