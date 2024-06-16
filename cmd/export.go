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
	Use:   "export",
	Short: "Exports a project from RapidDeploy.",
	Long:  `This command exports a project from RapidDeploy into the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		// Check the correct number of arguments
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		} else {
			exportProjectName = args[0]
		}

		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		/*************** Retrieve the project archive ***************/
		resData, statusCode, err := rdClient.call(http.MethodGet, "project/"+exportProjectName+"/export", nil, "application/zip")
		if err != nil {
			fmt.Printf("Unable to connect to server '%s'.\n", rdClient.BaseUrl)
			fmt.Printf("%v\n\n", err.Error())
			os.Exit(1)
		}
		if statusCode == 400 {
			fmt.Printf("Invalid project name: %s\n\n", exportProjectName)
			os.Exit(1)
		}
		if statusCode != 200 {
			fmt.Printf("Unable to connect to server '%s'.\n", rdClient.BaseUrl)
			fmt.Printf("Please, perform a new login before requesting any action.\n\n")
			os.Exit(1)
		}
		err = os.WriteFile(exportProjectName+".zip", resData, 0644)
		if err != nil {
			fmt.Println("Unable to create file:", exportProjectName+".zip")
			fmt.Printf("%v\n\n", err)
			os.Exit(1)
		}

		// Show resulting ZIP file
		fmt.Println("Project export file: " + exportProjectName + ".zip")
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(exportCmd)
}
