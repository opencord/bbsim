/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

 * Licensed under the Apache License, Version 2.0 (the License);
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an AS IS BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sadis

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/opencord/bbsim/internal/common"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opencord/bbsim/internal/bbsim/devices"
	"gotest.tools/assert"
)

func createMockDevices() (*devices.OltDevice, *devices.Onu) {

	// create a ONU
	onu := &devices.Onu{
		ID:        1,
		PonPortID: 1,
	}
	onu.SerialNumber = devices.NewSN(0, onu.PonPortID, onu.ID)

	// create 2 UNIs for the ONU
	unis := []devices.UniPortIf{
		&devices.UniPort{ID: 0, Onu: onu, MeId: omcilib.GenerateUniPortEntityId(1)},
		&devices.UniPort{ID: 1, Onu: onu, MeId: omcilib.GenerateUniPortEntityId(2)},
	}
	onu.UniPorts = unis

	// create a service on each UNI
	c_tag := 900
	for i, u := range onu.UniPorts {
		uni := u.(*devices.UniPort)
		mac := net.HardwareAddr{0x2e, 0x01, byte(1), byte(1), byte(0), byte(i)}
		uni.Services = []devices.ServiceIf{
			&devices.Service{Name: "hsia", CTag: c_tag + i, STag: 900, NeedsEapol: true, NeedsDhcp: true, NeedsIgmp: true, HwAddress: mac, TechnologyProfileID: 64},
		}
	}

	olt := &devices.OltDevice{
		ID: 0,
		Pons: []*devices.PonPort{
			{
				ID:     0,
				NumOnu: 1,
				Onus:   []*devices.Onu{onu},
			},
		},
	}

	return olt, onu
}

func TestSadisServer_GetOnuEntryV2(t *testing.T) {

	olt, onu := createMockDevices()

	for _, u := range onu.UniPorts {
		uni := u.(*devices.UniPort)

		entry, err := GetOnuEntryV2(olt, onu, fmt.Sprintf("%d", uni.ID+1))

		assert.NilError(t, err)

		assert.Equal(t, fmt.Sprintf("%s-%d", onu.Sn(), uni.ID+1), entry.ID)

		// we only have one service, thus get a single entry in the UniTagList
		assert.Equal(t, len(entry.UniTagList), 1)
		assert.Equal(t, entry.UniTagList[0].PonCTag, int(900+uni.ID))
		assert.Equal(t, entry.UniTagList[0].PonSTag, 900)
		assert.Equal(t, entry.UniTagList[0].DownstreamBandwidthProfile, "User_Bandwidth2")
		assert.Equal(t, entry.UniTagList[0].UpstreamBandwidthProfile, "User_Bandwidth1")
		assert.Equal(t, entry.UniTagList[0].TechnologyProfileID, 64)
		assert.Equal(t, entry.UniTagList[0].IsDhcpRequired, true)
		assert.Equal(t, entry.UniTagList[0].IsIgmpRequired, true)
	}
}

func TestSadisServer_ServeStaticConfig(t *testing.T) {
	olt, onu := createMockDevices()
	common.Config = &common.GlobalConfig{
		Olt: common.OltConfig{
			ID:          olt.ID,
			PonPorts:    1,
			OnusPonPort: 1,
			DeviceId:    net.HardwareAddr{0xA, 0xA, 0xA, 0xA, 0xA, byte(olt.ID)}.String(),
		},
	}

	s := &SadisServer{
		Olt: olt,
	}

	// Need to create a router that we can pass the request through so that the vars will be added to the context
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc(StaticConfigUrl, s.ServeStaticConfig)

	// check that v2 returns the expected result
	req, err := http.NewRequest("GET", "/v2/static", nil)
	if err != nil {
		t.Fatal(err)
	}
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	cfg := SadisConfig{}
	if err := json.Unmarshal(rr.Body.Bytes(), &cfg); err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, len(cfg.Sadis.Entries), 3) // 2 UNI and 1 OLT

	// OLT
	oltEntry := cfg.Sadis.Entries[0].(map[string]interface{})
	assert.Equal(t, oltEntry["hardwareIdentifier"], common.Config.Olt.DeviceId)

	// UNIs
	for i, u := range onu.UniPorts {
		uni := u.(*devices.UniPort)
		uniEntry := cfg.Sadis.Entries[i+1].(map[string]interface{})
		assert.Equal(t, uniEntry["id"], fmt.Sprintf("%s-%d", onu.Sn(), uni.ID+1))
	}
}

func TestSadisServer_GetOnuEntryV2_multi_service(t *testing.T) {

	mac := net.HardwareAddr{0x2e, byte(1), byte(1), byte(1), byte(1), byte(2)}

	hsia := devices.Service{Name: "hsia", HwAddress: net.HardwareAddr{0x2e, byte(1), byte(1), byte(1), byte(1), byte(1)},
		CTag: 900, STag: 900, TechnologyProfileID: 64}

	voip := devices.Service{Name: "voip", HwAddress: mac,
		CTag: 901, STag: 900, TechnologyProfileID: 65}

	vod := devices.Service{Name: "vod", HwAddress: net.HardwareAddr{0x2e, byte(1), byte(1), byte(1), byte(1), byte(3)},
		CTag: 902, STag: 900, TechnologyProfileID: 66}

	olt, onu := createMockDevices()

	onu.UniPorts[0].(*devices.UniPort).Services = []devices.ServiceIf{&hsia, &voip, &vod}

	uni := "1"

	entry, err := GetOnuEntryV2(olt, onu, uni)

	assert.NilError(t, err)

	assert.Equal(t, entry.ID, fmt.Sprintf("%s-%s", onu.Sn(), uni))

	assert.Equal(t, len(entry.UniTagList), 3)

	assert.Equal(t, entry.UniTagList[0].PonCTag, 900)
	assert.Equal(t, entry.UniTagList[0].PonSTag, 900)
	assert.Equal(t, entry.UniTagList[0].TechnologyProfileID, 64)

	assert.Equal(t, entry.UniTagList[1].PonCTag, 901)
	assert.Equal(t, entry.UniTagList[1].PonSTag, 900)
	assert.Equal(t, entry.UniTagList[1].TechnologyProfileID, 65)

	assert.Equal(t, entry.UniTagList[2].PonCTag, 902)
	assert.Equal(t, entry.UniTagList[2].PonSTag, 900)
	assert.Equal(t, entry.UniTagList[2].TechnologyProfileID, 66)
}
