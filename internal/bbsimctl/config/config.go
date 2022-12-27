/*
 * Portions Copyright 2019-2023 Open Networking Foundation (ONF) and the ONF Contributors
 * Original copyright 2019-present Ciena Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
	"time"
)

var (
	Version    string
	BuildTime  string
	CommitHash string
	GitStatus  string
)

var GlobalOptions struct {
	Config string `short:"c" long:"config" env:"BBSIMCTL_CONFIG" value-name:"FILE" default:"" description:"Location of client config file"`
	Server string `short:"s" long:"server" default:"" value-name:"SERVER:PORT" description:"IP/Host and port of XOS"`
	//Protoset string `long:"protoset" value-name:"FILENAME" description:"Load protobuf definitions from protoset instead of reflection api"`
	Debug bool `short:"d" long:"debug" description:"Enable debug mode"`
}

type GrpcConfigSpec struct {
	Timeout time.Duration `yaml:"timeout"`
}

type GlobalConfigSpec struct {
	Server string `yaml:"server"`
	Grpc   GrpcConfigSpec
}

var GlobalConfig = GlobalConfigSpec{
	Server: "localhost:50070",
	Grpc: GrpcConfigSpec{
		Timeout: time.Second * 10,
	},
}

var DmiConfig = GlobalConfigSpec{
	Server: "localhost:50075",
	Grpc: GrpcConfigSpec{
		Timeout: time.Second * 10,
	},
}

func ProcessGlobalOptions() {
	if len(GlobalOptions.Config) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Unable to discover the users home directory: %s\n", err)
		}
		GlobalOptions.Config = fmt.Sprintf("%s/.bbsim/config", home)
	}

	info, err := os.Stat(GlobalOptions.Config)
	if err == nil && !info.IsDir() {
		configFile, err := ioutil.ReadFile(GlobalOptions.Config)
		if err != nil {
			log.Printf("configFile.Get err   #%v ", err)
		}
		err = yaml.Unmarshal(configFile, &GlobalConfig)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}
	}

	// Override from environment
	//    in particualr, for passing env vars via `go test`
	env_server, present := os.LookupEnv("BBSIMCTL_SERVER")
	if present {
		GlobalConfig.Server = env_server
	}

	// Override from command line
	if GlobalOptions.Server != "" {
		GlobalConfig.Server = GlobalOptions.Server
	}

	// Generate error messages for required settings
	if GlobalConfig.Server == "" {
		log.Fatal("Server is not set. Please update config file or use the -s option")
	}

	//Try to resolve hostname if provided for the server
	if host, port, err := net.SplitHostPort(GlobalConfig.Server); err == nil {
		if addrs, err := net.LookupHost(host); err == nil {
			GlobalConfig.Server = net.JoinHostPort(addrs[0], port)
		}
	}
}
