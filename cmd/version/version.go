package version

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/metal-toolbox/component-inventory/cmd"
	"github.com/metal-toolbox/component-inventory/internal/version"
	"github.com/spf13/cobra"
)

var extended bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "get the current version",
	Run: func(c *cobra.Command, args []string) {
		//nolint:errcheck
		if extended {
			var out bytes.Buffer
			json.Indent(&out, version.Current().MustJSON(), "", "  ")
			out.WriteTo(os.Stdout)
			return
		}
		fmt.Printf("%s\n", version.Current().String())
	},
}

func init() {
	cmd.RootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&extended, "extended", "e", false, "extended build version info")
}
