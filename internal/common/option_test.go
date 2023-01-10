/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

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
	"fmt"
	"testing"

	"github.com/imdario/mergo"
	"gotest.tools/assert"
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

func TestLoadPonsConfigDefaults(t *testing.T) {
	Config = GetDefaultOps()
	// The default options define 1 PON per OLT
	// and 1 ONU per PON

	Services = []ServiceYaml{
		{
			Name: "test",
		},
	}

	ponsConf, err := getDefaultPonsConfig()

	assert.NilError(t, err, "Can't get defaults")

	assert.Equal(t, ponsConf.Number, uint32(1))
	assert.Equal(t, len(ponsConf.Ranges), 1)

	ranges := ponsConf.Ranges

	//The default should replicate the old way bbsim used to compute ranges
	assert.Equal(t, ranges[0].PonRange, IdRange{0, 0})
	assert.Equal(t, ranges[0].Technology, XGSPON.String())
	assert.Equal(t, ranges[0].OnuRange, IdRange{defaultOnuIdStart, defaultOnuIdStart})
	assert.Equal(t, ranges[0].AllocIdRange, IdRange{defaultAllocIdStart, defaultAllocIdStart + 4})
	assert.Equal(t, ranges[0].GemportRange, IdRange{defaultGemportIdStart, defaultGemportIdStart + 32})

	assert.NilError(t, validatePonsConfig(ponsConf), "Configuration is not valid")
}

func TestLoadPonsConfigFile(t *testing.T) {

	Config = GetDefaultOps()

	Services = []ServiceYaml{
		{
			Name: "test",
		},
	}

	ponsConf, err := getDefaultPonsConfig()

	assert.NilError(t, err, "Can't get defaults")

	yamlConf, err := loadBBSimPons("../../configs/pon-interfaces.yaml")

	assert.NilError(t, err, "Can't read config file")

	// merging Yaml and Default Values
	err = mergo.Merge(ponsConf, yamlConf, mergo.WithOverride)
	assert.NilError(t, err, "Can't merge YAML and Config")

	assert.Equal(t, ponsConf.Number, uint32(16))
	assert.Equal(t, len(ponsConf.Ranges), 1)

	ranges := ponsConf.Ranges

	assert.Equal(t, ranges[0].PonRange, IdRange{0, 15})
	assert.Equal(t, ranges[0].Technology, XGSPON.String())
	assert.Equal(t, ranges[0].OnuRange, IdRange{1, 255})
	assert.Equal(t, ranges[0].AllocIdRange, IdRange{256, 1024})
	assert.Equal(t, ranges[0].GemportRange, IdRange{256, 1024})

	assert.NilError(t, validatePonsConfig(ponsConf), "Configuration is not valid")
}

func getTestPonsConfiguration() *PonPortsConfig {
	return &PonPortsConfig{
		Number: 16,
		Ranges: []PonRangeConfig{
			{
				PonRange:     IdRange{0, 7},
				Technology:   GPON.String(),
				OnuRange:     IdRange{defaultOnuIdStart, defaultOnuIdStart},
				AllocIdRange: IdRange{defaultAllocIdStart, defaultAllocIdStart + 4},
				GemportRange: IdRange{defaultGemportIdStart, defaultGemportIdStart + 32},
			},
			{
				PonRange:     IdRange{8, 15},
				Technology:   XGSPON.String(),
				OnuRange:     IdRange{defaultOnuIdStart, defaultOnuIdStart},
				AllocIdRange: IdRange{defaultAllocIdStart, defaultAllocIdStart + 4},
				GemportRange: IdRange{defaultGemportIdStart, defaultGemportIdStart + 32},
			},
		},
	}
}

