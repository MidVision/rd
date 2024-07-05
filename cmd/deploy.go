// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

// TODO: default to localhost target.
// TODO: provide log file path as a flag.
// FIXME: "Configuration Name" not being shown in deployment summary.

package cmd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var deployPackage, targetName string
var synchronous, logfile bool

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
		if quiet {
			os.Stdout = nil
		}
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
			printStdError("\nInvalid target name '%s'\n", targetName)
			printStdError("The target name has to include the server, the ")
			printStdError("installation and the configuration names:\n")
			printStdError("e.g. SERVER.INSTALLATION.CONFIGURATION\n\n")
			os.Exit(1)
		}
		serverName := targetStrip[0]
		installName := targetStrip[1]
		configName := targetStrip[2]

		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		var urlBuffer bytes.Buffer
		urlBuffer.WriteString("deployment/" + projectName + "/runjob/deploy/" + serverName + "/" + installName + "/" + configName +
			"?packageName=" + deployPackage)
		for _, dictionaryArg := range dictionaryArguments {
			urlBuffer.WriteString("&dictionaryItem=" + dictionaryArg)
		}

		// Perform the REST call to get the data
		resData, _, _ := rdClient.call(http.MethodPut, urlBuffer.String(), nil, "text/xml", false)

		// Print deployment information in a table
		printTable := true
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoMergeCells(true)
		for _, message := range getDeploymentMessages(resData) {
			if strings.Contains(message.Span[1], "No entity found") {
				printTable = false
			} else {
				replacedTitle := strings.Replace(message.Span[0]+":", "Deployment Job ", "", -1)
				table.Append([]string{replacedTitle, message.Span[1]})
			}
		}

		if printTable {
			fmt.Println()
			table.Render()
			fmt.Println()
		} else {
			printStdError("\nInvalid target name '%s'\n", targetName)
			printStdError("Please check the server and the installation names.\n\n")
			os.Exit(1)
		}

		// Deploying project synchronously
		if synchronous {
			fmt.Println("Deploying project in synchronous mode...")
			jobId := getJobId(resData)
			checkSynchronousDeploy(jobId)
			fmt.Println()
		}
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
	deployCmd.Flags().BoolVarP(&synchronous, "sync", "s", false, "Waits for the deployment to finish.")
	deployCmd.Flags().BoolVarP(&logfile, "logfile", "l", false, "Retrieves the deployment log file. It must be used with the 'sync' option.")
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

func checkSynchronousDeploy(jobId string) {
	logFilename := ""
	timeToSleep := 0 * time.Second
	jobRunning := true
	for jobRunning {
		time.Sleep(timeToSleep)
		resData, _, _ := rdClient.call(http.MethodGet, "deployment/display/job/"+jobId, nil, "text/xml")
		jobStatus := getjobStatus(resData)
		fmt.Println("> Deployment status: " + jobStatus)
		if jobStatus == "DEPLOYING" || jobStatus == "QUEUED" || jobStatus == "STARTING" || jobStatus == "EXECUTING" {
			fmt.Println("  Deployment running, next check in 5 seconds...")
			timeToSleep = 5 * time.Second
		} else if jobStatus == "REQUESTED" || jobStatus == "REQUESTED_SCHEDULED" {
			fmt.Println("  Deployment in a REQUESTED state. Approval may be required in RapidDeploy to continue with the execution, next check in 30 seconds...")
			timeToSleep = 30 * time.Second
		} else if jobStatus == "SCHEDULED" {
			fmt.Println("  Deployment in a SCHEDULED state, the execution will start in a future date, next check in 5 minutes...")
			fmt.Println("  > Printing out deployment details: ")
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetAutoMergeCells(true)
			for _, message := range getDeploymentMessages(resData) {
				table.Append([]string{message.Span[0], message.Span[1]})
			}
			table.Render()
			timeToSleep = 300 * time.Second
		} else {
			jobRunning = false
			logFilename = getLogFilename(resData)
			fmt.Println("Deployment finished with status: " + jobStatus)
			if jobStatus != "FAILED" && jobStatus != "REJECTED" && jobStatus != "CANCELLED" && jobStatus != "UNEXECUTABLE" && jobStatus != "TIMEDOUT" && jobStatus != "UNKNOWN" {
				fmt.Printf("Project '%s' successfully deployed!\n", projectName)
			}
		}
	}
	if logFilename != "" && logfile {
		logFilePath, err := filepath.Abs(logFilename)
		if err != nil {
			printStdError("%v\n\n", err)
			os.Exit(1)
		}
		resData, _, _ := rdClient.call(http.MethodGet, "deployment/showlog/job/"+jobId, nil, "text/xml")
		err = os.WriteFile(logFilePath, resData, 0644)
		if err != nil {
			printStdError("\nUnable to create file: %s\n", logFilePath)
			printStdError("%v\n\n", err)
			os.Exit(1)
		}
		fmt.Printf("Log file available at '%s'\n", logFilePath)
	}
}

func getDeploymentMessages(htmlContent []byte) []*Li {
	htmlObject := new(Html)
	err := xml.Unmarshal(htmlContent, &htmlObject)
	if err != nil {
		printStdError("\n%v\n\n", err)
		os.Exit(1)
	}
	return htmlObject.Body.Div[1].Div[0].Ul.Li
}

func getjobStatus(htmlContent []byte) string {
	for _, message := range getDeploymentMessages(htmlContent) {
		if strings.Contains(message.Span[0], "Job Status") {
			return message.Span[1]
		}
	}
	return ""
}

func getJobId(htmlContent []byte) string {
	for _, message := range getDeploymentMessages(htmlContent) {
		if strings.Contains(message.Span[0], "Job ID") {
			return message.Span[1]
		}
	}
	return ""
}

func getLogFilename(htmlContent []byte) string {
	for _, message := range getDeploymentMessages(htmlContent) {
		if strings.Contains(message.Span[0], "File Path") {
			return filepath.Base(message.Span[1])
		}
	}
	return ""
}
