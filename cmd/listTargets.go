// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

package cmd

import (
	"encoding/xml"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var projectName string

// listTargetsCmd represents the listTargets command
var listTargetsCmd = &cobra.Command{
	Use:   "listTargets PROJECT_NAME",
	Short: "Lists the available targets for a project in RapidDeploy.",
	Long:  `Lists the available targets for a project in RapidDeploy.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check the correct number of arguments
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		} else {
			projectName = args[0]
		}

		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		// Perform the REST call to get the data
		resData, _, _ := rdClient.call(http.MethodGet, "project/"+projectName+"/list", nil, "text/xml", false)

		// Initialize the object that will contain the unmarshalled XML response
		rdTargets := new(Html)
		// Unmarshall the XML response
		err := xml.Unmarshal(resData, &rdTargets)
		if err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		// Print data in a table
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		if len(rdTargets.Body.Div[1].Div[0].Ul.Li) != 0 {
			table.SetHeader([]string{"Targets"})
			for _, target := range rdTargets.Body.Div[1].Div[0].Ul.Li {
				table.Append([]string{target.Span[1]})
			}
		}
		fmt.Println()
		table.Render()
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(listTargetsCmd)
}
