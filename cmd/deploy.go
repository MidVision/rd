// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

// TODO: provide a properties file as data ditionary.
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
	Use:   "deploy PROJECT_NAME [TARGET_NAME] [@@DICTIONARY_KEY_1@@=DICTIONARY_VALUE_1 @@DICTIONARY_KEY_2@@=DICTIONARY_VALUE_2 ...]",
	Short: "Deploys a RapidDeploy project to a specific target.",
	Long: `Deploys a RapidDeploy project's deploymen package to a specific target (i.e. SERVER.INSTALLATION.CONFIGURATION).

If no target name is specified the first one found with a 'localhost' hostname will be used.`,
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

		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		if targetName == "" {
			targetName = getDefaultTargetName(projectName)
		}

		if debug {
			fmt.Printf("[DEBUG] Deploying '%s' to '%s' with package '%s'...\n", projectName, targetName, deployPackage)
		}

		targetStrip := strings.Split(targetName, ".")
		if len(targetStrip) != 3 {
			printStdError("\nInvalid target name '%s'\n", targetName)
			printStdError("The target name has to include the server, the installation and the configuration names:\n")
			printStdError("e.g. SERVER.INSTALLATION.CONFIGURATION\n\n")
			os.Exit(1)
		}
		serverName := targetStrip[0]
		installName := targetStrip[1]
		configName := targetStrip[2]

		var urlBuffer bytes.Buffer
		urlBuffer.WriteString("deployment/" + projectName + "/runjob/deploy/" + serverName + "/" + installName + "/" + configName +
			"?packageName=" + deployPackage)
		for _, dictionaryArg := range dictionaryArguments {
			urlBuffer.WriteString("&dictionaryItem=" + dictionaryArg)
		}

		resData, resCode, _ := rdClient.call(http.MethodPut, urlBuffer.String(), nil, "text/xml", false)

		// Print deployment information in a table
		var table *tablewriter.Table
		if resCode == 200 {
			table = tablewriter.NewWriter(os.Stdout)
		} else {
			table = tablewriter.NewWriter(os.Stderr)
		}
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoMergeCells(true)
		for _, message := range getResponseMessages(resData) {
			if strings.Contains(message.Span[1], "No entity found") {
				printStdError("\nInvalid target name '%s'\n", targetName)
				printStdError("Please check the server and the installation names.\n\n")
				os.Exit(1)
			} else {
				replacedTitle := strings.Replace(message.Span[0]+":", "Deployment Job ", "", -1)
				table.Append([]string{replacedTitle, message.Span[1]})
			}
		}
		fmt.Println()
		table.Render()
		fmt.Println()

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
	deployCmd.Flags().StringVarP(&deployPackage, "package", "p", "", "The deployment package to deploy. It defaults to the latest version.")
}

// Returns a boolean value showing if the arguments were properly
// provided and the section of the 'items' slice that contains
// the elements that comply with a dictionary item argument syntax.
func parseArguments(args []string) (bool, []string) {
	// Not enough arguments
	if len(args) < 1 {
		return false, []string{}
	}

	// The first argument must be the project name
	if isDictionaryArg(args[0]) {
		return false, []string{}
	} else {
		projectName = args[0]
	}

	var dictionaryItems []string
	for i := 1; i < len(args); i++ {
		if isDictionaryArg(args[i]) {
			dictionaryItems = append(dictionaryItems, args[i])
		} else if len(dictionaryItems) == 0 && targetName == "" {
			targetName = args[i]
		} else {
			return false, []string{}
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

func getDefaultTargetName(projectName string) string {
	if debug {
		fmt.Printf("[DEBUG] Getting default target for project '%s'\n", projectName)
	}
	resData, resCode, _ := rdClient.call(http.MethodGet, "project/"+projectName+"/list", nil, "text/xml", false)
	rdTargets := new(Html)
	err := xml.Unmarshal(resData, &rdTargets)
	if err != nil {
		printStdError("\n%v\n\n", err)
		os.Exit(1)
	}
	if resCode != 200 {
		table := tablewriter.NewWriter(os.Stderr)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoMergeCells(true)
		for _, message := range getResponseMessages(resData) {
			replacedTitle := strings.Replace(message.Span[0]+":", "Deployment Job ", "", -1)
			table.Append([]string{replacedTitle, message.Span[1]})
		}
		printStdError("\n")
		table.Render()
		printStdError("\n")
		os.Exit(1)
	}
	if len(rdTargets.Body.Div[1].Div[0].Ul.Li) != 0 {
		for _, target := range rdTargets.Body.Div[1].Div[0].Ul.Li {
			targetName := target.Span[1]
			if debug {
				fmt.Printf("[DEBUG] Checking hostnames for target '%s'\n", targetName)
			}
			serverName := strings.Split(targetName, ".")[0]
			resData, _, _ := rdClient.call(http.MethodGet, "server/"+serverName, nil, "text/xml", false)
			server := new(Server)
			err := xml.Unmarshal(resData, &server)
			if err != nil {
				printStdError("\n%v\n\n", err)
				os.Exit(1)
			}
			if debug {
				fmt.Printf("[DEBUG] Hostnames found: %s\n", server.Hostnames)
			}
			if strings.Contains(server.Hostname, "localhost") {
				return targetName
			}
		}
	}
	return ""
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
			for _, message := range getResponseMessages(resData) {
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

func getResponseMessages(htmlContent []byte) []*Li {
	htmlObject := new(Html)
	err := xml.Unmarshal(htmlContent, &htmlObject)
	if err != nil {
		printStdError("\n%v\n\n", err)
		os.Exit(1)
	}
	return htmlObject.Body.Div[1].Div[0].Ul.Li
}

func getjobStatus(htmlContent []byte) string {
	for _, message := range getResponseMessages(htmlContent) {
		if strings.Contains(message.Span[0], "Job Status") {
			return message.Span[1]
		}
	}
	return ""
}

func getJobId(htmlContent []byte) string {
	for _, message := range getResponseMessages(htmlContent) {
		if strings.Contains(message.Span[0], "Job ID") {
			return message.Span[1]
		}
	}
	return ""
}

func getLogFilename(htmlContent []byte) string {
	for _, message := range getResponseMessages(htmlContent) {
		if strings.Contains(message.Span[0], "File Path") {
			return filepath.Base(message.Span[1])
		}
	}
	return ""
}
