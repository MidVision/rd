// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

// TODO: synchronous deployments.

package cmd

import (
	"encoding/xml"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var deployPackage, targetName string

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy PROJECT_NAME [PACKAGE_NAME] TARGET_NAME",
	Short: "Deploys a RapidDeploy project to a specified target.",
	Long: `Deploys a specified version of a RapidDeploy project
(deploymen package) to a specified project target
(i.e. SERVER.INSTALLATION.CONFIGURATION).

If no package name is specified the latest version
is used by default.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		// Check the correct number of arguments
		if len(args) == 2 {
			projectName = args[0]
			deployPackage = ""
			targetName = args[1]
		} else if len(args) == 3 {
			projectName = args[0]
			deployPackage = args[1]
			targetName = args[2]
		} else {
			cmd.Usage()
			os.Exit(1)
		}

		targetStrip := strings.Split(targetName, ".")
		if len(targetStrip) != 3 {
			fmt.Println("Invalid target name '" + targetName + "'")
			fmt.Println("The target name has to include the server, the")
			fmt.Println("installation and the configuration names:")
			fmt.Println("    e.g. SERVER.INSTALLATION.CONFIGURATION\n")
			os.Exit(1)
		}
		serverName := targetStrip[0]
		installName := targetStrip[1]
		configName := targetStrip[2]

		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Perform the REST call to get the data
		resData, statusCode, err := rdClient.call("PUT", "deployment/"+
			projectName+"/runjob/deploy/"+serverName+"/"+installName+"/"+configName+
			"?packageName="+deployPackage, nil)
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
					replacedTitle := strings.Replace(message.Span[0]+":", "Deployment Job ", "", -1)
					table.Append([]string{replacedTitle, message.Span[1]})
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
	RootCmd.AddCommand(deployCmd)
}
