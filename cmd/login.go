// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"net/url"
	"os"
	"text/tabwriter"
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
		// Parse the URL passed as a flag parameter
		parsedUrl, err := url.Parse(rdUrl)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Initialize the rdClient struct that will be saved
		// in the login session file for future calls to RapidDeploy
		rdClient.BaseUrl = parsedUrl
		rdClient.Username = username
		rdClient.Password = password

		// This is necessary so an error is not thrown for empty authentication token
		rdClient.AuthToken = "token"

		resData, statusCode, err := rdClient.call("POST", "user/create/token", rdClient, "text/plain")
		rdClient.AuthToken = string(resData)

		// Perform a call to see the URL and authontication token are correct
		fmt.Printf("Trying to log in to '%s'...\n\n", rdUrl)
		// FIXME: to check connection use 'listGroups' until we create a generic web service call!
		_, statusCode, err = rdClient.call("GET", "group/list", nil, "text/xml")

		// Login failed - the call throws an error
		if err != nil {
			fmt.Printf("Unable to connect to server '%s'.\n", parsedUrl)
			fmt.Printf("%v\n\n", err.Error())
			os.Exit(1)
		}

		// Login failed - the call returns a failure status code
		if statusCode != 200 {
			fmt.Printf("Unable to connect to server '%s'.\n", parsedUrl)
			fmt.Printf("Server returned response code %v: %v\n", statusCode, http.StatusText(statusCode))
			fmt.Printf("Please check the credentials.\n\n")
			os.Exit(1)
		}

		// Print table of successful login
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '*', 0)
		fmt.Fprintf(w, "\t\t\n")
		fmt.Fprintf(w, "\t Successfully logged in to '%s' \t\n", parsedUrl.String())
		fmt.Fprintf(w, "\t\t\n")
		fmt.Fprintln(w)
		w.Flush()

		// Save the login session file
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
	loginCmd.Flags().StringVarP(&password, "password", "", "mvadmin", "Password used to connect to the RapidDeploy server.")
}
