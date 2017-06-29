// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

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

type (
	// Struct type that will hold the XML response from the REST call
	Servers struct {
		Server []*Server `xml:"Server,omitempty"`
	}

	Server struct {
		BuildStore            string                   `xml:"buildStore,omitempty"`
		Displayname           string                   `xml:"displayname,omitempty"`
		EnvironmentProperties []*EnvironmentProperties `xml:"environmentProperties,omitempty"`
		Hostname              string                   `xml:"hostname,omitempty"`
		Hostnames             []string                 `xml:"hostnames,omitempty"`
		Optlock               int                      `xml:"optlock,omitempty"`
		PluginDataSet         *PluginDataSet           `xml:"pluginDataSet,omitempty"`
		Product               string                   `xml:"product,omitempty"`
		ServerEnabled         bool                     `xml:"serverEnabled,omitempty"`
		Version               string                   `xml:"version,omitempty"`
	}

	EnvironmentProperties struct {
		Id    int    `xml:"id,omitempty"`
		Key   string `xml:"key,omitempty"`
		Value string `xml:"value,omitempty"`
	}
)

// listServersCmd represents the listServers command
var listServersCmd = &cobra.Command{
	Use:   "listServers",
	Short: "Lists the available servers in RapidDeploy.",
	Long:  `Lists the available servers in RapidDeploy.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Perform the REST call to get the data
		resData, statusCode, err := rdClient.call("GET", "server/list", nil)
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
		rdServers := new(Servers)
		// Unmarshall the XML response
		err = xml.Unmarshal(resData, &rdServers)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Print data in a table
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		if len(rdServers.Server) != 0 {
			table.SetHeader([]string{"Display name", "Hostnames", "OS type & Version", "Enabled?"})
			for _, server := range rdServers.Server {
				table.Append([]string{server.Displayname, strings.Join(server.Hostnames, "\n"),
					server.Product + " " + server.Version, strconv.FormatBool(server.ServerEnabled)})
			}
		} else {
			table.Append([]string{"No servers available to show"})
		}
		table.Render()
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(listServersCmd)
}
