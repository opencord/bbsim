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

package devices

import (
	"testing"

	"gotest.tools/assert"

	"github.com/opencord/bbsim/internal/common"
)

func Test_KafkaEventTopic(t *testing.T) {
	common.Options.BBSim.KafkaEventTopic = "test"

	numPon := 1
	numOnu := 1

	olt := createMockOlt(numPon, numOnu)

	err := common.InitializePublisher(olt.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, common.Topic, "test")
}
