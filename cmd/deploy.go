// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

// TODO: synchronous deployments.
// TODO: default to localhost target.

package cmd

import (
	"bytes"
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
	Use:   "deploy PROJECT_NAME [PACKAGE_NAME] TARGET_NAME [@@DICTIONARY_KEY@@=DICTIONARY_VALUE ...]",
	Short: "Deploys a RapidDeploy project to a specified target.",
	Long: `Deploys a specified version of a RapidDeploy project
(deploymen package) to a specified project target
(i.e. SERVER.INSTALLATION.CONFIGURATION).

If no package name is specified the latest version
is used by default.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()

		// Check the arguments are properly provided
		resParse, dictionaryArguments := parseArguments(args)
		if !resParse {
			cmd.Usage()
			os.Exit(1)
		}

		if debug {
			fmt.Printf("[DEBUG] Deploying '%s' to '%s' with package '%s'...\n", projectName, targetName, deployPackage)
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

		var urlBuffer bytes.Buffer
		urlBuffer.WriteString("deployment/" + projectName + "/runjob/deploy/" + serverName + "/" + installName + "/" + configName +
			"?packageName=" + deployPackage)
		for _, dictionaryArg := range dictionaryArguments {
			urlBuffer.WriteString("&dictionaryItem=" + dictionaryArg)
		}

		// Perform the REST call to get the data
		resData, statusCode, err := rdClient.call("PUT", urlBuffer.String(), nil, "text/xml")
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

// Returns a boolean value showing if the arguments were properly
// provided and the section of the 'items' slice that contains
// the elements that comply with a dictionary item argument syntax.
func parseArguments(args []string) (bool, []string) {

	// Not enough arguments
	if len(args) < 2 {
		return false, []string{}
	}

	// The first argument must be the project name
	if isDictionaryArg(args[0]) {
		return false, []string{}
	} else {
		projectName = args[0]
	}

	// The second argument must be the deployment package or target name
	if isDictionaryArg(args[1]) {
		return false, []string{}
	} else {
		targetName = args[1]
	}

	var dictionaryItems []string
	if len(args) > 2 {

		// The third argument must be the target name or the first dictionary item
		if isDictionaryArg(args[2]) {
			dictionaryItems = append(dictionaryItems, args[2])
		} else {
			deployPackage = args[1]
			targetName = args[2]
		}

		// The rest of the arguments must be dictionary items
		for i := 3; i < len(args); i++ {
			if isDictionaryArg(args[i]) {
				dictionaryItems = append(dictionaryItems, args[i])
			} else {
				return false, []string{}
			}
		}
	}
	return true, dictionaryItems
}

// Checks if 's' complies with the dictionary item argument syntax:
// @@DICTIONARY_KEY@@=DICTIONARY_VALUE
func isDictionaryArg(s string) bool {
	keyValuePair := strings.Split(s, "=")
	if strings.Contains(s, "=") &&
		len(keyValuePair) <= 2 &&
		strings.HasPrefix(keyValuePair[0], "@@") &&
		strings.HasSuffix(keyValuePair[0], "@@") {
		return true
	} else {
		return false
	}
}
