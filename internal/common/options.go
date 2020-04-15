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
	"errors"
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
)

var tagAllocationValues = []string{
	"unknown",
	"shared",
	"unique",
}

type TagAllocation int

func (t TagAllocation) String() string {
	return tagAllocationValues[t]
}

func tagAllocationFromString(s string) (TagAllocation, error) {
	for i, v := range tagAllocationValues {
		if v == s {
			return TagAllocation(i), nil
		}
	}
	log.WithFields(log.Fields{
		"ValidValues": strings.Join(tagAllocationValues[1:], ", "),
	}).Errorf("%s-is-not-a-valid-tag-allocation", s)
	return TagAllocation(0), errors.New(fmt.Sprintf("%s-is-not-a-valid-tag-allocation", s))
}

const (
	_ TagAllocation = iota
	TagAllocationShared
	TagAllocationUnique
)


var sadisFormatValues = []string{
	"unknown",
	"att",
	"dt",
	"tt",
}

type SadisFormat int

func (s SadisFormat) String() string {
	return sadisFormatValues[s]
}

func sadisFormatFromString(s string) (SadisFormat, error) {
	for i, v := range sadisFormatValues {
		if v == s {
			return SadisFormat(i), nil
		}
	}
	log.WithFields(log.Fields{
		"ValidValues": strings.Join(sadisFormatValues[1:], ", "),
	}).Errorf("%s-is-not-a-valid-sadis-format", s)
	return SadisFormat(0), errors.New(fmt.Sprintf("%s-is-not-a-valid-sadis-format", s))
}

const (
	_ SadisFormat = iota
	SadisFormatAtt
	SadisFormatDt
	SadisFormatTt
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
	PortStatsInterval  int    `yaml: "port_stats_interval"`
}

type BBSimConfig struct {
	EnableDhcp           bool          `yaml:"enable_dhcp"`
	EnableAuth           bool          `yaml:"enable_auth"`
	LogLevel             string        `yaml:"log_level"`
	LogCaller            bool          `yaml:"log_caller"`
	Delay                int           `yaml:"delay"`
	CpuProfile           *string       `yaml:"cpu_profile"`
	CTagAllocation       TagAllocation `yaml:"c_tag_allocation"`
	CTag                 int           `yaml:"c_tag"`
	STagAllocation       TagAllocation `yaml:"s_tag_allocation"`
	STag                 int           `yaml:"s_tag"`
	OpenOltAddress       string        `yaml:"openolt_address"`
	ApiAddress           string        `yaml:"api_address"`
	RestApiAddress       string        `yaml:"rest_api_address"`
	LegacyApiAddress     string        `yaml:"legacy_api_address"`
	LegacyRestApiAddress string        `yaml:"legacy_rest_api_address"`
	SadisRestAddress     string        `yaml:"sadis_rest_address"`
	SadisServer          bool          `yaml:"sadis_server"`
	SadisFormat          SadisFormat   `yaml:"sadis_format"`
	KafkaAddress         string        `yaml:"kafka_address"`
	Events               bool          `yaml:"enable_events"`
	ControlledActivation string        `yaml:"controlled_activation"`
	EnablePerf           bool          `yaml:"enable_perf"`
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
			STagAllocation:       TagAllocationShared,
			STag:                 900,
			CTagAllocation:       TagAllocationUnique,
			CTag:                 900,
			EnableDhcp:           false,
			EnableAuth:           false,
			LogLevel:             "debug",
			LogCaller:            false,
			Delay:                200,
			OpenOltAddress:       ":50060",
			ApiAddress:           ":50070",
			RestApiAddress:       ":50071",
			LegacyApiAddress:     ":50072",
			LegacyRestApiAddress: ":50073",
			SadisRestAddress:     ":50074",
			SadisServer:          true,
			SadisFormat:          SadisFormatAtt,
			KafkaAddress:         ":9092",
			Events:               false,
			ControlledActivation: "default",
			EnablePerf:           false,
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
			PortStatsInterval:  20,
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

	// TODO convert from string to TagAllocation

	return yamlConfig, nil
}

