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
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var tagAllocationValues = []string{
	"unknown",
	"shared",
	"unique",
}

const (
	BP_FORMAT_MEF  = "mef"
	BP_FORMAT_IETF = "ietf"
)

type TagAllocation int

func (t TagAllocation) String() string {
	return tagAllocationValues[t]
}

func tagAllocationFromString(s string) (TagAllocation, error) {
	for i, v := range tagAllocationValues {
		if v == strings.TrimSpace(s) {
			return TagAllocation(i), nil
		}
	}
	log.WithFields(log.Fields{
		"ValidValues": strings.Join(tagAllocationValues[1:], ", "),
	}).Errorf("%s-is-not-a-valid-tag-allocation", s)
	return TagAllocation(0), fmt.Errorf("%s-is-not-a-valid-tag-allocation", s)
}

const (
	_ TagAllocation = iota
	TagAllocationShared
	TagAllocationUnique
)

type BBRCliOptions struct {
	*GlobalConfig
	BBSimIp      string
	BBSimPort    string
	BBSimApiPort string
	LogFile      string
}

type GlobalConfig struct {
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
	NniSpeed           uint32 `yaml:"nni_speed"`
	OnusPonPort        uint32 `yaml:"onus_per_port"`
	ID                 int    `yaml:"id"`
	OltRebootDelay     int    `yaml:"reboot_delay"`
	PortStatsInterval  int    `yaml:"port_stats_interval"`
	OmciResponseRate   uint8  `yaml:"omci_response_rate"`
	UniPorts           uint32 `yaml:"uni_ports"`
	PotsPorts          uint32 `yaml:"pots_ports"`
	NniDhcpTrapVid     uint32 `yaml:"nni_dhcp_trap_vid"`
}

type PonPortsConfig struct {
	Number uint32           `yaml:"num_pon_ports"`
	Ranges []PonRangeConfig `yaml:"ranges"`
}

type IdRange struct {
	StartId uint32 `yaml:"start"`
	EndId   uint32 `yaml:"end"`
}

type PonTechnology int

var ponTechnologyValues = []string{
	"GPON", "XGS-PON",
}

func (t PonTechnology) String() string {
	return ponTechnologyValues[t]
}

const (
	GPON PonTechnology = iota
	XGSPON
)

func PonTechnologyFromString(s string) (PonTechnology, error) {
	for i, val := range ponTechnologyValues {
		if val == s {
			return PonTechnology(i), nil
		}
	}
	log.WithFields(log.Fields{
		"ValidValues": strings.Join(ponTechnologyValues[:], ", "),
	}).Errorf("%s-is-not-a-valid-pon-technology", s)
	return -1, fmt.Errorf("%s-is-not-a-valid-pon-technology", s)
}

//Constants for default allocation ranges
const (
	defaultOnuIdStart          = 1
	defaultAllocIdStart        = 1024
	defaultGemPortIdPerAllocId = 8
	defaultGemportIdStart      = 1024
)

type PonRangeConfig struct {
	PonRange     IdRange `yaml:"pon_id_range"`
	Technology   string  `yaml:"tech"`
	OnuRange     IdRange `yaml:"onu_id_range"`
	AllocIdRange IdRange `yaml:"alloc_id_range"`
	GemportRange IdRange `yaml:"gemport_id_range"`
}

