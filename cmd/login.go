// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

// TODO: implement the login interactively.

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"text/tabwriter"
)

const (
	awsTokenUrl   = "http://169.254.169.254/latest/api/token"
	instanceIdUrl = "http://169.254.169.254/latest/meta-data/instance-id"
	machineIdFile = "/etc/machine-id"
	defaultRdPass = "mvadmin"
)

// These values are set with the flags
var rdUrl string
var username string
var password string

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Creates a session with a RapidDeploy server.",
	Long: `This command creates a session with a RapidDeploy server.

It performs a login to a RapidDeploy server and keeps this
session active for future commands to the RapidDeploy server.

This session can be finished by calling the 'logout' command or by
calling this command again.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()

		loginResult := false

		if password == "" {
			if debug {
				fmt.Printf("[DEBUG] Loging in with default AWS password...\n")
			}
			// Get default AWS password and try first login
			header := make(map[string]string)
			header["X-aws-ec2-metadata-token-ttl-seconds"] = "21600"

			awsToken, _, _ := call(http.MethodPut, awsTokenUrl, nil, header)
			if debug {
				fmt.Printf("[DEBUG] AWS API token = %v\n", string(awsToken))
			}

			delete(header, "X-aws-ec2-metadata-token-ttl-seconds")
			header["X-aws-ec2-metadata-token"] = string(awsToken)

			instanceId, _, _ := call(http.MethodGet, instanceIdUrl, nil, header)
			if debug {
				fmt.Printf("[DEBUG] AWS instance ID = %v\n", string(instanceId))
			}
			loginResult, _ = checkLogin(rdUrl, username, string(instanceId))

			if !loginResult {
				// Get default Azure password and try second login
				machineId, _ := ioutil.ReadFile(machineIdFile)
				if debug {
					fmt.Printf("[DEBUG] Loging in with default Azure password...\n")
					fmt.Printf("[DEBUG] Azure machine ID = %v\n", string(machineId))
				}
				loginResult, _ = checkLogin(rdUrl, username, string(machineId))
			}

			if !loginResult {
				// Try default RapidDeploy password
				if debug {
					fmt.Printf("[DEBUG] Loging in with default RapidDeploy password...\n")
				}
				loginResult, _ = checkLogin(rdUrl, username, defaultRdPass)
			}
		} else {
			loginResult, _ = checkLogin(rdUrl, username, password)
		}

		if !loginResult {
			fmt.Printf("Unable to connect to server '%s'.\n", rdUrl)
			fmt.Printf("Please check the credentials.\n\n")
			os.Exit(1)
		}

		// Print table of successful login
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '*', 0)
		fmt.Fprintf(w, "\t\t\n")
		fmt.Fprintf(w, "\t Successfully logged in to '%s' \t\n", rdUrl)
		fmt.Fprintf(w, "\t\t\n")
		fmt.Fprintln(w)
		w.Flush()

		// Save the rdClient struct into the login session file for future calls to RapidDeploy
		if err := rdClient.saveLoginFile(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	// Add the command to the Cobra framework
	RootCmd.AddCommand(loginCmd)

	// The flags defined for this command
	loginCmd.Flags().StringVarP(&rdUrl, "url", "", "http://localhost:9090/MidVision", "URL used to connect to the RapidDeploy server.")
	loginCmd.Flags().StringVarP(&username, "username", "", "mvadmin", "Username used to connect to the RapidDeploy server.")
	loginCmd.Flags().StringVarP(&password, "password", "", "", "Password used to connect to the RapidDeploy server.")
}

func checkLogin(loginUrl, loginUser, loginPass string) (bool, error) {
	parsedUrl, err := url.Parse(loginUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Initialize the rdClient
	rdClient.BaseUrl = parsedUrl
	rdClient.Username = loginUser
	rdClient.Password = loginPass

	// This is necessary so an error is not thrown for empty authentication token
	rdClient.AuthToken = "token"

	reqData, err := json.Marshal(rdClient)
	if err != nil {
		if debug {
			fmt.Printf("[ERROR] Marshal rdClient: %v\n", err)
		}
		return false, err
	}
	resData, statusCode, err := rdClient.call(http.MethodPost, "user/create/token", reqData, "text/plain")
	rdClient.AuthToken = string(resData)

	if debug {
		fmt.Printf("[DEBUG] Trying to log in to '%s'...\n", loginUrl)
	}

	// Perform a ramdom call to see the URL and authentication token are correct
	// FIXME: to check connection use 'listGroups' until we create a generic web service call!
	_, statusCode, err = rdClient.call("GET", "group/list", nil, "text/xml")

	// Login failed - the call throws an error
	if err != nil {
		if debug {
			fmt.Printf("[DEBUG] Unable to connect to server '%s': %s\n", parsedUrl, err)
		}
		return false, err
	}

	// Login failed - the call returns a failure status code
	if statusCode != 200 {
		if debug {
			fmt.Printf("[DEBUG] Unable to connect to server '%s'...\n", parsedUrl)
			fmt.Printf("[DEBUG] Server returned response code %v: %v\n\n", statusCode, http.StatusText(statusCode))
		}
		return false, nil
	}
	return true, nil

}
