package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform/backend"
	"github.com/hashicorp/terraform/command/jsonconfig"
	"github.com/hashicorp/terraform/tfdiags"
)

// ShowConfigCommand is a Command implementation that reads and outputs the
// contents of a Terraform plan or state file.
type ShowConfigCommand struct {
	Meta
}

func (c *ShowConfigCommand) Run(args []string) int {
	args, err := c.Meta.process(args, false)
	if err != nil {
		return 1
	}

	cmdFlags := c.Meta.defaultFlagSet("show-config")
	var jsonOutput bool
	var atPos int
	cmdFlags.BoolVar(&jsonOutput, "json", false, "produce JSON output")
	cmdFlags.IntVar(&atPos, "at-pos", -1, "at position (in byte offset from the beginning of a file)")
	cmdFlags.Usage = func() { c.Ui.Error(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		c.Ui.Error(fmt.Sprintf("Error parsing command-line flags: %s\n", err.Error()))
		return 1
	}

	args = cmdFlags.Args()
	if len(args) > 1 {
		c.Ui.Error("TODO")
		cmdFlags.Usage()
		return 1
	}

	// Check for user-supplied plugin path
	if c.pluginPath, err = c.loadPluginPath(); err != nil {
		c.Ui.Error(fmt.Sprintf("Error loading plugin path: %s", err))
		return 1
	}

	var diags tfdiags.Diagnostics

	// Load the backend
	b, backendDiags := c.Backend(nil)
	diags = diags.Append(backendDiags)
	if backendDiags.HasErrors() {
		c.showDiagnostics(diags)
		return 1
	}

	// We require a local backend
	local, ok := b.(backend.Local)
	if !ok {
		c.showDiagnostics(diags) // in case of any warnings in here
		c.Ui.Error(ErrUnsupportedLocalOp)
		return 1
	}

	// the command expects the config dir to always be the cwd
	cwd, err := os.Getwd()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error getting cwd: %s", err))
		return 1
	}

	// TODO: Determine if a config file was passed to the command

	// Build the operation
	opReq := c.Operation(b)
	opReq.ConfigDir = cwd
	opReq.ConfigLoader, err = c.initConfigLoader()
	opReq.AllowUnsetVariables = true
	if err != nil {
		diags = diags.Append(err)
		c.showDiagnostics(diags)
		return 1
	}

	// Get the context
	ctx, _, ctxDiags := local.Context(opReq)
	diags = diags.Append(ctxDiags)
	if ctxDiags.HasErrors() {
		c.showDiagnostics(diags)
		return 1
	}

	// Get the schemas from the context
	schemas := ctx.Schemas()

	if jsonOutput == true {
		config := ctx.Config()
		jsonConfig, err := jsonconfig.Marshal(config, schemas)

		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to marshal config to json: %s", err))
			return 1
		}
		c.Ui.Output(string(jsonConfig))
	} else {
		// TODO
	}

	return 0
}

func (c *ShowConfigCommand) Help() string {
	helpText := `
Usage: terraform show-config [options] [path]

  Reads and outputs Terraform configuration.
  If no path is specified, the configuration
  from the current workdir will be shown.

Options:

  -json   If specified, output the configuration in
          a machine-readable form.

`
	return strings.TrimSpace(helpText)
}

func (c *ShowConfigCommand) Synopsis() string {
	return "Inspect Terraform configuration"
}
