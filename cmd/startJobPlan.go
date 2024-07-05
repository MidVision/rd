// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

// TODO: synchronous deploy.

package cmd

import (
	"encoding/xml"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"net/http"
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
		if quiet {
			os.Stdout = nil
		}
		// Check the correct number of arguments
		if len(args) == 1 {
			jobPlanId = args[0]
		} else {
			cmd.Usage()
			os.Exit(1)
		}

		if _, err := strconv.Atoi(jobPlanId); err != nil {
			printStdError("\nInvalid job plan ID provided, it must be a numeric value.\n\n")
			os.Exit(1)
		}

		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		// Perform the REST call to get the data
		resData, _, _ := rdClient.call(http.MethodPut, "deployment/jobPlan/run/"+jobPlanId, nil, "text/xml", false)

		// Initialize the object that will contain the unmarshalled XML response
		rdDeploy := new(Html)
		// Unmarshall the XML response
		err := xml.Unmarshal(resData, &rdDeploy)
		if err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		// Print data in a table
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		if len(rdDeploy.Body.Div[1].Div[0].Ul.Li) != 0 {
			table.SetAutoMergeCells(true)
			for _, message := range rdDeploy.Body.Div[1].Div[0].Ul.Li {
				replacedTitle := strings.Replace(message.Span[1], "com.midvision.rapiddeploy.domain.jobplan.", "", -1)
				table.Append([]string{message.Span[0], replacedTitle})
			}
		}
		fmt.Println()
		table.Render()
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(startJobPlanCmd)
}
