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

import "flag"

type BBSimCliOptions struct {
	OltID        int
	NumNniPerOlt int
	NumPonPerOlt int
	NumOnuPerPon int
	STag         int
	CTagInit     int
	Auth         bool
	Dhcp         bool
	ProfileCpu   *string
	LogLevel     string
	LogCaller    bool
	Delay        int
}

type BBRCliOptions struct {
	*BBSimCliOptions
	BBSimIp      string
	BBSimPort    string
	BBSimApiPort string
	LogFile      string
}

func GetBBSimOpts() *BBSimCliOptions {

	olt_id := flag.Int("olt_id", 0, "Number of OLT devices to be emulated")
	nni := flag.Int("nni", 1, "Number of NNI ports per OLT device to be emulated")
	pon := flag.Int("pon", 1, "Number of PON ports per OLT device to be emulated")
	onu := flag.Int("onu", 1, "Number of ONU devices per PON port to be emulated")

	auth := flag.Bool("auth", false, "Set this flag if you want authentication to start automatically")
	dhcp := flag.Bool("dhcp", false, "Set this flag if you want DHCP to start automatically")

	s_tag := flag.Int("s_tag", 900, "S-Tag value")
	c_tag_init := flag.Int("c_tag", 900, "C-Tag starting value, each ONU will get a sequential one (targeting 1024 ONUs per BBSim instance the range is big enough)")

	profileCpu := flag.String("cpuprofile", "", "write cpu profile to file")

	logLevel := flag.String("logLevel", "debug", "Set the log level (trace, debug, info, warn, error)")
	logCaller := flag.Bool("logCaller", false, "Whether to print the caller filename or not")

	delay := flag.Int("delay", 200, "The delay between ONU DISCOVERY batches in milliseconds (1 ONU per each PON PORT at a time")

	flag.Parse()

	o := new(BBSimCliOptions)

	o.OltID = int(*olt_id)
	o.NumNniPerOlt = int(*nni)
	o.NumPonPerOlt = int(*pon)
	o.NumOnuPerPon = int(*onu)
	o.STag = int(*s_tag)
	o.CTagInit = int(*c_tag_init)
	o.ProfileCpu = profileCpu
	o.LogLevel = *logLevel
	o.LogCaller = *logCaller
	o.Auth = *auth
	o.Dhcp = *dhcp
	o.Delay = *delay

	return o
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
