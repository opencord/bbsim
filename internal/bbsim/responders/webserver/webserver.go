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

package webserver

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/bbsim/responders/sadis"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var logger = log.WithFields(log.Fields{
	"module": "WEBSERVER",
})

type imageRequestCount struct {
	Requests int `json:"requests"`
}

var imageRequests = 0

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

	// expose the requests counter
	router.HandleFunc("/images-count", func(w http.ResponseWriter, r *http.Request) {
		c := imageRequestCount{Requests: imageRequests}
		response, err := json.Marshal(c)

		if err != nil {
			logger.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("Cannot parse imageRequestCount to JSON")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(response)
	})

	// Choose the folder to serve (this is the location inside the container)
	staticDir := "/app/configs/"
	fileServer := http.FileServer(http.Dir(staticDir))

	// Create the route
	router.PathPrefix("/images/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, err := filepath.Abs(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		path = strings.Replace(path, "/images/", "", 1)

		path = filepath.Join(staticDir, path)

		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			// file does not exist, return 404
			http.Error(w, "file-not-found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		imageRequests = imageRequests + 1
		logger.WithFields(log.Fields{
			"count": imageRequests,
		}).Info("Got image request")

		http.StripPrefix("/images/", fileServer).ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(addr, router))

	wg.Done()
}