func GetPonConfigById(id uint32) (*PonRangeConfig, error) {
	if PonsConfig == nil {
		return nil, fmt.Errorf("pons-config-nil")
	}

	for _, r := range PonsConfig.Ranges {
		if id >= r.PonRange.StartId && id <= r.PonRange.EndId {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("pon-config-for-id-%d-not-found", id)
}

type BBSimConfig struct {
	ConfigFile                  string
	ServiceConfigFile           string
	PonsConfigFile              string
	DhcpRetry                   bool    `yaml:"dhcp_retry"`
	AuthRetry                   bool    `yaml:"auth_retry"`
	LogLevel                    string  `yaml:"log_level"`
	LogCaller                   bool    `yaml:"log_caller"`
	Delay                       int     `yaml:"delay"`
	CpuProfile                  *string `yaml:"cpu_profile"`
	OpenOltAddress              string  `yaml:"openolt_address"`
	ApiAddress                  string  `yaml:"api_address"`
	RestApiAddress              string  `yaml:"rest_api_address"`
	LegacyApiAddress            string  `yaml:"legacy_api_address"`
	LegacyRestApiAddress        string  `yaml:"legacy_rest_api_address"`
	SadisRestAddress            string  `yaml:"sadis_rest_address"`
	SadisServer                 bool    `yaml:"sadis_server"`
	KafkaAddress                string  `yaml:"kafka_address"`
	Events                      bool    `yaml:"enable_events"`
	ControlledActivation        string  `yaml:"controlled_activation"`
	EnablePerf                  bool    `yaml:"enable_perf"`
	KafkaEventTopic             string  `yaml:"kafka_event_topic"`
	DmiServerAddress            string  `yaml:"dmi_server_address"`
	BandwidthProfileFormat      string  `yaml:"bp_format"`
	InjectOmciUnknownMe         bool    `yaml:"inject_omci_unknown_me"`
	InjectOmciUnknownAttributes bool    `yaml:"inject_omci_unknown_attributes"`
	OmccVersion                 int     `yaml:"omcc_version"`
}

type BBRConfig struct {
	Log       string `yaml:"log"`
	LogLevel  string `yaml:"log_level"`
	LogCaller bool   `yaml:"log_caller"`
}

type ServiceYaml struct {
	Name                string
	CTag                int    `yaml:"c_tag"`
	STag                int    `yaml:"s_tag"`
	NeedsEapol          bool   `yaml:"needs_eapol"`
	NeedsDhcp           bool   `yaml:"needs_dhcp"`
	NeedsIgmp           bool   `yaml:"needs_igmp"`
	NeedsPPPoE          bool   `yaml:"needs_pppoe"`
	CTagAllocation      string `yaml:"c_tag_allocation"`
	STagAllocation      string `yaml:"s_tag_allocation"`
	TechnologyProfileID int    `yaml:"tp_id"`
	UniTagMatch         int    `yaml:"uni_tag_match"`
	ConfigureMacAddress bool   `yaml:"configure_mac_address"`
	EnableMacLearning   bool   `yaml:"enable_mac_learning"`
	UsPonCTagPriority   uint8  `yaml:"us_pon_c_tag_priority"`
	UsPonSTagPriority   uint8  `yaml:"us_pon_s_tag_priority"`
	DsPonCTagPriority   uint8  `yaml:"ds_pon_c_tag_priority"`
	DsPonSTagPriority   uint8  `yaml:"ds_pon_s_tag_priority"`
}
type YamlServiceConfig struct {
	Workflow string
	Services []ServiceYaml `yaml:"services,flow"`
}

func (cfg *YamlServiceConfig) String() string {
	str := fmt.Sprintf("[workflow: %s, Services: ", cfg.Workflow)

	for _, s := range cfg.Services {
		str = fmt.Sprintf("%s[", str)
		str = fmt.Sprintf("%sname=%s, c_tag=%d, s_tag=%d, ",
			str, s.Name, s.CTag, s.STag)
		str = fmt.Sprintf("%sc_tag_allocation=%s, s_tag_allocation=%s, ",
			str, s.CTagAllocation, s.STagAllocation)
		str = fmt.Sprintf("%sneeds_eapol=%t, needs_dhcp=%t, needs_igmp=%t",
			str, s.NeedsEapol, s.NeedsDhcp, s.NeedsIgmp)
		str = fmt.Sprintf("%stp_id=%d, uni_tag_match=%d",
			str, s.TechnologyProfileID, s.UniTagMatch)
		str = fmt.Sprintf("%s]", str)
	}
	str = fmt.Sprintf("%s]", str)
	return str
}

var (
	Config     *GlobalConfig
	Services   []ServiceYaml
	PonsConfig *PonPortsConfig
)

// Load the BBSim configuration. This is a combination of CLI parameters and YAML files
// We proceed in this order:
// - Read CLI parameters
// - Using those we read the yaml files (config and services)
// - we merge the configuration (CLI has priority over yaml files)
func LoadConfig() {

	Config = GetDefaultOps()

	cliConf := readCliParams()

	yamlConf, err := loadBBSimConf(cliConf.BBSim.ConfigFile)

	if err != nil {
		log.WithFields(log.Fields{
			"file": cliConf.BBSim.ConfigFile,
			"err":  err,
		}).Fatal("Can't read config file")
	}

	// merging Yaml and Default Values
	if err := mergo.Merge(Config, yamlConf, mergo.WithOverride); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Can't merge YAML and Config")
	}

	// merging CLI values on top of the yaml ones
	if err := mergo.Merge(Config, cliConf, mergo.WithOverride); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Can't merge CLI and Config")
	}

	services, err := loadBBSimServices(Config.BBSim.ServiceConfigFile)

	if err != nil {
		log.WithFields(log.Fields{
			"file": Config.BBSim.ServiceConfigFile,
			"err":  err,
		}).Fatal("Can't read services file")
	}

	Services = services

	//A blank filename means we should fall back to bbsim defaults
	if Config.BBSim.PonsConfigFile == "" {
		PonsConfig, err = getDefaultPonsConfig()
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatal("Can't load Pon interfaces defaults.")
		}
	} else {
		PonsConfig, err = loadBBSimPons(Config.BBSim.PonsConfigFile)

		if err != nil {
			log.WithFields(log.Fields{
				"file": Config.BBSim.PonsConfigFile,
				"err":  err,
			}).Fatal("Can't read services file")
		}
	}

	if err := validatePonsConfig(PonsConfig); err != nil {
		log.WithFields(log.Fields{
			"file": Config.BBSim.PonsConfigFile,
			"err":  err,
		}).Fatal("Invalid Pon interfaces configuration")
	}
}

