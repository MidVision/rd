// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

// TODO: synchronous deploy.

package cmd

import (
	"encoding/xml"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
)

var jobPlanId string

// startJobPlanCmd represents the startJobPlan command
var startJobPlanCmd = &cobra.Command{
	Use:   "startJobPlan JOBPLAN_ID",
	Short: "Starts a RapidDeploy job plan.",
	Long: `Starts a RapidDeploy job plan.

In order to provide a job plan ID you previously
may need to run the 'listJobPlans' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		// Check the correct number of arguments
		if len(args) == 1 {
			jobPlanId = args[0]
		} else {
			cmd.Usage()
			os.Exit(1)
		}

		if _, err := strconv.Atoi(jobPlanId); err != nil {
			fmt.Println("Invalid job plan ID provided, it must be a numeric value.\n")
			os.Exit(1)
		}

		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Perform the REST call to get the data
		resData, statusCode, err := rdClient.call("PUT", "deployment/jobPlan/run/"+jobPlanId, nil)
		if err != nil {
			fmt.Printf("Unable to connect to server '%s'.\n", rdClient.BaseUrl)
			fmt.Printf("%v\n\n", err.Error())
			os.Exit(1)
		}
		if statusCode != 200 && statusCode != 400 {
			fmt.Printf("Unable to connect to server '%s'.\n", rdClient.BaseUrl)
			fmt.Printf("Please, perform a new login before requesting any action.\n\n")
			os.Exit(1)
		}

		// Initialize the object that will contain the unmarshalled XML response
		rdDeploy := new(Html)
		// Unmarshall the XML response
		err = xml.Unmarshal(resData, &rdDeploy)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Print data in a table
		printTable := true
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		if len(rdDeploy.Body.Div[1].Div[0].Ul.Li) != 0 {
			table.SetAutoMergeCells(true)
			for _, message := range rdDeploy.Body.Div[1].Div[0].Ul.Li {
				if strings.Contains(message.Span[1], "No entity found") {
					printTable = false
				} else {
					replacedTitle := strings.Replace(message.Span[1], "com.midvision.rapiddeploy.domain.jobplan.", "", -1)
					table.Append([]string{message.Span[0], replacedTitle})
				}
			}
		}
		if printTable {
			table.Render()
		} else {
			fmt.Println("Invalid target name '" + targetName + "'")
			fmt.Println("Please check the server and the installation names.")
		}
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(startJobPlanCmd)
}
