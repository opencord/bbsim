/*
 * Portions copyright 2019-present Open Networking Foundation
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

package commands

import (
	"context"
	"fmt"
	"github.com/jessevdk/go-flags"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	"github.com/opencord/cordctl/pkg/format"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (
	DEFAULT_ALARM_LIST_FORMAT = "table{{ .Name }}"
)

type AlarmListOutput struct {
	Name string
}

type AlarmRaise struct {
	Parameters []string `short:"p" description:"Additional Alarm Parameter in name=value form"`
	Args       struct {
		Name         string
		SerialNumber string
	} `positional-args:"yes" required:"yes"`
}

type AlarmClear struct {
	Parameters []string `short:"p" description:"Additional Alarm Parameter in name=value form"`
	Args       struct {
		Name         string
		SerialNumber string
	} `positional-args:"yes" required:"yes"`
}

type AlarmList struct{}

type AlarmOptions struct {
	Raise AlarmRaise `command:"raise"`
	Clear AlarmClear `command:"clear"`
	List  AlarmList  `command:"list"`
}

var AlarmNameMap = map[string]pb.AlarmType_Types{"DyingGasp": pb.AlarmType_DYING_GASP,
	"StartupFailure":           pb.AlarmType_ONU_STARTUP_FAILURE,
	"SignalDegrade":            pb.AlarmType_ONU_SIGNAL_DEGRADE,
	"DriftOfWindow":            pb.AlarmType_ONU_DRIFT_OF_WINDOW,
	"LossOfOmciChannel":        pb.AlarmType_ONU_LOSS_OF_OMCI_CHANNEL,
	"SignalsFailure":           pb.AlarmType_ONU_SIGNALS_FAILURE,
	"TransmissionInterference": pb.AlarmType_ONU_TRANSMISSION_INTERFERENCE_WARNING,
	"ActivationFailure":        pb.AlarmType_ONU_ACTIVATION_FAILURE,
	"ProcessingError":          pb.AlarmType_ONU_PROCESSING_ERROR,
	"LossOfKeySyncFailure":     pb.AlarmType_ONU_LOSS_OF_KEY_SYNC_FAILURE,

	// Break out OnuAlarm into its subcases.
	"LossOfSignal":   pb.AlarmType_ONU_ALARM_LOS,
	"LossOfBurst":    pb.AlarmType_ONU_ALARM_LOB,
	"LOPC_MISS":      pb.AlarmType_ONU_ALARM_LOPC_MISS,
	"LOPC_MIC_ERROR": pb.AlarmType_ONU_ALARM_LOPC_MIC_ERROR,
	"LossOfFrame":    pb.AlarmType_ONU_ALARM_LOFI,
	"LossOfPloam":    pb.AlarmType_ONU_ALARM_LOAMI,

	// Whole-PON / Non-onu-specific
	"PonLossOfSignal": pb.AlarmType_LOS,
}

func alarmNameToEnum(name string) (*pb.AlarmType_Types, error) {
	v, okay := AlarmNameMap[name]
	if !okay {
		return nil, fmt.Errorf("Unknown Alarm Name: %v", name)
	}

	return &v, nil
}

// add optional parameters from the command-line to the AlarmRequest
func addParameters(parameters []string, req *pb.AlarmRequest) error {
	req.Parameters = make([]*pb.AlarmParameter, len(parameters))
	for i, kv := range parameters {
		parts := strings.Split(kv, "=")
		if len(parts) != 2 {
			return fmt.Errorf("Invalid parameter %v", kv)
		}
		req.Parameters[i] = &pb.AlarmParameter{Key: parts[0], Value: parts[1]}
	}
	return nil
}

func RegisterAlarmCommands(parser *flags.Parser) {
	parser.AddCommand("alarm", "Alarm Commands", "Commands to raise and clear alarms", &AlarmOptions{})
}

func (o *AlarmRaise) Execute(args []string) error {
	alarmType, err := alarmNameToEnum(o.Args.Name)
	if err != nil {
		return err
	}

	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.AlarmRequest{AlarmType: *alarmType,
		SerialNumber: o.Args.SerialNumber,
		Status:       "on"}

	err = addParameters(o.Parameters, &req)
	if err != nil {
		return err
	}

	res, err := client.SetAlarmIndication(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot raise alarm: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

func (o *AlarmClear) Execute(args []string) error {
	alarmType, err := alarmNameToEnum(o.Args.Name)
	if err != nil {
		return err
	}

	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.AlarmRequest{AlarmType: *alarmType,
		SerialNumber: o.Args.SerialNumber,
		Status:       "off"}

	err = addParameters(o.Parameters, &req)
	if err != nil {
		return err
	}

	res, err := client.SetAlarmIndication(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot clear alarm: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

func (o *AlarmList) Execute(args []string) error {
	alarmNames := make([]AlarmListOutput, len(AlarmNameMap))
	i := 0
	for k := range AlarmNameMap {
		alarmNames[i] = AlarmListOutput{Name: k}
		i++
	}
	// print out
	tableFormat := format.Format(DEFAULT_ALARM_LIST_FORMAT)
	tableFormat.Execute(os.Stdout, true, alarmNames)
	return nil
}
