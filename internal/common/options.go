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
	"flag"
	"fmt"
	"io/ioutil"
	"net"

	"gopkg.in/yaml.v2"
)

type BBRCliOptions struct {
	*BBSimYamlConfig
	BBSimIp      string
	BBSimPort    string
	BBSimApiPort string
	LogFile      string
}

type BBSimYamlConfig struct {
	BBSim BBSimConfig
	Olt   OltConfig
	BBR   BBRConfig
}

type OltConfig struct {
	Model              string `yaml:"model"`
	Vendor             string `yaml:"vendor"`
	HardwareVersion    string `yaml:"hardware_version"`
	FirmwareVersion    string `yaml:"firmware_version"`
	DeviceId           string `yaml:"device_id"`
	DeviceSerialNumber string `yaml:"device_serial_number"`
	PonPorts           uint32 `yaml:"pon_ports"`
	NniPorts           uint32 `yaml:"nni_ports"`
	OnusPonPort        uint32 `yaml:"onus_per_port"`
	Technology         string `yaml:"technology"`
	ID                 int    `yaml:"id"`
	OltRebootDelay     int    `yaml:"reboot_delay"`
}

type BBSimConfig struct {
	EnableDhcp           bool    `yaml:"enable_dhcp"`
	EnableAuth           bool    `yaml:"enable_auth"`
	LogLevel             string  `yaml:"log_level"`
	LogCaller            bool    `yaml:"log_caller"`
	Delay                int     `yaml:"delay"`
	CpuProfile           *string `yaml:"cpu_profile"`
	CTagInit             int     `yaml:"c_tag"`
	STag                 int     `yaml:"s_tag"`
	OpenOltAddress       string  `yaml:"openolt_address"`
	ApiAddress           string  `yaml:"api_address"`
	RestApiAddress       string  `yaml:"rest_api_address"`
	LegacyApiAddress     string  `yaml:"legacy_api_address"`
	LegacyRestApiAddress string  `yaml:"legacy_rest_api_address"`
}

type BBRConfig struct {
	Log       string `yaml:"log"`
	LogLevel  string `yaml:"log_level"`
	LogCaller bool   `yaml:"log_caller"`
}

var Options *BBSimYamlConfig

func init() {
	// load settings from config file first
	Options, _ = LoadBBSimConf("configs/bbsim.yaml")
}

func getDefaultOps() *BBSimYamlConfig {

	c := &BBSimYamlConfig{
		BBSimConfig{
			STag:                 900,
			CTagInit:             900,
			EnableDhcp:           false,
			EnableAuth:           false,
			LogLevel:             "debug",
			LogCaller:            false,
			Delay:                200,
			OpenOltAddress:       "0.0.0.0:50060",
			ApiAddress:           "0.0.0.0:50070",
			RestApiAddress:       "0.0.0.0:50071",
			LegacyApiAddress:     "0.0.0.0:50072",
			LegacyRestApiAddress: "0.0.0.0:50073",
		},
		OltConfig{
			Vendor:             "BBSim",
			Model:              "asfvolt16",
			HardwareVersion:    "emulated",
			FirmwareVersion:    "",
			DeviceSerialNumber: "BBSM00000001",
			PonPorts:           1,
			NniPorts:           1,
			OnusPonPort:        1,
			Technology:         "XGS-PON",
			ID:                 0,
			OltRebootDelay:     10,
		},
		BBRConfig{
			LogLevel:  "debug",
			LogCaller: false,
		},
	}
	return c
}

// LoadBBSimConf loads the BBSim configuration from a YAML file
func LoadBBSimConf(filename string) (*BBSimYamlConfig, error) {
	yamlConfig := getDefaultOps()

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Cannot load BBSim configuration file: %s. Using defaults.\n", err)
		return yamlConfig, nil
	}

	err = yaml.Unmarshal(yamlFile, yamlConfig)
	if err != nil {
		fmt.Printf("Error parsing YAML file: %s\n", err)
	}

	return yamlConfig, nil
}

// GetBBSimOpts loads the BBSim configuration file and overides options with corresponding CLI flags if set
func GetBBSimOpts() *BBSimYamlConfig {
	conf := Options

	olt_id := flag.Int("olt_id", conf.Olt.ID, "OLT device ID")
	nni := flag.Int("nni", int(conf.Olt.NniPorts), "Number of NNI ports per OLT device to be emulated")
	pon := flag.Int("pon", int(conf.Olt.PonPorts), "Number of PON ports per OLT device to be emulated")
	onu := flag.Int("onu", int(conf.Olt.OnusPonPort), "Number of ONU devices per PON port to be emulated")

	s_tag := flag.Int("s_tag", conf.BBSim.STag, "S-Tag initial value")
	c_tag_init := flag.Int("c_tag", conf.BBSim.CTagInit, "C-Tag starting value, each ONU will get a sequential one (targeting 1024 ONUs per BBSim instance the range is big enough)")

	auth := flag.Bool("auth", conf.BBSim.EnableAuth, "Set this flag if you want authentication to start automatically")
	dhcp := flag.Bool("dhcp", conf.BBSim.EnableDhcp, "Set this flag if you want DHCP to start automatically")

	profileCpu := flag.String("cpuprofile", "", "write cpu profile to file")

	logLevel := flag.String("logLevel", conf.BBSim.LogLevel, "Set the log level (trace, debug, info, warn, error)")
	logCaller := flag.Bool("logCaller", conf.BBSim.LogCaller, "Whether to print the caller filename or not")

	delay := flag.Int("delay", conf.BBSim.Delay, "The delay between ONU DISCOVERY batches in milliseconds (1 ONU per each PON PORT at a time")

	flag.Parse()

	conf.Olt.ID = int(*olt_id)
	conf.Olt.NniPorts = uint32(*nni)
	conf.Olt.PonPorts = uint32(*pon)
	conf.Olt.OnusPonPort = uint32(*onu)
	conf.BBSim.STag = int(*s_tag)
	conf.BBSim.CTagInit = int(*c_tag_init)
	conf.BBSim.CpuProfile = profileCpu
	conf.BBSim.LogLevel = *logLevel
	conf.BBSim.LogCaller = *logCaller
	conf.BBSim.EnableAuth = *auth
	conf.BBSim.EnableDhcp = *dhcp
	conf.BBSim.Delay = *delay

	// update device id if not set
	if conf.Olt.DeviceId == "" {
		conf.Olt.DeviceId = net.HardwareAddr{0xA, 0xA, 0xA, 0xA, 0xA, byte(conf.Olt.ID)}.String()
	}

	return conf
}

func GetBBROpts() BBRCliOptions {

	bbsimIp := flag.String("bbsimIp", "127.0.0.1", "BBSim IP")
	bbsimPort := flag.String("bbsimPort", "50060", "BBSim Port")
	bbsimApiPort := flag.String("bbsimApiPort", "50070", "BBSim API Port")
	logFile := flag.String("logfile", "", "Log to a file")

	options := GetBBSimOpts()

	bbrOptions := BBRCliOptions{
		options,
		*bbsimIp,
		*bbsimPort,
		*bbsimApiPort,
		*logFile,
	}

	return bbrOptions
}
