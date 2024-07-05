// Copyright © 2024 Rafael Ruiz Palacios <support@midvision.com>

// TODO: provide output path as a flag.

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var exportProjectName string

var exportCmd = &cobra.Command{
	Use:   "export PROJECT_NAME",
	Short: "Exports a project from RapidDeploy.",
	Long:  `This command exports a project from RapidDeploy into the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check the correct number of arguments
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		} else {
			exportProjectName = args[0]
		}

		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		/*************** Retrieve the project archive ***************/
		resData, statusCode, err := rdClient.call(http.MethodGet, "project/"+exportProjectName+"/export", nil, "application/zip")
		if statusCode == 400 {
			fmt.Printf("Invalid project name: %s\n\n", exportProjectName)
			os.Exit(1)
		}
		err = os.WriteFile(exportProjectName+".zip", resData, 0644)
		if err != nil {
			printStdError("\nUnable to create file: %s.zip\n", exportProjectName)
			printStdError("%v\n\n", err)
			os.Exit(1)
		}

		// Show resulting ZIP file
		fmt.Println()
		fmt.Println("Project export file: " + exportProjectName + ".zip")
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(exportCmd)
}
