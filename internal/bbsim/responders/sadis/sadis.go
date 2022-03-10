/*
 * Copyright 2018-present Open Networking Foundation

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
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
)

var sadisLogger = log.WithFields(log.Fields{
	"module": "SADIS",
})

const (
	BaseConfigUrl   = "/{version}/cfg"
	StaticConfigUrl = "/{version}/static"
	SadisEntryUrl   = "/{version}/subscribers/{ID}"
	SadisBwUrl      = "/{version}/bandwidthprofiles/{ID}"
)

type SadisServer struct {
	Olt *devices.OltDevice
}

// bandwidthProfiles contains some dummy profiles
var bandwidthProfiles = map[string][]*SadisBWPEntry{
	common.BP_FORMAT_MEF: {
		{ID: "User_Bandwidth1", AIR: 100000, CBS: 10000, CIR: 30000, EBS: 1000, EIR: 100000},
		{ID: "User_Bandwidth2", AIR: 100000, CBS: 5000, CIR: 100000, EBS: 5000, EIR: 100000},
		{ID: "User_Bandwidth3", AIR: 100000, CBS: 5000, CIR: 1000000, EBS: 5000, EIR: 1000000},
		{ID: "Default", AIR: 100000, CBS: 30, CIR: 600, EBS: 30, EIR: 400},
	},
	common.BP_FORMAT_IETF: {
		{ID: "User_Bandwidth1", CBS: 10000, CIR: 30000, GIR: 100000, PIR: 20000, PBS: 1000},
		{ID: "User_Bandwidth2", CBS: 5000, CIR: 100000, GIR: 100000, PIR: 30000, PBS: 5000},
		{ID: "User_Bandwidth3", CBS: 5000, CIR: 1000000, GIR: 100000, PIR: 40000, PBS: 5000},
		{ID: "Default", CBS: 30, CIR: 600, GIR: 0, PIR: 32000, PBS: 30},
	},
}

// SadisConfig is the top-level SADIS configuration struct
type SadisConfig struct {
	Sadis            SadisEntries            `json:"sadis"`
	BandwidthProfile BandwidthProfileEntries `json:"bandwidthprofile"`
}

type SadisEntries struct {
	Integration SadisIntegration `json:"integration"`
	Entries     []interface{}    `json:"entries,omitempty"`
}
type BandwidthProfileEntries struct {
	Integration SadisIntegration `json:"integration"`
	Entries     []*SadisBWPEntry `json:"entries,omitempty"`
}

type SadisIntegration struct {
	URL   string `json:"url,omitempty"`
	Cache struct {
		Enabled bool   `json:"enabled"`
		MaxSize int    `json:"maxsize"`
		TTL     string `json:"ttl"`
	} `json:"cache"`
}

type SadisOltEntry struct {
	ID                 string `json:"id"`
	HardwareIdentifier string `json:"hardwareIdentifier"`
	IPAddress          string `json:"ipAddress"`
	NasID              string `json:"nasId"`
	UplinkPort         int    `json:"uplinkPort"`
	NniDhcpTrapVid     int    `json:"nniDhcpTrapVid,omitempty"`
}

type SadisOnuEntryV2 struct {
	ID         string        `json:"id"`
	NasPortID  string        `json:"nasPortId"`
	CircuitID  string        `json:"circuitId"`
	RemoteID   string        `json:"remoteId"`
	UniTagList []SadisUniTag `json:"uniTagList"` // this can be SadisUniTagAtt, SadisUniTagDt
}

type SadisUniTag struct {
	UniTagMatch                int    `json:"uniTagMatch,omitempty"`
	PonCTag                    int    `json:"ponCTag,omitempty"`
	PonSTag                    int    `json:"ponSTag,omitempty"`
	TechnologyProfileID        int    `json:"technologyProfileId,omitempty"`
	UpstreamBandwidthProfile   string `json:"upstreamBandwidthProfile,omitempty"`
	DownstreamBandwidthProfile string `json:"downstreamBandwidthProfile,omitempty"`
	IsDhcpRequired             bool   `json:"isDhcpRequired,omitempty"`
	IsIgmpRequired             bool   `json:"isIgmpRequired,omitempty"`
	IsPPPoERequired            bool   `json:"isPppoeRequired,omitempty"`
	ConfiguredMacAddress       string `json:"configuredMacAddress,omitempty"`
	EnableMacLearning          bool   `json:"enableMacLearning,omitempty"`
	UsPonCTagPriority          uint8  `json:"usPonCTagPriority,omitempty"`
	UsPonSTagPriority          uint8  `json:"usPonSTagPriority,omitempty"`
	DsPonCTagPriority          uint8  `json:"dsPonCTagPriority,omitempty"`
	DsPonSTagPriority          uint8  `json:"dsPonSTagPriority,omitempty"`
	ServiceName                string `json:"serviceName,omitempty"`
}

// SADIS BandwithProfile Entry
type SadisBWPEntry struct {
	// common attributes
	ID  string `json:"id"`
	CBS int    `json:"cbs"`
	CIR int    `json:"cir"`
	// MEF attributes
	AIR int `json:"air,omitempty"`
	EBS int `json:"ebs,omitempty"`
	EIR int `json:"eir,omitempty"`
	// IETF attributes
	GIR int `json:"gir,omitempty"`
	PIR int `json:"pir,omitempty"`
	PBS int `json:"pbs,omitempty"`
}

// GetSadisConfig returns a full SADIS configuration struct ready to be marshalled into JSON
func GetSadisConfig(olt *devices.OltDevice, version string) *SadisConfig {
	sadisEntries, _ := GetSadisEntries(olt, version)
	bwpEntries := getBWPEntries(version)

	conf := &SadisConfig{}
	conf.Sadis = *sadisEntries
	conf.BandwidthProfile = *bwpEntries

	return conf
}

func GetSadisEntries(olt *devices.OltDevice, version string) (*SadisEntries, error) {
	solt, _ := GetOltEntry(olt)

	entries := []interface{}{}
	entries = append(entries, solt)

	a := strings.Split(common.Config.BBSim.SadisRestAddress, ":")
	port := a[len(a)-1]

	integration := SadisIntegration{}
	integration.URL = "http://bbsim:" + port + "/" + version + "/subscribers/%s"
	integration.Cache.Enabled = false
	integration.Cache.MaxSize = 50
	integration.Cache.TTL = "PT0m"

	sadis := &SadisEntries{
		integration,
		entries,
	}

	return sadis, nil
}

func GetOltEntry(olt *devices.OltDevice) (*SadisOltEntry, error) {
	ip, _ := common.GetIPAddr("nni") // TODO verify which IP to report
	solt := &SadisOltEntry{
		ID:                 olt.SerialNumber,
		HardwareIdentifier: common.Config.Olt.DeviceId,
		IPAddress:          ip,
		NasID:              olt.SerialNumber,
		UplinkPort:         16777216, // TODO currently assumes we only have one NNI port
		NniDhcpTrapVid:     olt.NniDhcpTrapVid,
	}
	return solt, nil
}

func GetOnuEntryV2(olt *devices.OltDevice, onu *devices.Onu, uniStr string) (*SadisOnuEntryV2, error) {
	uniSuffix := "-" + uniStr

	sonuv2 := &SadisOnuEntryV2{
		ID: onu.Sn() + uniSuffix,
	}

	uniId, err := strconv.ParseUint(uniStr, 10, 32)
	if err != nil {
		return nil, err
	}

	// find the correct UNI
	// NOTE that in SADIS uni.Id 0 corresponds to BBSM00000101-1
	uni, err := onu.FindUniById(uint32(uniId - 1))
	if err != nil {
		return nil, err
	}

	// createUniTagList
	for _, s := range uni.Services {

		service := s.(*devices.Service)

		tag := SadisUniTag{
			ServiceName:                service.Name,
			IsIgmpRequired:             service.NeedsIgmp,
			IsDhcpRequired:             service.NeedsDhcp,
			IsPPPoERequired:            service.NeedsPPPoE,
			TechnologyProfileID:        service.TechnologyProfileID,
			UpstreamBandwidthProfile:   "User_Bandwidth1",
			DownstreamBandwidthProfile: "User_Bandwidth2",
			EnableMacLearning:          service.EnableMacLearning,
			PonCTag:                    service.CTag,
			PonSTag:                    service.STag,
		}

		if service.UniTagMatch != 0 {
			tag.UniTagMatch = service.UniTagMatch
		}

		if service.ConfigureMacAddress {
			tag.ConfiguredMacAddress = service.HwAddress.String()
		}

		if service.UsPonCTagPriority != 0 {
			tag.UsPonCTagPriority = service.UsPonCTagPriority
		}

		if service.UsPonSTagPriority != 0 {
			tag.UsPonSTagPriority = service.UsPonSTagPriority
		}

		if service.DsPonCTagPriority != 0 {
			tag.DsPonCTagPriority = service.DsPonCTagPriority
		}

		if service.DsPonSTagPriority != 0 {
			tag.DsPonSTagPriority = service.DsPonSTagPriority
		}

		sonuv2.UniTagList = append(sonuv2.UniTagList, tag)
	}

	return sonuv2, nil
}

func getBWPEntries(version string) *BandwidthProfileEntries {
	a := strings.Split(common.Config.BBSim.SadisRestAddress, ":")
	port := a[len(a)-1]

	integration := SadisIntegration{}
	integration.URL = "http://bbsim:" + port + "/" + version + "/bandwidthprofiles/%s"
	integration.Cache.Enabled = true
	integration.Cache.MaxSize = 40
	integration.Cache.TTL = "PT1m"

	bwp := &BandwidthProfileEntries{
		Integration: integration,
	}

	return bwp
}

func (s *SadisServer) ServeBaseConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	vars := mux.Vars(r)

	if vars["version"] != "v1" && vars["version"] != "v2" {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("{}"))
		return
	}

	sadisConf := GetSadisConfig(s.Olt, vars["version"])

	sadisJSON, _ := json.Marshal(sadisConf)

	_, _ = w.Write([]byte(sadisJSON))

}

func (s *SadisServer) ServeStaticConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	if vars["version"] == "v1" {
		// TODO format error
		http.Error(w, fmt.Sprintf("api-v1-unsupported"), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	sadisConf := GetSadisConfig(s.Olt, vars["version"])

	sadisConf.Sadis.Integration.URL = ""
	for i := range s.Olt.Pons {
		for _, onu := range s.Olt.Pons[i].Onus {
			if vars["version"] == "v2" {
				for _, u := range onu.UniPorts {
					uni := u.(*devices.UniPort)
					sonuV2, _ := GetOnuEntryV2(s.Olt, onu, fmt.Sprintf("%d", uni.ID+1))
					sadisConf.Sadis.Entries = append(sadisConf.Sadis.Entries, sonuV2)
				}
			}
		}
	}

	sadisConf.BandwidthProfile.Integration.URL = ""
	sadisConf.BandwidthProfile.Entries = bandwidthProfiles[common.Config.BBSim.BandwidthProfileFormat]

	sadisJSON, _ := json.Marshal(sadisConf)
	sadisLogger.Tracef("SADIS JSON: %s", sadisJSON)

	_, _ = w.Write([]byte(sadisJSON))

}

func (s *SadisServer) ServeEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	// check if the requested ID is for the OLT
	if s.Olt.SerialNumber == vars["ID"] {
		sadisLogger.WithFields(log.Fields{
			"OltSn": s.Olt.SerialNumber,
		}).Debug("Received SADIS OLT request")

		sadisConf, _ := GetOltEntry(s.Olt)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(sadisConf)
		return
	}

	i := strings.Split(vars["ID"], "-") // split ID to get serial number and uni port
	if len(i) != 2 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("{}"))
		sadisLogger.Warnf("Received invalid SADIS SubscriberId: %s", vars["ID"])
		return
	}
	sn, uni := i[0], i[len(i)-1]

	onu, err := s.Olt.FindOnuBySn(sn)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("{}"))
		sadisLogger.WithFields(log.Fields{
			"OnuSn": sn,
			"OnuId": "NA",
		}).Warnf("Requested Subscriber entry not found for OnuSn: %s", vars["ID"])
		return
	}

	sadisLogger.WithFields(log.Fields{
		"OnuId": onu.ID,
		"OnuSn": sn,
		"UniId": uni,
	}).Debug("Received SADIS request")

	if vars["version"] == "v1" {
		// TODO format error
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode("Sadis v1 is not supported anymore, please go back to an earlier BBSim version")
	} else if vars["version"] == "v2" {
		w.WriteHeader(http.StatusOK)
		sadisConf, _ := GetOnuEntryV2(s.Olt, onu, uni)
		_ = json.NewEncoder(w).Encode(sadisConf)
	}

}

func (s *SadisServer) ServeBWPEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["ID"]

	if vars["version"] != "v1" && vars["version"] != "v2" {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("{}"))
		return
	}

	sadisLogger.Debugf("Received request for SADIS bandwidth profile %s", id)

	for _, bwpEntry := range bandwidthProfiles[common.Config.BBSim.BandwidthProfileFormat] {
		if bwpEntry.ID == id {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(bwpEntry)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("{}"))
}
