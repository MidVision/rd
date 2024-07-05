// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"text/tabwriter"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Checks if a login session is established.",
	Long: `Checks if a login session is established to
a RapidDeploy server and shows the server URL.`,
	Run: func(cmd *cobra.Command, args []string) {
		if quiet {
			os.Stdout = nil
		}
		// Load the login session file
		if err := rdClient.loadLoginFile(); err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		// Check connection to the server
		// FIXME: to check connection use 'listGroups' until we create a generic web service call!
		rdClient.call(http.MethodGet, "group/list", nil, "text/plain")

		// Session active - print message
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '*', 0)
		fmt.Fprintf(w, "\n\t\t\n")
		fmt.Fprintf(w, "\t Successfully logged in to '%s' \t\n", rdClient.BaseUrl.String())
		fmt.Fprintf(w, "\t\t\n\n")
		w.Flush()
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
