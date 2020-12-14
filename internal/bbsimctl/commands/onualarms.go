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
	"os"
	"strings"

	"github.com/opencord/bbsim/internal/common"

	"github.com/jessevdk/go-flags"
	"github.com/olekukonko/tablewriter"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	log "github.com/sirupsen/logrus"
)

type AlarmNameString string

type AlarmListOutput struct {
	Name string
}

type AlarmRaise struct {
	Parameters []string `short:"p" description:"Additional Alarm Parameter in name=value form"`
	Args       struct {
		Name         AlarmNameString
		SerialNumber OnuSnString
	} `positional-args:"yes" required:"yes"`
}

type AlarmClear struct {
	Parameters []string `short:"p" description:"Additional Alarm Parameter in name=value form"`
	Args       struct {
		Name         AlarmNameString
		SerialNumber OnuSnString
	} `positional-args:"yes" required:"yes"`
}

type AlarmList struct{}

type AlarmOptions struct {
	Raise AlarmRaise `command:"raise"`
	Clear AlarmClear `command:"clear"`
	List  AlarmList  `command:"list"`
}

// add optional parameters from the command-line to the ONUAlarmRequest
func addParameters(parameters []string, req *pb.ONUAlarmRequest) error {
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

// Execute alarm raise
func (o *AlarmRaise) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.ONUAlarmRequest{AlarmType: string(o.Args.Name),
		SerialNumber: string(o.Args.SerialNumber),
		Status:       "on"}

	err := addParameters(o.Parameters, &req)
	if err != nil {
		return err
	}

	res, err := client.SetOnuAlarmIndication(ctx, &req)
	if err != nil {
		log.Fatalf("Cannot raise alarm: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

// Execute alarm clear
func (o *AlarmClear) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.ONUAlarmRequest{AlarmType: string(o.Args.Name),
		SerialNumber: string(o.Args.SerialNumber),
		Status:       "off"}

	err := addParameters(o.Parameters, &req)
	if err != nil {
		return err
	}

	res, err := client.SetOnuAlarmIndication(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot clear alarm: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

// Execute ONU alarm list
func (o *AlarmList) Execute(args []string) error {
	OnuAlarmsValue := [][]string{}
	OnuAlarmstable := tablewriter.NewWriter(os.Stdout)
	fmt.Fprintf(os.Stdout, "ONU Alarms List:\n")
	OnuAlarmstable.SetHeader([]string{"ONU Alarms"})

	alarmNames := make([]AlarmListOutput, len(common.ONUAlarms))
	i := 0
	for k := range common.ONUAlarms {
		alarmNames[i] = AlarmListOutput{Name: k}
		OnuAlarmsValue = append(OnuAlarmsValue, []string{k})
		i++
	}
	for _, v := range OnuAlarmsValue {
		OnuAlarmstable.Append(v)
	}
	OnuAlarmstable.Render()
	return nil
}

func (onuSn *AlarmNameString) Complete(match string) []flags.Completion {
	list := make([]flags.Completion, 0)
	for k := range common.ONUAlarms {
		if strings.HasPrefix(k, match) {
			list = append(list, flags.Completion{Item: k})
		}
	}

	return list
}
