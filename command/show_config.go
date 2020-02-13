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
	cmdFlags.IntVar(&atPos, "at-pos", -1, "at position (in byte offset format)")
	cmdFlags.Usage = func() { c.Ui.Error(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		c.Ui.Error(fmt.Sprintf("Error parsing command-line flags: %s\n", err.Error()))
		return 1
	}

	configPath, err := ModulePath(cmdFlags.Args())
	if err != nil {
		c.Ui.Error(err.Error())
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

  -at-pos At position (in byte offset format)

  -json   If specified, output the configuration in
          a machine-readable form.

`
	return strings.TrimSpace(helpText)
}

func (c *ShowConfigCommand) Synopsis() string {
	return "Inspect Terraform configuration"
}
