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
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	"gopkg.in/yaml.v2"
)

const copyrightNotice = `
# Portions copyright 2019-present Open Networking Foundation
# Original copyright 2019-present Ciena Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
`

type ConfigOptions struct{}

func RegisterConfigCommands(parent *flags.Parser) {
	_, _ = parent.AddCommand("config", "generate bbsimctl configuration", "Commands to generate bbsimctl configuration", &ConfigOptions{})
}

func (options *ConfigOptions) Execute(args []string) error {
	//GlobalConfig
	config.ProcessGlobalOptions()

	b, err := yaml.Marshal(config.GlobalConfig)
	if err != nil {
		return err
	}
	fmt.Println(copyrightNotice)
	fmt.Println(string(b))

	fmt.Println("BBSimCtl details:")
	fmt.Printf("\tVersion: %s \n", config.Version)
	fmt.Printf("\tBuildTime: %s \n", config.BuildTime)
	fmt.Printf("\tCommitHash: %s \n", config.CommitHash)
	fmt.Printf("\tGitStatus: %s \n", config.GitStatus)
	return nil
}