func readCliParams() *GlobalConfig {

	conf := GetDefaultOps()

	configFile := flag.String("config", conf.BBSim.ConfigFile, "Configuration file path")
	servicesFile := flag.String("services", conf.BBSim.ServiceConfigFile, "Service Configuration file path")
	ponsFile := flag.String("pon_port_config_file", conf.BBSim.PonsConfigFile, "Pon Interfaces Configuration file path")
	sadisBpFormat := flag.String("bp_format", conf.BBSim.BandwidthProfileFormat, "Bandwidth profile format, 'mef' or 'ietf'")

	olt_id := flag.Int("olt_id", conf.Olt.ID, "OLT device ID")
	nni := flag.Int("nni", int(conf.Olt.NniPorts), "Number of NNI ports per OLT device to be emulated")
	nni_speed := flag.Uint("nni_speed", uint(conf.Olt.NniSpeed), "Reported speed of the NNI ports in Mbps")
	pon := flag.Int("pon", int(conf.Olt.PonPorts), "Number of PON ports per OLT device to be emulated")
	onu := flag.Int("onu", int(conf.Olt.OnusPonPort), "Number of ONU devices per PON port to be emulated")
	uni := flag.Int("uni", int(conf.Olt.UniPorts), "Number of Ethernet UNI Ports per ONU device to be emulated")
	pots := flag.Int("pots", int(conf.Olt.PotsPorts), "Number of POTS UNI Ports per ONU device to be emulated")
	NniDchpTrapVid := flag.Int("nni_dhcp_trap_vid", int(conf.Olt.NniDhcpTrapVid), "Vlan to trap the DHCP packets on")

	oltRebootDelay := flag.Int("oltRebootDelay", conf.Olt.OltRebootDelay, "Time that BBSim should before restarting after a reboot")
	omci_response_rate := flag.Int("omci_response_rate", int(conf.Olt.OmciResponseRate), "Amount of OMCI messages to respond to")

	openolt_address := flag.String("openolt_address", conf.BBSim.OpenOltAddress, "IP address:port")
	api_address := flag.String("api_address", conf.BBSim.ApiAddress, "IP address:port")
	rest_api_address := flag.String("rest_api_address", conf.BBSim.RestApiAddress, "IP address:port")
	dmi_server_address := flag.String("dmi_server_address", conf.BBSim.DmiServerAddress, "IP address:port")

	profileCpu := flag.String("cpuprofile", "", "write cpu profile to file")

	logLevel := flag.String("logLevel", conf.BBSim.LogLevel, "Set the log level (trace, debug, info, warn, error)")
	logCaller := flag.Bool("logCaller", conf.BBSim.LogCaller, "Whether to print the caller filename or not")

	delay := flag.Int("delay", conf.BBSim.Delay, "The delay between ONU DISCOVERY batches in milliseconds (1 ONU per each PON PORT at a time")

	controlledActivation := flag.String("ca", conf.BBSim.ControlledActivation, "Set the mode for controlled activation of PON ports and ONUs")
	enablePerf := flag.Bool("enableperf", conf.BBSim.EnablePerf, "Setting this flag will cause BBSim to not store data like traffic schedulers, flows of ONUs etc..")
	enableEvents := flag.Bool("enableEvents", conf.BBSim.Events, "Enable sending BBSim events on configured kafka server")
	kafkaAddress := flag.String("kafkaAddress", conf.BBSim.KafkaAddress, "IP:Port for kafka")
	kafkaEventTopic := flag.String("kafkaEventTopic", conf.BBSim.KafkaEventTopic, "Ability to configure the topic on which BBSim publishes events on Kafka")
	dhcpRetry := flag.Bool("dhcpRetry", conf.BBSim.DhcpRetry, "Set this flag if BBSim should retry DHCP upon failure until success")
	authRetry := flag.Bool("authRetry", conf.BBSim.AuthRetry, "Set this flag if BBSim should retry EAPOL (Authentication) upon failure until success")
	injectOmciUnknownMe := flag.Bool("injectOmciUnknownMe", conf.BBSim.InjectOmciUnknownMe, "Generate an extra MibDB packet with ClassID 37 (Intentionally left blank)")
	injectOmciUnknownAttributes := flag.Bool("injectOmciUnknownAttributes", conf.BBSim.InjectOmciUnknownAttributes, "Modifies the ONU2-G MibDB packet to add Unknown Attributes")
	omccVersion := flag.Int("omccVersion", conf.BBSim.OmccVersion, "Set OMCC version to be returned in OMCI response of ME Onu2G")

	flag.Parse()

	conf.Olt.ID = int(*olt_id)
	conf.Olt.NniPorts = uint32(*nni)
	conf.Olt.NniSpeed = uint32(*nni_speed)
	conf.Olt.PonPorts = uint32(*pon)
	conf.Olt.UniPorts = uint32(*uni)
	conf.Olt.PotsPorts = uint32(*pots)
	conf.Olt.OnusPonPort = uint32(*onu)
	conf.Olt.NniDhcpTrapVid = uint32(*NniDchpTrapVid)
	conf.Olt.OltRebootDelay = *oltRebootDelay
	conf.Olt.OmciResponseRate = uint8(*omci_response_rate)
	conf.BBSim.ConfigFile = *configFile
	conf.BBSim.ServiceConfigFile = *servicesFile
	conf.BBSim.PonsConfigFile = *ponsFile
	conf.BBSim.CpuProfile = profileCpu
	conf.BBSim.LogLevel = *logLevel
	conf.BBSim.LogCaller = *logCaller
	conf.BBSim.Delay = *delay
	conf.BBSim.ControlledActivation = *controlledActivation
	conf.BBSim.EnablePerf = *enablePerf
	conf.BBSim.Events = *enableEvents
	conf.BBSim.KafkaAddress = *kafkaAddress
	conf.BBSim.OpenOltAddress = *openolt_address
	conf.BBSim.ApiAddress = *api_address
	conf.BBSim.RestApiAddress = *rest_api_address
	conf.BBSim.KafkaEventTopic = *kafkaEventTopic
	conf.BBSim.AuthRetry = *authRetry
	conf.BBSim.DhcpRetry = *dhcpRetry
	conf.BBSim.DmiServerAddress = *dmi_server_address
	conf.BBSim.InjectOmciUnknownMe = *injectOmciUnknownMe
	conf.BBSim.InjectOmciUnknownAttributes = *injectOmciUnknownAttributes
	conf.BBSim.OmccVersion = *omccVersion

	// update device id if not set
	if conf.Olt.DeviceId == "" {
		conf.Olt.DeviceId = net.HardwareAddr{0xA, 0xA, 0xA, 0xA, 0xA, byte(conf.Olt.ID)}.String()
	}

	// check that the BP format is valid
	if (*sadisBpFormat != BP_FORMAT_MEF) && (*sadisBpFormat != BP_FORMAT_IETF) {
		log.Fatalf("Invalid parameter 'bp_format', supported values are %s and %s, you provided %s", BP_FORMAT_MEF, BP_FORMAT_IETF, *sadisBpFormat)
	}
	conf.BBSim.BandwidthProfileFormat = *sadisBpFormat

	return conf
}

