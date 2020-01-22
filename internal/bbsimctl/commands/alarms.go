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
	log "github.com/sirupsen/logrus"
)

type AlarmRaise struct {
	Args struct {
		Name         string
		SerialNumber string
	} `positional-args:"yes" required:"yes"`
}

type AlarmClear struct {
	Args struct {
		Name         string
		SerialNumber string
	} `positional-args:"yes" required:"yes"`
}

type AlarmOptions struct {
	Raise AlarmRaise `command:"raise"`
	Clear AlarmClear `command:"clear"`
}

func RegisterAlarmCommands(parser *flags.Parser) {
	parser.AddCommand("alarm", "Alarm Commands", "Commands to raise and clear alarms", &AlarmOptions{})
}

func (o *AlarmRaise) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.ONUAlarmRequest{AlarmName: o.Args.Name,
		SerialNumber: o.Args.SerialNumber,
		Status:       "on"}

	res, err := client.SetONUAlarmIndication(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot raise alarm: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

func (o *AlarmClear) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.ONUAlarmRequest{AlarmName: o.Args.Name,
		SerialNumber: o.Args.SerialNumber,
		Status:       "off"}

	res, err := client.SetONUAlarmIndication(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot clear alarm: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}
