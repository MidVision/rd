// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

// TODO: unify all error and standard messages with constants.

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// This is the client used to perform the REST calls to RapidDeploy.
// It will be used in the different commands.
var rdClient *RDClient = &RDClient{}

// For debugging purposes
var debug, quiet bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "rd",
	Short:   "Command line interface for the RapidDeploy tool.",
	Long:    `RapidDeploy CLI - Command line interface for the RapidDeploy tool.`,
	Version: "1.4",
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
	RootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Executes in quiet mode. Does not show any output.")
}
