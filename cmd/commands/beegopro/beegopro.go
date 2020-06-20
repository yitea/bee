// Copyright 2013 bee authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.
package beegopro

import (
	"strings"

	"github.com/beego/bee/cmd/commands"
	"github.com/beego/bee/cmd/commands/version"
	"github.com/beego/bee/logger"
	"github.com/beego/bee/pkg/beegopro"
)

var CmdBeegoPro = &commands.Command{
	UsageLine: "beegopro [command]",
	Short:     "Source code generator",
	Long:      ``,
	PreRun:    func(cmd *commands.Command, args []string) { version.ShowShortVersionBanner() },
	Run:       BeegoPro,
}

func init() {
	commands.AvailableCommands = append(commands.AvailableCommands, CmdBeegoPro)
}

func BeegoPro(cmd *commands.Command, args []string) int {
	if len(args) < 1 {
		beeLogger.Log.Fatal("Command is missing")
	}

	gcmd := args[0]
	switch gcmd {
	case "generate":
		beegopro.DefaultBeegoPro.Generate(true)
	case "config":
	default:
		beeLogger.Log.Fatal("Command is missing")
	}
	beeLogger.Log.Successf("%s successfully generated!", strings.Title(gcmd))
	return 0
}