func GetDefaultOps() *GlobalConfig {

	c := &GlobalConfig{
		BBSimConfig{
			ConfigFile:        "configs/bbsim.yaml",
			ServiceConfigFile: "configs/att-services.yaml",
			// PonsConfigFile is left intentionally blank here
			// to use the default values computed at runtime depending
			// on the loaded Services
			PonsConfigFile:              "",
			LogLevel:                    "debug",
			LogCaller:                   false,
			Delay:                       200,
			OpenOltAddress:              ":50060",
			ApiAddress:                  ":50070",
			RestApiAddress:              ":50071",
			LegacyApiAddress:            ":50072",
			LegacyRestApiAddress:        ":50073",
			SadisRestAddress:            ":50074",
			SadisServer:                 true,
			KafkaAddress:                ":9092",
			Events:                      false,
			ControlledActivation:        "default",
			EnablePerf:                  false,
			KafkaEventTopic:             "",
			DhcpRetry:                   false,
			AuthRetry:                   false,
			DmiServerAddress:            ":50075",
			BandwidthProfileFormat:      BP_FORMAT_MEF,
			InjectOmciUnknownMe:         false,
			InjectOmciUnknownAttributes: false,
			OmccVersion:                 0xA3,
		},
		OltConfig{
			Vendor:             "BBSim",
			Model:              "asfvolt16",
			HardwareVersion:    "emulated",
			FirmwareVersion:    "",
			DeviceSerialNumber: "BBSM00000001",
			PonPorts:           1,
			NniPorts:           1,
			NniSpeed:           10000, //Mbps
			OnusPonPort:        1,
			ID:                 0,
			OltRebootDelay:     60,
			PortStatsInterval:  20,
			OmciResponseRate:   10,
			UniPorts:           4,
			PotsPorts:          0,
			NniDhcpTrapVid:     0,
		},
		BBRConfig{
			LogLevel:  "debug",
			LogCaller: false,
		},
	}
	return c
}

