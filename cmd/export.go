// Copyright © 2024 Rafael Ruiz Palacios <support@midvision.com>

// TODO: provide output path as a flag.
// TODO: create directories of the output flag recursively.

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"path/filepath"
)

var exportProjectName string

var exportCmd = &cobra.Command{
	Use:   "export PROJECT_NAME",
	Short: "Exports a project from RapidDeploy.",
	Long:  `This command exports a project from RapidDeploy into the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		if quiet {
			os.Stdout = nil
		}
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
		resData, statusCode, _ := rdClient.call(http.MethodGet, "project/"+exportProjectName+"/export", nil, "application/zip", false)
		if statusCode == 400 {
			printStdError("\nInvalid project name: %s\n\n", exportProjectName)
			os.Exit(1)
		}
		exportProjectFile := exportProjectName + ".zip"
		exportProjectAbsPath, err := filepath.Abs(exportProjectFile)
		if err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}
		err = os.WriteFile(exportProjectAbsPath, resData, 0644)
		if err != nil {
			printStdError("\nUnable to create file: %s\n", exportProjectAbsPath)
			printStdError("%v\n\n", err)
			os.Exit(1)
		}

		// Show resulting ZIP file
		fmt.Println()
		fmt.Println("Project export file: " + exportProjectAbsPath)
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(exportCmd)
}
