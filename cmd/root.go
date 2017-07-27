// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

// This is the client used to perform the REST calls to RapidDeploy.
// It will be used in the different commands.
var rdClient *RDClient

// For debugging purposes
var debug bool = false

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "rd",
	Short: "Command line interface for the RapidDeploy tool.",
	Long:  `RapidDeploy CLI - Command line interface for the RapidDeploy tool.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Shows debugging information.")

	// Initialize the RapidDeploy client for the different REST calls
	rdClient = &RDClient{client: http.DefaultClient}
}