func TestPonsValidationTechnology(t *testing.T) {
	ponsConf := getTestPonsConfiguration()
	assert.NilError(t, validatePonsConfig(ponsConf), "Test configuration is not valid")

	ponsConf.Ranges[0].Technology = XGSPON.String()
	assert.NilError(t, validatePonsConfig(ponsConf), "Correct technology considered invalid")

	ponsConf.Ranges[0].Technology = GPON.String()
	assert.NilError(t, validatePonsConfig(ponsConf), "Correct technology considered invalid")

	ponsConf.Ranges[0].Technology = "TEST"
	assert.ErrorContains(t, validatePonsConfig(ponsConf), "technology", "Incorrect technology considered valid")
}

func TestPonsValidationPortsInRanges(t *testing.T) {
	ponsConf := getTestPonsConfiguration()
	assert.NilError(t, validatePonsConfig(ponsConf), "Test configuration is not valid")

	//The second range now misses pon 8
	ponsConf.Ranges[1].PonRange.StartId = 9

	assert.ErrorContains(t, validatePonsConfig(ponsConf), "not-defined", "Missing pon definition considered valid")

	//The second range defines pon 7 a second time
	ponsConf.Ranges[1].PonRange.StartId = 7

	assert.ErrorContains(t, validatePonsConfig(ponsConf), "duplicate", "Duplicate pon definition considered valid")

	//Get back to a known good configuration
	ponsConf = getTestPonsConfiguration()
	//The second range uses an Id that is out of range
	ponsConf.Ranges[1].PonRange.EndId = 16

	assert.ErrorContains(t, validatePonsConfig(ponsConf), "max", "Out of range pon definition considered valid")

	//Fix the start of the second range
	ponsConf.Ranges[1].PonRange.EndId = 15

	ponsConf.Number = 0

	assert.ErrorContains(t, validatePonsConfig(ponsConf), "no-pon-ports", "Zero pons considered valid")
}

func TestPonsValidationRangeLimits(t *testing.T) {
	ponsConf := getTestPonsConfiguration()
	assert.NilError(t, validatePonsConfig(ponsConf), "Test configuration is not valid")

	ponsConf.Ranges[0].PonRange = IdRange{0, 0}
	ponsConf.Ranges[1].PonRange = IdRange{1, 15}

	assert.NilError(t, validatePonsConfig(ponsConf), "Single pon range considered invalid")

	//Get back to a known good configuration
	ponsConf = getTestPonsConfiguration()
	ponsConf.Ranges[0].PonRange = IdRange{5, 4}

	assert.ErrorContains(t, validatePonsConfig(ponsConf), "limits", "Invalid pons range limits considered valid")

	ponsConf = getTestPonsConfiguration()
	ponsConf.Ranges[0].OnuRange = IdRange{5, 4}

	assert.ErrorContains(t, validatePonsConfig(ponsConf), "limits", "Invalid onus range limits considered valid")

	ponsConf = getTestPonsConfiguration()
	ponsConf.Ranges[0].AllocIdRange = IdRange{5, 4}

	assert.ErrorContains(t, validatePonsConfig(ponsConf), "limits", "Invalid alloc-ids range limits considered valid")

	ponsConf = getTestPonsConfiguration()
	ponsConf.Ranges[0].GemportRange = IdRange{5, 4}

	assert.ErrorContains(t, validatePonsConfig(ponsConf), "limits", "Invalid gemports range limits considered valid")
}

func TestGetPonConfigById(t *testing.T) {
	PonsConfig = getTestPonsConfiguration()

	for id := uint32(0); id < PonsConfig.Number-1; id++ {
		conf, err := GetPonConfigById(id)
		assert.NilError(t, err, fmt.Sprintf("Cannot get configuration for pon %d", id))

		if id > conf.PonRange.EndId || id < conf.PonRange.StartId {
			assert.NilError(t, err, fmt.Sprintf("Got wrong configuration for pon %d", id))
		}
	}

	_, err := GetPonConfigById(16)

	assert.Assert(t, err != nil, "Invalid pon id returned configuration")

	_, err = GetPonConfigById(100)

	assert.Assert(t, err != nil, "Invalid pon id returned configuration")
}
