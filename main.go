package main

import (
	"fmt"
	"os"

	"github.com/EngineerBetter/control-tower/commands"
	"github.com/fatih/color"

	"github.com/urfave/cli"
)

// ControlTowerVersion is a compile-time variable set with -ldflags
var ControlTowerVersion = "COMPILE_TIME_VARIABLE_main_ControlTowerVersion"
var blue = color.New(color.FgCyan, color.Bold).SprintfFunc()

func main() {
	app := cli.NewApp()
	app.Name = "Control-Tower"
	app.Usage = "A CLI tool to deploy Concourse CI"
	app.Version = ControlTowerVersion
	app.Commands = commands.Commands
	app.Flags = commands.GlobalFlags
	cli.AppHelpTemplate = fmt.Sprintf(`%s

See 'control-tower help <command>' to read about a specific command.

Built by %s %s

`, cli.AppHelpTemplate, blue("EngineerBetter"), blue("http://engineerbetter.com"))

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
