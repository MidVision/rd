// Copyright © 2024 Rafael Ruiz Palacios <support@midvision.com>

// TODO: make the RDClient more generic so it can handle generic PUT calls.

package cmd

import (
	"bytes"
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
		fmt.Println()
		// Check the correct number of arguments
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		} else {
			importProjectPath = args[0]
		}

		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Resolve the absolute URL for the request
		reqUrl, err := rdClient.BaseUrl.Parse(rdClient.BaseUrl.EscapedPath() + "/ws/project/import")
		if err != nil {
			fmt.Printf("%v\n\n", err.Error())
			os.Exit(1)
		}

		// Prepare the body of the request
		fileArray, err := ioutil.ReadFile(importProjectPath)

		if debug {
			fmt.Printf("[DEBUG] Importing project file: %v\n", importProjectPath)
		}

		// Create the HTTP request
		req, err := http.NewRequest(http.MethodPut, reqUrl.String(), bytes.NewBuffer(fileArray))
		if err != nil {
			fmt.Printf("%v\n\n", err.Error())
			os.Exit(1)
		}

		// Set the headers of the request
		req.Header.Add("Content-Type", "application/zip")
		req.Header.Add("Authorization", rdClient.AuthToken)

		if debug {
			fmt.Printf("[DEBUG] Request URL = %v\n", req.URL)
			fmt.Printf("[DEBUG] Request method = %v\n", req.Method)
			fmt.Printf("[DEBUG] Request header = %v\n", req.Header)
			fmt.Printf("[DEBUG] Authentication token = %v\n", rdClient.AuthToken)
		}

		// Perform the request
		res, err := rdClient.client.Do(req)
		if err != nil {
			fmt.Printf("%v\n\n", err.Error())
			os.Exit(1)
		}

		// Read the response
		resData, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("%v\n\n", err.Error())
			os.Exit(1)
		}

		if debug {
			fmt.Printf("[DEBUG] Response code = %v\n", res.StatusCode)
			fmt.Printf("[DEBUG] Response body = %v\n", string(resData))
		}

		res.Body.Close()

		if res.StatusCode == 200 {
			fmt.Println("File '" + importProjectPath + "' imported successfuly.")
		} else if res.StatusCode == 400 {
			// Initialize the object that will contain the unmarshalled XML response
			htmlResponse := new(Html)
			// Unmarshall the XML response
			err = xml.Unmarshal(resData, &htmlResponse)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if len(htmlResponse.Body.Div[1].Div[0].Ul.Li) != 0 {
				for _, message := range htmlResponse.Body.Div[1].Div[0].Ul.Li {
					fmt.Println(message.Span[1])
				}
			}
		} else {
			fmt.Println("Unexpected error ocurred, please run the command with the debug flag for further information.")
		}
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(importCmd)
}
