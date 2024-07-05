// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Finishes the session with the RapidDeploy server.",
	Long: `This command finishes the session with the RapidDeploy server.

It performs a logout from the RapidDeploy server.`,
	Run: func(cmd *cobra.Command, args []string) {
		if quiet {
			os.Stdout = nil
		}

		// Remove the login session file - log out
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '*', 0)
		fmt.Fprintf(w, "\n\t\t\n")
		if err := rdClient.removeLoginFile(); err != nil {
			fmt.Fprintf(w, "\t WARNING: No login session found. Please, perform a login before requesting any action. \t\n")
		} else {
			fmt.Fprintf(w, "\t Successfully logged out from RapidDeploy. \t\n")
		}
		fmt.Fprintf(w, "\t\t\n\n")
		w.Flush()
	},
}

func init() {
	RootCmd.AddCommand(logoutCmd)
}
