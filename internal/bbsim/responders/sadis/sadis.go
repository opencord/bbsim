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
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
)

var sadisLogger = log.WithFields(log.Fields{
	"module": "SADIS",
})

type sadisServer struct {
	olt *devices.OltDevice
}

// bandwidthProfiles contains some dummy profiles
var bandwidthProfiles = []interface{}{
	&SadisBWPEntry{ID: "User_Bandwidth1", AIR: 100000, CBS: 10000, CIR: 30000, EBS: 1000, EIR: 20000},
	&SadisBWPEntry{ID: "User_Bandwidth2", AIR: 100000, CBS: 5000, CIR: 100000, EBS: 5000, EIR: 1000000},
	&SadisBWPEntry{ID: "User_Bandwidth3", AIR: 100000, CBS: 5000, CIR: 100000, EBS: 5000, EIR: 1000000},
	&SadisBWPEntry{ID: "Default", AIR: 100000, CBS: 30, CIR: 600, EBS: 30, EIR: 400},
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
	Entries     []interface{}    `json:"entries,omitempty"`
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
}

type SadisOnuEntry struct {
	ID                         string `json:"id"`
	CTag                       int    `json:"cTag"`
	STag                       int    `json:"sTag"`
	NasPortID                  string `json:"nasPortId"`
	CircuitID                  string `json:"circuitId"`
	RemoteID                   string `json:"remoteId"`
	TechnologyProfileID        int    `json:"technologyProfileId"`
	UpstreamBandwidthProfile   string `json:"upstreamBandwidthProfile"`
	DownstreamBandwidthProfile string `json:"downstreamBandwidthProfile"`
}

// SADIS BandwithProfile Entry
type SadisBWPEntry struct {
	ID  string `json:"id"`
	AIR int    `json:"air"`
	CBS int    `json:"cbs"`
	CIR int    `json:"cir"`
	EBS int    `json:"ebs"`
	EIR int    `json:"eir"`
}

// GetSadisConfig returns a full SADIS configuration struct ready to be marshalled into JSON
func GetSadisConfig(olt *devices.OltDevice) *SadisConfig {
	sadisEntries, _ := GetSadisEntries(olt)
	bwpEntries := getBWPEntries()

	conf := &SadisConfig{}
	conf.Sadis = *sadisEntries
	conf.BandwidthProfile = *bwpEntries

	return conf
}

func GetSadisEntries(olt *devices.OltDevice) (*SadisEntries, error) {
	solt, _ := GetOltEntry(olt)

	entries := []interface{}{}
	entries = append(entries, solt)

	a := strings.Split(common.Options.BBSim.SadisRestAddress, ":")
	port := a[len(a)-1]

	integration := SadisIntegration{}
	integration.URL = "http://bbsim:" + port + "/subscribers/%s"
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
		HardwareIdentifier: common.Options.Olt.DeviceId,
		IPAddress:          ip,
		NasID:              olt.SerialNumber,
		UplinkPort:         1048576, // TODO currently assumes we only have on NNI port
	}
	return solt, nil
}

func GetOnuEntry(olt *devices.OltDevice, onu *devices.Onu, uniId string) (*SadisOnuEntry, error) {
	uniSuffix := "-" + uniId
	sonu := &SadisOnuEntry{
		ID:                         onu.Sn() + uniSuffix,
		CTag:                       onu.CTag,
		STag:                       onu.STag,
		NasPortID:                  onu.Sn() + uniSuffix,
		CircuitID:                  onu.Sn() + uniSuffix,
		RemoteID:                   olt.SerialNumber,
		TechnologyProfileID:        64,
		UpstreamBandwidthProfile:   "User_Bandwidth1",
		DownstreamBandwidthProfile: "Default",
	}

	return sonu, nil
}

func getBWPEntries() *BandwidthProfileEntries {
	a := strings.Split(common.Options.BBSim.SadisRestAddress, ":")
	port := a[len(a)-1]

	integration := SadisIntegration{}
	integration.URL = "http://bbsim:" + port + "/bandwidthprofiles/%s"
	integration.Cache.Enabled = true
	integration.Cache.MaxSize = 40
	integration.Cache.TTL = "PT1m"

	bwp := &BandwidthProfileEntries{
		Integration: integration,
	}

	return bwp
}

func (s *sadisServer) ServeBaseConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	sadisConf := GetSadisConfig(s.olt)

	sadisJSON, _ := json.Marshal(sadisConf)
	sadisLogger.Tracef("SADIS JSON: %s", sadisJSON)

	w.Write([]byte(sadisJSON))

}

func (s *sadisServer) ServeStaticConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	sadisConf := GetSadisConfig(s.olt)

	sadisConf.Sadis.Integration.URL = ""
	for i := range s.olt.Pons {
		for _, onu := range s.olt.Pons[i].Onus {
			// FIXME currently we only support one UNI per ONU
			sonu, _ := GetOnuEntry(s.olt, onu, "1")
			sadisConf.Sadis.Entries = append(sadisConf.Sadis.Entries, sonu)
		}
	}

	sadisConf.BandwidthProfile.Integration.URL = ""
	sadisConf.BandwidthProfile.Entries = bandwidthProfiles

	sadisJSON, _ := json.Marshal(sadisConf)
	sadisLogger.Tracef("SADIS JSON: %s", sadisJSON)

	w.Write([]byte(sadisJSON))

}

func (s *sadisServer) ServeEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	// check if the requested ID is for the OLT
	if s.olt.SerialNumber == vars["ID"] {
		sadisLogger.WithFields(log.Fields{
			"OltSn": s.olt.SerialNumber,
		}).Debug("Received SADIS OLT request")

		sadisConf, _ := GetOltEntry(s.olt)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(sadisConf)
		return
	}

	i := strings.Split(vars["ID"], "-") // split ID to get serial number and uni port
	if len(i) != 2 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("{}"))
		sadisLogger.Warnf("Received invalid SADIS subscriber request: %s", vars["ID"])
		return
	}
	sn, uni := i[0], i[len(i)-1]

	onu, err := s.olt.FindOnuBySn(sn)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{}"))
		sadisLogger.WithFields(log.Fields{
			"OnuSn": sn,
			"OnuId": "NA",
		}).Warnf("Received invalid SADIS subscriber request: %s", vars["ID"])
		return
	}

	sadisLogger.WithFields(log.Fields{
		"OnuId":     onu.ID,
		"OnuSn":     sn,
		"OnuPortNo": uni,
	}).Debug("Received SADIS request")

	sadisConf, err := GetOnuEntry(s.olt, onu, uni)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sadisConf)
}

func (s *sadisServer) ServeBWPEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["ID"]
	sadisLogger.Debugf("Received request for SADIS bandwidth profile %s", id)

	for _, e := range bandwidthProfiles {
		bwpEntry := e.(*SadisBWPEntry)
		if bwpEntry.ID == id {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(bwpEntry)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("{}"))
}

// StartRestServer starts REST server which returns a SADIS configuration for the currently simulated OLT
func StartRestServer(olt *devices.OltDevice, wg *sync.WaitGroup) {
	addr := common.Options.BBSim.SadisRestAddress
	sadisLogger.Infof("SADIS server listening on %s", addr)
	s := &sadisServer{
		olt: olt,
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/cfg", s.ServeBaseConfig)
	router.HandleFunc("/static", s.ServeStaticConfig)
	router.HandleFunc("/subscribers/{ID}", s.ServeEntry)
	router.HandleFunc("/bandwidthprofiles/{ID}", s.ServeBWPEntry)

	log.Fatal(http.ListenAndServe(addr, router))

	wg.Done()
}
