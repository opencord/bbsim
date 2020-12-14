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

package commands

import (
	"context"
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type DMIOptions struct {
	Events DmiEventOptions `command:"events"`
}

type DmiEventCreate struct {
	Args struct {
		Name string
	} `positional-args:"yes" required:"yes"`
}

type DmiEventOptions struct {
	Create DmiEventCreate `command:"create"`
}

func RegisterDMICommands(parser *flags.Parser) {
	_, _ = parser.AddCommand("dmi", "DMI Commands", "Commands to create events", &DMIOptions{})
}

func dmiEventGrpcClient() (bbsim.BBsimDmiClient, *grpc.ClientConn) {
	conn, err := grpc.Dial(config.DmiConfig.Server, grpc.WithInsecure())
	if err != nil {
		log.Errorf("BBsimDmiClient connection failed  : %v", err)
		return nil, conn
	}
	return bbsim.NewBBsimDmiClient(conn), conn
}

// Execute create event
func (o *DmiEventCreate) Execute(args []string) error {
	client, conn := dmiEventGrpcClient()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := bbsim.DmiEvent{EventName: o.Args.Name}
	res, err := client.CreateEvent(ctx, &req)
	if err != nil {
		log.Errorf("Cannot create DMI event: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}
