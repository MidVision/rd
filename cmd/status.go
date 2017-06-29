// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
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
		fmt.Println()
		// Load the login session file
		if err := rdClient.loadLoginFile(); err != nil {
			// If any error is thrown the session doesn't exist
			fmt.Println(err.Error())
			// FIXME: to check connection use 'listGroups' until we create a generic web service call!
		} else if _, statusCode, err := rdClient.call("GET", "group/list", nil); err != nil || statusCode != 200 {
			fmt.Printf("Unable to connect to server '%s'.\n", rdClient.BaseUrl)
			fmt.Printf("Please, perform a new login before requesting any action.\n\n")
		} else {
			// Otherwise the session is already stablished
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 0, 8, 1, '*', 0)
			fmt.Fprintf(w, "\t\t\n")
			fmt.Fprintf(w, "\t Successfully logged in to '%s' \t\n", rdClient.BaseUrl.String())
			fmt.Fprintf(w, "\t\t\n")
			fmt.Fprintln(w)
			w.Flush()
		}
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
