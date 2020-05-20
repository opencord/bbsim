// Copyright 2019-present Ciena Corporation
// SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/opencord/bbsim/internal/bbsimctl/completion"
)

type BashOptions struct{}

type CompletionOptions struct {
	BashOptions `command:"bash"`
}

func RegisterCompletionCommands(parent *flags.Parser) {
	parent.AddCommand("completion", "generate shell compleition", "Commands to generate shell completion information", &CompletionOptions{})
}

func (options *BashOptions) Execute(args []string) error {
	fmt.Println(completion.Bash)
	return nil
}
