// Copyright © 2024 Rafael Ruiz Palacios <support@midvision.com>

package cmd

import (
	"encoding/xml"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
)

var importProjectPath string

var importCmd = &cobra.Command{
	Use:   "import PROJECT_FILE_PATH",
	Short: "Imports a project into RapidDeploy.",
	Long:  `This command imports a project into RapidDeploy. You need to provide the absolute, or relative to current directory, path to the project ZIP file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if quiet {
			os.Stdout = nil
		}
		// Check the correct number of arguments
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		} else {
			importProjectPath = args[0]
		}

		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		// Prepare the body of the request
		fileArray, err := ioutil.ReadFile(importProjectPath)
		if err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		if debug {
			fmt.Printf("[DEBUG] Importing project file: %v\n", importProjectPath)
		}

		/*************** Import the project archive ***************/
		resData, statusCode, _ := rdClient.call(http.MethodPut, "project/import", fileArray, "application/zip", false)
		if statusCode == 200 {
			fmt.Println()
			fmt.Println("File '" + importProjectPath + "' imported successfuly.")
			fmt.Println()
		} else if statusCode == 400 {
			htmlResponse := new(Html)
			err = xml.Unmarshal(resData, &htmlResponse)
			if err != nil {
				printStdError("\n%v\n\n", err)
				os.Exit(1)
			}
			if len(htmlResponse.Body.Div[1].Div[0].Ul.Li) != 0 {
				for _, message := range htmlResponse.Body.Div[1].Div[0].Ul.Li {
					printStdError("\n%v\n\n", message.Span[1])
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(importCmd)
}
