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

	"github.com/jessevdk/go-flags"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/alarmsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	"github.com/opencord/cordctl/pkg/format"
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

	req := pb.OLTAlarmRequest{InterfaceType: string(o.Args.Name),
		InterfaceID: uint32(o.Args.IntfID),
		Status:      "on"}

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

	req := pb.OLTAlarmRequest{InterfaceType: string(o.Args.Name),
		InterfaceID: uint32(o.Args.IntfID),
		Status:      "off"}

	res, err := client.SetOltAlarmIndication(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot clear OLT alarm: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

// Execute alarm list
func (o *OltAlarmList) Execute(args []string) error {
	alarmNames := make([]AlarmListOutput, len(alarmsim.OltAlarmNameMap))
	i := 0
	for k := range alarmsim.OltAlarmNameMap {
		alarmNames[i] = AlarmListOutput{Name: k}
		i++
	}
	// print out
	tableFormat := format.Format(DEFAULT_OLT_ALARM_LIST_FORMAT)
	tableFormat.Execute(os.Stdout, true, alarmNames)
	return nil
}

func (o *OltAlarmNameString) Complete(match string) []flags.Completion {
	list := make([]flags.Completion, 0)
	for k := range alarmsim.OltAlarmNameMap {
		if strings.HasPrefix(k, match) {
			list = append(list, flags.Completion{Item: k})
		}
	}

	return list
}
