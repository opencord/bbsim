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

package webserver

import (
	"github.com/gorilla/mux"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/bbsim/responders/sadis"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

var logger = log.WithFields(log.Fields{
	"module": "WEBSERVER",
})

// StartRestServer starts REST server which esposes:
// - a SADIS configuration for the currently simulated OLT
// - static files for software image download
func StartRestServer(olt *devices.OltDevice, wg *sync.WaitGroup) {
	addr := common.Config.BBSim.SadisRestAddress
	logger.Infof("WEBSERVER server listening on %s", addr)
	s := &sadis.SadisServer{
		Olt: olt,
	}

	router := mux.NewRouter().StrictSlash(true)

	// sadis routes
	router.HandleFunc(sadis.BaseConfigUrl, s.ServeBaseConfig)
	router.HandleFunc(sadis.StaticConfigUrl, s.ServeStaticConfig)
	router.HandleFunc(sadis.SadisEntryUrl, s.ServeEntry)
	router.HandleFunc(sadis.SadisBwUrl, s.ServeBWPEntry)

	// Choose the folder to serve (this is the location inside the container)
	staticDir := "/app/configs/"

	// Create the route
	router.
		PathPrefix("/images/").
		Handler(http.StripPrefix("/images/", http.FileServer(http.Dir(staticDir))))

	log.Fatal(http.ListenAndServe(addr, router))

	wg.Done()
}