// GetBBSimOpts loads the BBSim configuration file and over-rides options with corresponding CLI flags if set
func GetBBSimOpts() *BBSimYamlConfig {
	conf := Options

	olt_id := flag.Int("olt_id", conf.Olt.ID, "OLT device ID")
	nni := flag.Int("nni", int(conf.Olt.NniPorts), "Number of NNI ports per OLT device to be emulated")
	pon := flag.Int("pon", int(conf.Olt.PonPorts), "Number of PON ports per OLT device to be emulated")
	onu := flag.Int("onu", int(conf.Olt.OnusPonPort), "Number of ONU devices per PON port to be emulated")

	openolt_address := flag.String("openolt_address", conf.BBSim.OpenOltAddress, "IP address:port")
	api_address := flag.String("api_address", conf.BBSim.ApiAddress, "IP address:port")
	rest_api_address := flag.String("rest_api_address", conf.BBSim.RestApiAddress, "IP address:port")

	s_tag_allocation := flag.String("s_tag_allocation", conf.BBSim.STagAllocation.String(), "Use 'unique' for incremental values, 'shared' to use the same value in all the ONUs")
	s_tag := flag.Int("s_tag", conf.BBSim.STag, "S-Tag initial value")

	c_tag_allocation := flag.String("c_tag_allocation", conf.BBSim.CTagAllocation.String(), "Use 'unique' for incremental values, 'shared' to use the same value in all the ONUs")
	c_tag := flag.Int("c_tag", conf.BBSim.CTag, "C-Tag starting value, each ONU will get a sequential one (targeting 1024 ONUs per BBSim instance the range is big enough)")

	sadisFormat := flag.String("sadisFormat", conf.BBSim.SadisFormat.String(), fmt.Sprintf("Which format should sadis expose? [%s]", strings.Join(sadisFormatValues[1:], "|")))

	auth := flag.Bool("auth", conf.BBSim.EnableAuth, "Set this flag if you want authentication to start automatically")
	dhcp := flag.Bool("dhcp", conf.BBSim.EnableDhcp, "Set this flag if you want DHCP to start automatically")

	profileCpu := flag.String("cpuprofile", "", "write cpu profile to file")

	logLevel := flag.String("logLevel", conf.BBSim.LogLevel, "Set the log level (trace, debug, info, warn, error)")
	logCaller := flag.Bool("logCaller", conf.BBSim.LogCaller, "Whether to print the caller filename or not")

	delay := flag.Int("delay", conf.BBSim.Delay, "The delay between ONU DISCOVERY batches in milliseconds (1 ONU per each PON PORT at a time")

	controlledActivation := flag.String("ca", conf.BBSim.ControlledActivation, "Set the mode for controlled activation of PON ports and ONUs")
	enablePerf := flag.Bool("enableperf", conf.BBSim.EnablePerf, "Setting this flag will cause BBSim to not store data like traffic schedulers, flows of ONUs etc..")
	enableEvents := flag.Bool("enableEvents", conf.BBSim.Events, "Enable sending BBSim events on configured kafka server")
	kafkaAddress := flag.String("kafkaAddress", conf.BBSim.KafkaAddress, "IP:Port for kafka")
	flag.Parse()

	sTagAlloc, err := tagAllocationFromString(*s_tag_allocation)
	if err != nil {
		log.Fatal(err)
	}

	cTagAlloc, err := tagAllocationFromString(*c_tag_allocation)
	if err != nil {
		log.Fatal(err)
	}

	sf, err := sadisFormatFromString(*sadisFormat)
	if err != nil {
		log.Fatal(err)
	}

	if sf == SadisFormatTt {
		log.Fatalf("Sadis format %s is not yet supported", sf.String())
	}

	conf.Olt.ID = int(*olt_id)
	conf.Olt.NniPorts = uint32(*nni)
	conf.Olt.PonPorts = uint32(*pon)
	conf.Olt.OnusPonPort = uint32(*onu)
	conf.BBSim.STagAllocation = sTagAlloc
	conf.BBSim.STag = int(*s_tag)
	conf.BBSim.CTagAllocation = cTagAlloc
	conf.BBSim.CTag = int(*c_tag)
	conf.BBSim.CpuProfile = profileCpu
	conf.BBSim.LogLevel = *logLevel
	conf.BBSim.LogCaller = *logCaller
	conf.BBSim.EnableAuth = *auth
	conf.BBSim.EnableDhcp = *dhcp
	conf.BBSim.Delay = *delay
	conf.BBSim.ControlledActivation = *controlledActivation
	conf.BBSim.EnablePerf = *enablePerf
	conf.BBSim.Events = *enableEvents
	conf.BBSim.KafkaAddress = *kafkaAddress
	conf.BBSim.OpenOltAddress = *openolt_address
	conf.BBSim.ApiAddress = *api_address
	conf.BBSim.RestApiAddress = *rest_api_address
	conf.BBSim.SadisFormat = sf

	// update device id if not set
	if conf.Olt.DeviceId == "" {
		conf.Olt.DeviceId = net.HardwareAddr{0xA, 0xA, 0xA, 0xA, 0xA, byte(conf.Olt.ID)}.String()
	}

	Options = conf
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