func getDefaultPonsConfig() (*PonPortsConfig, error) {

	if Config == nil {
		return nil, fmt.Errorf("Config is nil")
	}
	if Services == nil {
		return nil, fmt.Errorf("Services is nil")
	}

	//The default should replicate the old way bbsim used to compute resource ranges based on the configuration
	// 1 allocId per Service * UNI
	allocIdPerOnu := uint32(Config.Olt.UniPorts * uint32(len(Services)))
	return &PonPortsConfig{
		Number: Config.Olt.PonPorts,
		Ranges: []PonRangeConfig{
			{
				PonRange:   IdRange{0, Config.Olt.PonPorts - 1},
				Technology: XGSPON.String(),
				// we need one ONU ID available per ONU, but the smaller the range the smaller the pool created in the openolt adapter
				OnuRange: IdRange{defaultOnuIdStart, defaultOnuIdStart + (Config.Olt.OnusPonPort - 1)},
				// 1 allocId per Service * UNI * ONU
				AllocIdRange: IdRange{defaultAllocIdStart, defaultAllocIdStart + (Config.Olt.OnusPonPort * allocIdPerOnu)},
				// up to 8 gemport-id per tcont/alloc-id
				GemportRange: IdRange{defaultGemportIdStart, defaultGemportIdStart + Config.Olt.OnusPonPort*allocIdPerOnu*defaultGemPortIdPerAllocId},
			},
		},
	}, nil
}

