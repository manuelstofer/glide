package main

import (
	"github.com/Masterminds/glide/cmd"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/cookoo/cli"

	"flag"
	"fmt"
	"os"
)

var version string = "DEV"

const Summary = "Manage Go projects with ease."
const Usage = `Manage dependencies, naming, and GOPATH for your Go projects.

Examples:
	$ glide init
	$ glide in
	$ glide install
	$ glide update
	$ glide rebuild

COMMANDS
========

Utilities:

- help: Show this help message (alias of -h)
- status: Print a status report.
- version: Print the version and exit.

Dependency management:

- init: Initialize a new project, creating a template glide.yaml.
- install: Install all packages in the glide.yaml.
- update: Update existing packages (alias: 'up').
- rebuild: Rebuild ('go build') the dependencies.

Project tools:

- in: Glide into a commandline shell preconfigured for your project (with
  GOPATH set).
- into: "glide into /my/project" is the same as running "cd /my/project && glide in"
- gopath: Emits the GOPATH for the current project. Useful for things like
  manually setting GOPATH: GOPATH=$(glide gopath)

FILES
=====

Each project should have a 'glide.yaml' file in the project directory. Files
look something like this:

	package: github.com/Masterminds/glide
	imports:
		- package: github.com/Masterminds/cookoo
		  vcs: git
		  ref: 1.1.0
		  subpackages: **
  		- package: github.com/kylelemons/go-gypsy
		  subpackages: yaml
`

func main() {
	reg, router, cxt := cookoo.Cookoo()

	routes(reg, cxt)

	if err := router.HandleRequest("@startup", cxt, false); err != nil {
		fmt.Printf("Oops! %s\n", err)
		os.Exit(1)
	}

	/*
	next := cxt.Get("subcommand", "help").(string)
	if router.HasRoute(next) {
		if err := router.HandleRequest(next, cxt, false); err != nil {
			fmt.Printf("Oops! %s\n", err)
			os.Exit(1)
		}
	} else {
		if err := router.HandleRequest("@plugin", cxt, false); err != nil {
			fmt.Printf("Oops! %s\n", err)
			os.Exit(1)
		}
	}
	*/

}

func routes(reg *cookoo.Registry, cxt cookoo.Context) {

	flags := flag.NewFlagSet("global", flag.PanicOnError)
	flags.Bool("h", false, "Print help text.")
	flags.Bool("q", false, "Quiet (no info or debug messages)")

	cxt.Put("os.Args", os.Args)

	reg.Route("@startup", "Parse args and send to the right subcommand.").
		Does(cli.ShiftArgs, "_").Using("n").WithDefault(1).
		Does(cli.ParseArgs, "remainingArgs").
		Using("flagset").WithDefault(flags).
		Using("args").From("cxt:os.Args").
		Does(cli.ShowHelp, "help").
		Using("show").From("cxt:h cxt:help").
		Using("summary").WithDefault(Summary).
		Using("usage").WithDefault(Usage).
		Using("flags").WithDefault(flags).
		Does(cmd.BeQuiet, "quiet").
		Using("quiet").From("cxt:q").
		//Does(subcommand, "subcommand").
		Does(cli.RunSubcommand, "subcommand").
		Using("default").WithDefault("help").
		Using("offset").WithDefault(0).
		Using("args").From("cxt:remainingArgs")


	reg.Route("@ready", "Prepare for glide commands.").
		Does(cmd.ReadyToGlide, "ready").
		Does(cmd.ParseYaml, "cfg")

	reg.Route("help", "Print help.").
		Does(cli.ShowHelp, "help").
		Using("show").WithDefault(true).
		Using("summary").WithDefault(Summary).
		Using("usage").WithDefault(Usage).
		Using("flags").WithDefault(flags)

	reg.Route("version", "Print the version and exit.").Does(showVersion, "_")

	reg.Route("into", "Creates a new Glide shell.").
		Does(cmd.AlreadyGliding, "isGliding").
		Does(cli.ShiftArgs, "toPath").Using("n").WithDefault(2).
		Does(cmd.Into, "in").Using("into").From("cxt:toPath").
		Using("into").WithDefault("").From("cxt:toPath").
		Includes("@ready")

	reg.Route("in", "Set GOPATH and supporting env vars.").
		Does(cmd.AlreadyGliding, "isGliding").
		Includes("@ready").
		//Does(cli.ShiftArgs, "toPath").Using("n").WithDefault(1).
		Does(cmd.Into, "in").
		Using("into").WithDefault("").From("cxt:toPath").
		Using("conf").From("cxt:cfg")

	reg.Route("gopath", "Return the GOPATH for the present project.").
		Does(cmd.In, "gopath")

	reg.Route("out", "Set GOPATH back to former val.").
		Does(cmd.Out, "gopath")

	reg.Route("install", "Install dependencies.").
		Includes("@ready").
		Does(cmd.Mkdir, "dir").Using("dir").WithDefault("_vendor").
		Does(cmd.LinkPackage, "alias").
		Does(cmd.GetImports, "dependencies").Using("conf").From("cxt:cfg").
		Does(cmd.SetReference, "version").Using("conf").From("cxt:cfg").
		Does(cmd.Rebuild, "rebuild").Using("conf").From("cxt:cfg")

	reg.Route("up", "Update dependencies (alias of 'update')").
		Does(cookoo.ForwardTo, "fwd").Using("route").WithDefault("update")

	reg.Route("update", "Update dependencies.").
		Includes("@ready").
		Does(cmd.CowardMode, "_").
		Does(cmd.UpdateImports, "dependencies").Using("conf").From("cxt:cfg").
		Does(cmd.SetReference, "version").Using("conf").From("cxt:cfg").
		Does(cmd.Rebuild, "rebuild").Using("conf").From("cxt:cfg")

	reg.Route("rebuild", "Rebuild dependencies").
		Includes("@ready").
		Does(cmd.CowardMode, "_").
		Does(cmd.Rebuild, "rebuild").Using("conf").From("cxt:cfg")

	reg.Route("init", "Initialize Glide").
		Does(cmd.InitGlide, "init")

	reg.Route("status", "Status").
		Does(cmd.Status, "status")

	reg.Route("@plugin", "Try to send to a plugin.").
		Includes("@ready").
		Does(cmd.DropToShell, "plugin")
}

/* Switched to cli.RunSubcommand
func subcommand(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	args := p.Get("args", []string{"help"}).([]string)
	if len(args) == 0 {
		return "help", nil
	}
	return args[0], nil
}
*/

func showVersion(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fmt.Println(version)
	return version, nil
}
