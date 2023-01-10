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

package commands

import (
	"context"
	"fmt"

	"github.com/jessevdk/go-flags"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type LoggingOptions struct {
	Args struct {
		Level  string
		Caller bool
	} `positional-args:"yes" required:"yes"`
}

func RegisterLoggingCommands(parent *flags.Parser) {
	_, _ = parent.AddCommand("log", "set bbsim log level", "Commands to set the log level", &LoggingOptions{})
}

func (options *LoggingOptions) Execute(args []string) error {
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

	req := pb.LogLevel{
		Level:  options.Args.Level,
		Caller: options.Args.Caller,
	}

	logLevel, err := c.SetLogLevel(ctx, &req)

	if err != nil {
		log.Fatalf("could not set log level: %v", err)
	}

	fmt.Println("New log settings:")
	fmt.Println(fmt.Sprintf("\tLevel: %s", logLevel.Level))
	fmt.Println(fmt.Sprintf("\tReportCaller: %t", logLevel.Caller))

	return nil
}