// LoadBBSimConf loads the BBSim configuration from a YAML file
func loadBBSimConf(filename string) (*GlobalConfig, error) {
	yamlConfig := GetDefaultOps()

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"filename": filename,
		}).Error("Cannot load BBSim configuration file. Using defaults.")
		return yamlConfig, nil
	}

	err = yaml.Unmarshal(yamlFile, yamlConfig)
	if err != nil {
		return nil, err
	}

	return yamlConfig, nil
}

// loadBBSimPons loads the configuration of PON interfaces from a YAML file
func loadBBSimPons(filename string) (*PonPortsConfig, error) {
	yamlPonsConfig, err := getDefaultPonsConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Can't load Pon interfaces defaults.")
		return nil, err
	}

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"filename": filename,
		}).Error("Cannot load Pon interfaces configuration file. Using defaults.")
		return yamlPonsConfig, nil
	}

	err = yaml.Unmarshal(yamlFile, yamlPonsConfig)
	if err != nil {
		return nil, err
	}

	return yamlPonsConfig, nil
}

// validatePonsConfig checks if the configuration to use for the definition of Pon interfaces is valid
func validatePonsConfig(pons *PonPortsConfig) error {

	if pons.Number == 0 {
		return fmt.Errorf("no-pon-ports")
	}

	definedPorts := make([]int, pons.Number)

	for rIndex, resRange := range pons.Ranges {
		if _, err := PonTechnologyFromString(resRange.Technology); err != nil {
			return err
		}

		if resRange.PonRange.EndId < resRange.PonRange.StartId {
			return fmt.Errorf("invalid-pon-ports-limits-in-range-%d", rIndex)
		}

		//Keep track of the defined pons
		for p := resRange.PonRange.StartId; p <= resRange.PonRange.EndId; p++ {
			if p > uint32(len(definedPorts)-1) {
				return fmt.Errorf("pon-port-%d-in-range-%d-but-max-is-%d", p, rIndex, pons.Number-1)
			}
			definedPorts[p]++

			if definedPorts[p] > 1 {
				return fmt.Errorf("pon-port-%d-has-duplicate-definition-in-range-%d", p, rIndex)
			}
		}

		if resRange.OnuRange.EndId < resRange.OnuRange.StartId {
			return fmt.Errorf("invalid-onus-limits-in-range-%d", rIndex)
		}
		if resRange.AllocIdRange.EndId < resRange.AllocIdRange.StartId {
			return fmt.Errorf("invalid-allocid-limits-in-range-%d", rIndex)
		}
		if resRange.GemportRange.EndId < resRange.GemportRange.StartId {
			return fmt.Errorf("invalid-gemport-limits-in-range-%d", rIndex)
		}
	}

	//Check if the ranges define all the pons
	for i, num := range definedPorts {
		if num < 1 {
			return fmt.Errorf("pon-port-%d-is-not-defined-in-ranges", i)
		}
	}

	return nil
}

// LoadBBSimServices parses a file describing the services that need to be created for each UNI
func loadBBSimServices(filename string) ([]ServiceYaml, error) {

	yamlServiceCfg := YamlServiceConfig{}

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(yamlFile), &yamlServiceCfg)
	if err != nil {
		return nil, err
	}

	for _, service := range yamlServiceCfg.Services {

		if service.CTagAllocation == "" || service.STagAllocation == "" {
			log.Fatal("c_tag_allocation and s_tag_allocation are mandatory fields")
		}

		if _, err := tagAllocationFromString(string(service.CTagAllocation)); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatal("c_tag_allocation is not valid")
		}
	}

	log.WithFields(log.Fields{
		"services": yamlServiceCfg.String(),
	}).Debug("BBSim services description correctly loaded")
	return yamlServiceCfg.Services, nil
}

// This is only used by BBR
func GetBBROpts() BBRCliOptions {

	bbsimIp := flag.String("bbsimIp", "127.0.0.1", "BBSim IP")
	bbsimPort := flag.String("bbsimPort", "50060", "BBSim Port")
	bbsimApiPort := flag.String("bbsimApiPort", "50070", "BBSim API Port")
	logFile := flag.String("logfile", "", "Log to a file")

	LoadConfig()

	bbrOptions := BBRCliOptions{
		Config,
		*bbsimIp,
		*bbsimPort,
		*bbsimApiPort,
		*logFile,
	}

	return bbrOptions
}
