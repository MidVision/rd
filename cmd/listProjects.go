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

type (
	Projects struct {
		Project []*Project `xml:"Project,omitempty"`
	}

	Project struct {
		CreateDate            string         `xml:"createDate,omitempty"`
		Description           string         `xml:"description,omitempty"`
		Enabled               bool           `xml:"enabled,omitempty"`
		LogDirectory          string         `xml:"logDirectory,omitempty"`
		Name                  string         `xml:"name,omitempty"`
		Optlock               int            `xml:"optlock,omitempty"`
		OrchestrationFileName string         `xml:"orchestrationFileName,omitempty"`
		Owner                 *Owner         `xml:"owner,omitempty"`
		PluginDataSet         *PluginDataSet `xml:"pluginDataSet,omitempty"`
	}

	Owner struct {
		Description string `xml:"description,omitempty"`
		Email       string `xml:"email,omitempty"`
		Enabled     bool   `xml:"enabled,omitempty"`
		Firstname   string `xml:"firstname,omitempty"`
		Lastname    string `xml:"lastname,omitempty"`
		Optlock     int    `xml:"optlock,omitempty"`
		SourceType  bool   `xml:"sourceType,omitempty"`
		Username    string `xml:"username,omitempty"`
	}
)

// listProjectsCmd represents the listProjects command
var listProjectsCmd = &cobra.Command{
	Use:   "listProjects",
	Short: "Lists the available projects in RapidDeploy.",
	Long:  `Lists the available projects in RapidDeploy.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		// Perform the REST call to get the data
		resData, _, _ := rdClient.call(http.MethodGet, "project/list", nil, "text/xml")

		// Initialize the object that will contain the unmarshalled XML response
		rdProjects := new(Projects)
		// Unmarshall the XML response
		err := xml.Unmarshal(resData, &rdProjects)
		if err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		// Print data in a table
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		if len(rdProjects.Project) != 0 {
			table.SetHeader([]string{"Name", "Description"})
			for _, project := range rdProjects.Project {
				table.Append([]string{project.Name, project.Description})
			}
		} else {
			table.Append([]string{"No projects available to show"})
		}
		fmt.Println()
		table.Render()
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(listProjectsCmd)
}
