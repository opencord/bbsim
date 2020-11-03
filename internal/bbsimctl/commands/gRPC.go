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
	"github.com/jessevdk/go-flags"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	"google.golang.org/grpc"
	"log"
)

type closeChannelgRPCOptions struct{}

type startChannelgRPCOptions struct{}

type gRPCOptions struct {
	Start             OltGet            `command:"start"`
	Stop             OltNNIs            `command:"stop"`
}

func RegistergRPCCommands(parent *flags.Parser) {
	_, _ = parent.AddCommand("grpc", "manages bbsim gRPC API server", "Command to manage the gRPC API of BBSIM, " +
		"starting or stopping it to simulating a loss of gRPC connection to the OLT agent", &gRPCOptions{})
}

func (opts *closeChannelgRPCOptions) Execute(args []string) error {
	conn, err := grpc.Dial(config.GlobalConfig.Server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil
	}
	defer conn.Close()
	c := pb.NewBBSimClient(conn)

	// Contact the server and print out its response.

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()
	err := c.StopgRPCChannel(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not stop gRPC channel, %v", err)
		return err
	}
	return nil
}

func (opts *startChannelgRPCOptions) Execute(args []string) error {
	conn, err := grpc.Dial(config.GlobalConfig.Server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil
	}
	defer conn.Close()
	c := pb.NewBBSimClient(conn)

	// Contact the server and print out its response.

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()
	err := c.StartgRPCChannel(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not stop gRPC channel, %v", err)
		return err
	}
	return nil
}



