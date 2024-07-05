// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

package cmd

import (
	"encoding/xml"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"strconv"
)

type (
	JobPlans struct {
		JobPlan []*JobPlan `xml:"JobPlan,omitempty"`
	}

	JobPlan struct {
		Description  string `xml:"description,omitempty"`
		Id           int    `xml:"id,omitempty"`
		Name         string `xml:"name,omitempty"`
		PlanData     string `xml:"planData,omitempty"`
		SecurityName string `xml:"securityName,omitempty"`
	}
)

// listJobPlansCmd represents the listJobPlans command
var listJobPlansCmd = &cobra.Command{
	Use:   "listJobPlans",
	Short: "Lists the available job plans in RapidDeploy.",
	Long:  `Lists the available job plans in RapidDeploy.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		// Perform the REST call to get the data
		resData, _, _ := rdClient.call(http.MethodGet, "deployment/jobPlan/list", nil, "text/xml")

		// Initialize the object that will contain the unmarshalled XML response
		rdJobPlans := new(JobPlans)
		// Unmarshall the XML response
		err := xml.Unmarshal(resData, &rdJobPlans)
		if err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		// FIXME: check the problem with the ASCII characters in the description!!!
		//        There is some problem with the &#xD; character and printing the table.
		// Print data in a table
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		if len(rdJobPlans.JobPlan) != 0 {
			table.SetHeader([]string{"ID", "Name", "Owner"}) //, "Description"})
			for _, jobPlan := range rdJobPlans.JobPlan {
				if len(jobPlan.Description) >= 0 {
					table.Append([]string{strconv.Itoa(int(jobPlan.Id)), jobPlan.Name, jobPlan.SecurityName}) //, jobPlan.Description})
				}
			}
		} else {
			table.Append([]string{"No job plans available to show"})
		}
		fmt.Println()
		table.Render()
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(listJobPlansCmd)
}
