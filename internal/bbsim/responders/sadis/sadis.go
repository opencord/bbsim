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

// SadisConfig is the top-level SADIS configuration struct
type SadisConfig struct {
	OrgOpencordSadis struct {
		Sadis            SadisEntries            `json:"sadis"`
		BandwidthProfile BandwidthProfileEntries `json:"bandwidthprofile"`
	} `json:"org.opencord.sadis"`
}

type SadisEntries struct {
	Integration SadisIntegration `json:"integration"`
	Entries     []interface{}    `json:"entries"`
}
type BandwidthProfileEntries struct {
	Integration SadisIntegration `json:"integration"`
	Entries     []interface{}    `json:"entries"`
}

type SadisIntegration struct {
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
	NasPortID                  string `json:"NasPortId"`
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
	CBS int    `json:"cba"`
	CIR int    `json:"cir"`
	EBS int    `json:"ebs"`
	EIR int    `json:"eir"`
}

// GetSadisConfig returns a full SADIS configuration struct ready to be marshalled into JSON
func GetSadisConfig(olt *devices.OltDevice) *SadisConfig {
	sadisEntries, _ := GetSadisEntries(olt)
	bwpEntries := getDummyBWPEntries()

	conf := &SadisConfig{}
	conf.OrgOpencordSadis.Sadis = *sadisEntries
	conf.OrgOpencordSadis.BandwidthProfile = *bwpEntries

	return conf
}

func GetSadisEntries(olt *devices.OltDevice) (*SadisEntries, error) {

	solt, _ := GetOltEntry(olt)
	entries := []interface{}{}
	entries = append(entries, solt)

	for i := range olt.Pons {
		for _, onu := range olt.Pons[i].Onus {
			sonu, _ := GetOnuEntry(olt, onu)
			entries = append(entries, sonu)
		}
	}

	integration := SadisIntegration{}
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

func GetOnuEntry(olt *devices.OltDevice, onu *devices.Onu) (*SadisOnuEntry, error) {
	idSuffix := "-1"
	sonu := &SadisOnuEntry{
		ID:                         onu.Sn() + idSuffix,
		CTag:                       onu.CTag,
		STag:                       onu.STag,
		NasPortID:                  onu.Sn() + idSuffix,
		CircuitID:                  onu.Sn() + idSuffix,
		RemoteID:                   olt.SerialNumber,
		TechnologyProfileID:        64,
		UpstreamBandwidthProfile:   "User_Bandwidth1",
		DownstreamBandwidthProfile: "Default",
	}

	return sonu, nil
}

func getDummyBWPEntries() *BandwidthProfileEntries {

	integration := SadisIntegration{}
	integration.Cache.Enabled = true
	integration.Cache.MaxSize = 40
	integration.Cache.TTL = "PT1m"

	bwp := &BandwidthProfileEntries{
		integration,
		[]interface{}{
			&SadisBWPEntry{ID: "User_Bandwidth1", AIR: 100000, CBS: 10000, CIR: 30000, EBS: 1000, EIR: 20000},
			&SadisBWPEntry{ID: "User_Bandwidth2", AIR: 100000, CBS: 5000, CIR: 100000, EBS: 5000, EIR: 1000000},
			&SadisBWPEntry{ID: "User_Bandwidth3", AIR: 100000, CBS: 5000, CIR: 100000, EBS: 5000, EIR: 1000000},
			&SadisBWPEntry{ID: "Default", AIR: 100000, CBS: 30, CIR: 600, EBS: 30, EIR: 400},
		},
	}

	return bwp
}

func (s *sadisServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	sadisConf := GetSadisConfig(s.olt)

	sadisJSON, _ := json.Marshal(sadisConf)
	sadisLogger.Tracef("SADIS JSON: %s", sadisJSON)

	w.Write([]byte(sadisJSON))

}

func (s *sadisServer) ServeEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	sn := strings.Split(vars["ID"], "-") // remove suffix from id to get serial number
	onu, err := s.olt.FindOnuBySn(sn[0])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{}"))
		sadisLogger.WithFields(log.Fields{
			"OnuSn": sn[0],
			"OnuId": "NA",
		}).Warn("Received SADIS request")
		return
	}

	sadisLogger.WithFields(log.Fields{
		"OnuSn": sn[0],
		"OnuId": onu.ID,
	}).Debug("Received SADIS request")

	sadisConf, err := GetOnuEntry(s.olt, onu)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sadisConf)
}

func (s *sadisServer) ServeBWPEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["ID"]
	sadisLogger.Debugf("Received request for SADIS bandwidth profile %s", id)

	bwp := getDummyBWPEntries()

	for _, e := range bwp.Entries {
		bwpEntry := e.(*SadisBWPEntry)
		if bwpEntry.ID == id { // is this case sensitive?
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
	sadisLogger.Infof("SADIS server listening on: %s", addr)
	s := &sadisServer{
		olt: olt,
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", s.ServeHTTP)
	router.HandleFunc("/subscriber/{ID}", s.ServeEntry)
	router.HandleFunc("/bandwidthprofiles/{ID}", s.ServeBWPEntry)

	log.Fatal(http.ListenAndServe(addr, router))

	wg.Done()
}
