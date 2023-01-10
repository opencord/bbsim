/*
 * Copyright 2019-2023 Open Networking Foundation (ONF) and the ONF Contributors
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
	"github.com/opencord/bbsim/internal/common"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/olekukonko/tablewriter"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	log "github.com/sirupsen/logrus"
)

const (
	DEFAULT_OLT_ALARM_LIST_FORMAT = "table{{ .Name }}"
)

type OltInterfaceStatus string

type OltInterfaceID uint32

type OltAlarmNameString string

type OltAlarmListOutput struct {
	Name string
}

type OltAlarmRaise struct {
	Args struct {
		Name   OltAlarmNameString
		IntfID OltInterfaceID
	} `positional-args:"yes" required:"yes"`
}

type OltAlarmClear struct {
	Args struct {
		Name   OltAlarmNameString
		IntfID OltInterfaceID
	} `positional-args:"yes" required:"yes"`
}

type OltAlarmList struct{}

type OltAlarmOptions struct {
	Raise OltAlarmRaise `command:"raise"`
	Clear OltAlarmClear `command:"clear"`
	List  OltAlarmList  `command:"list"`
}

// Execute alarm raise
func (o *OltAlarmRaise) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.OLTAlarmRequest{
		InterfaceID: uint32(o.Args.IntfID),
		Status:      "on"}

	if string(o.Args.Name) == common.OltPonLos {
		req.InterfaceType = "pon"
	} else if string(o.Args.Name) == common.OltNniLos {
		req.InterfaceType = "nni"
	} else {
		return fmt.Errorf("Unknown OLT alarm type")
	}

	res, err := client.SetOltAlarmIndication(ctx, &req)
	if err != nil {
		log.Fatalf("Cannot raise OLT alarm: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

// Execute alarm clear
func (o *OltAlarmClear) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.OLTAlarmRequest{
		InterfaceID: uint32(o.Args.IntfID),
		Status:      "off"}

	if string(o.Args.Name) == common.OltPonLos {
		req.InterfaceType = "pon"
	} else if string(o.Args.Name) == common.OltNniLos {
		req.InterfaceType = "nni"
	} else {
		return fmt.Errorf("Unknown OLT alarm type")
	}

	res, err := client.SetOltAlarmIndication(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot clear OLT alarm: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

// Execute OLT alarm list
func (o *OltAlarmList) Execute(args []string) error {
	OltAlarmsValue := [][]string{}
	OltAlarmstable := tablewriter.NewWriter(os.Stdout)
	fmt.Fprintf(os.Stdout, "OLT Alarms List:\n")
	OltAlarmstable.SetHeader([]string{"OLT Alarms"})

	alarmNames := make([]AlarmListOutput, len(common.OLTAlarms))
	i := 0
	for k := range common.OLTAlarms {
		alarmNames[i] = AlarmListOutput{Name: k}
		OltAlarmsValue = append(OltAlarmsValue, []string{k})
		i++
	}
	for _, v := range OltAlarmsValue {
		OltAlarmstable.Append(v)
	}
	OltAlarmstable.Render()
	return nil
}

func (o *OltAlarmNameString) Complete(match string) []flags.Completion {
	list := make([]flags.Completion, 0)
	for k := range common.OLTAlarms {
		if strings.HasPrefix(k, match) {
			list = append(list, flags.Completion{Item: k})
		}
	}

	return list
}
