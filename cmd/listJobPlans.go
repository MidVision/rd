// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

package cmd

import (
	"encoding/xml"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
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
		fmt.Println()
		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Perform the REST call to get the data
		resData, statusCode, err := rdClient.call("GET", "deployment/jobPlan/list", nil, "text/xml")
		if err != nil {
			fmt.Printf("Unable to connect to server '%s'.\n", rdClient.BaseUrl)
			fmt.Printf("%v\n\n", err.Error())
			os.Exit(1)
		}
		if statusCode != 200 {
			fmt.Printf("Unable to connect to server '%s'.\n", rdClient.BaseUrl)
			fmt.Printf("Please, perform a new login before requesting any action.\n\n")
			os.Exit(1)
		}

		// Initialize the object that will contain the unmarshalled XML response
		rdJobPlans := new(JobPlans)
		// Unmarshall the XML response
		err = xml.Unmarshal(resData, &rdJobPlans)
		if err != nil {
			fmt.Println(err)
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
		table.Render()
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(listJobPlansCmd)
}
