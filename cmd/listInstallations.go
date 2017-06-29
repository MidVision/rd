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
	Environments struct {
		Environment []*Environment `xml:"environment,omitempty"json:"environment,omitempty"`
	}

	Environment struct {
		EnvType            *EnvType `xml:"envType,omitempty"`
		EnvTypeName        string   `xml:"envTypeName,omitempty"`
		EnvironmentEnabled bool     `xml:"environmentEnabled,omitempty"`
		Hostname           string   `xml:"hostname,omitempty"`
		Id                 int      `xml:"id,omitempty"`
		Name               string   `xml:"name,omitempty"`
		Optlock            int      `xml:"optlock,omitempty"`
		Owner              string   `xml:"owner,omitempty"`
		ServerDisplayName  string   `xml:"serverDisplayName,omitempty"`
		SnapshotsPath      string   `xml:"snapshotsPath,omitempty"`
		Validated          bool     `xml:"validated,omitempty"`
	}

	EnvType struct {
		Live                       string `xml:"live,attr"`
		Name                       string `xml:"name,attr"`
		ConfigurationApprovalGroup string `xml:"configurationApprovalGroup,omitempty"`
	}
)

var serverName string

// listInstallationsCmd represents the listInstallations command
var listInstallationsCmd = &cobra.Command{
	Use:   "listInstallations SERVER_NAME",
	Short: "Lists the available installations for a server in RapidDeploy.",
	Long:  `Lists the available installations for a server in RapidDeploy.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		// Check the correct number of arguments
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		} else {
			serverName = args[0]
		}

		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Perform the REST call to get the data
		resData, statusCode, err := rdClient.call("GET", "environment/"+serverName+"/list", nil)
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
		rdEnvironments := new(Environments)
		// Unmarshall the XML response
		err = xml.Unmarshal(resData, &rdEnvironments)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Print data in a table
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		if len(rdEnvironments.Environment) != 0 {
			table.SetHeader([]string{"Name", "Environment", "Live?", "Approval group", "Enabled?"})
			for _, environment := range rdEnvironments.Environment {
				table.Append([]string{environment.Name, environment.EnvTypeName, environment.EnvType.Live,
					environment.EnvType.ConfigurationApprovalGroup, strconv.FormatBool(environment.EnvironmentEnabled)})
			}
		} else {
			table.Append([]string{"No installations available to show for server '" + serverName + "'"})
		}
		table.Render()
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(listInstallationsCmd)
}
