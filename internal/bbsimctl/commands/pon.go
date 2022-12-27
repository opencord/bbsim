/*
 * Portions Copyright 2019-2023 Open Networking Foundation (ONF) and the ONF Contributors
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

type PonPoweronAllOnus struct {
	Args struct {
		IntfID OltInterfaceID
	} `positional-args:"yes" required:"yes"`
}

type PonShutdownAllOnus struct {
	Args struct {
		IntfID OltInterfaceID
	} `positional-args:"yes" required:"yes"`
}

type PONOptions struct {
	PoweronAllOnus  PonPoweronAllOnus  `command:"poweronAllONUs"`
	ShutdownAllOnus PonShutdownAllOnus `command:"shutdownAllONUs"`
}

func RegisterPonCommands(parser *flags.Parser) {
	_, _ = parser.AddCommand("pon", "PON Commands", "Commands to query and manipulate the PON port", &PONOptions{})
}

func (pon *PonPoweronAllOnus) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.PONRequest{
		PonPortId: uint32(pon.Args.IntfID),
	}

	res, err := client.PoweronONUsOnPON(ctx, &req)
	if err != nil {
		log.Errorf("Cannot poweron all ONUs on PON port: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

func (pon *PonShutdownAllOnus) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.PONRequest{
		PonPortId: uint32(pon.Args.IntfID),
	}

	res, err := client.ShutdownONUsOnPON(ctx, &req)
	if err != nil {
		log.Errorf("Cannot shutdown all ONUs on PON port: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}
