// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

// TODO: use direct methods, direct actions, instead of just one 'call' method.

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
)

const (
	// The filename to save the connection details in the home folder.
	loginFile = ".rapiddeploy"
)

type (
	// The client for the REST calls to RapidDeploy.
	RDClient struct {
		// URL and authentication token used to perform the calls.
		// These parameters will be saved as a JSON file in the home folder.
		BaseUrl   *url.URL `json:"url"`
		AuthToken string   `json:"token"`
		Username  string   `json:"param1"`
		Password  string   `json:"param2"`
	}

	// Declared here as it is used in different entity structs.
	PluginDataSet struct {
		Id         int16  `xml:"id,omitempty"`
		PluginData string `xml:"pluginData,omitempty"`
	}

	//**********************************************
	// Generic structure of a web service response.
	Html struct {
		Head *Head `xml:"head,omitempty"`
		Body *Body `xml:"body,omitempty"`
	}

	Head struct {
		Title string `xml:"title,omitempty"`
	}

	Body struct {
		Div []*Div `xml:"div,omitempty"`
	}

	Div struct {
		H2  string `xml:"h2,omitempty"`
		Div []*Div `xml:"div,omitempty"`
		Ul  *Ul    `xml:"ul,omitempty"`
	}

	Ul struct {
		Li []*Li `xml:"li,omitempty"`
	}

	Li struct {
		Span []string `xml:"span,omitempty"`
	}
	//**********************************************
)

func (rdc *RDClient) loadLoginFile() error {
	loginFilePath := path.Join(getHome(), loginFile)

	if _, err := os.Stat(loginFilePath); err != nil {
		return fmt.Errorf("No login session found!\nPlease, perform a login before requesting any action.")
	}

	content, err := ioutil.ReadFile(loginFilePath)
	if err != nil {
		return fmt.Errorf("Invalid login session found!\nPlease, perform a new login before requesting any action.")
	}

	if err := json.Unmarshal(content, rdc); err != nil {
		return fmt.Errorf("Invalid login session found!\nPlease, perform a new login before requesting any action.")
	}
	return nil
}

func (rdc *RDClient) saveLoginFile() error {
	loginFilePath := path.Join(getHome(), loginFile)

	content, err := json.MarshalIndent(rdc, "", "\t")
	if err != nil {
		return err
	} else {
		return ioutil.WriteFile(loginFilePath, content, 0600)
	}
}

func (rdc *RDClient) removeLoginFile() error {
	loginFilePath := path.Join(getHome(), loginFile)
	return os.Remove(loginFilePath)
}

func (rdc *RDClient) call(method string, relUrl string, bodyContent []byte, contentType string, check400Arg ...bool) (resData []byte, statusCode int, err error) {
	check400 := true
	if len(check400Arg) > 0 {
		check400 = check400Arg[0]
	}

	if rdc.BaseUrl == nil {
		return nil, -1, fmt.Errorf("No URL found in login session.\nPlease, perform a new login before requesting any action.")
	}

	if rdc.AuthToken == "" {
		return nil, -1, fmt.Errorf("No authentication token found in login session.\nPlease, perform a new login before requesting any action.")
	}

	// Resolve the absolute URL for the request
	reqUrl, err := rdc.BaseUrl.Parse(rdc.BaseUrl.EscapedPath() + "/ws/" + relUrl)
	if err != nil {
		return nil, -1, err
	}

	header := make(map[string]string)
	// Set the headers of the request
	if contentType == "" {
		contentType = "text/plain"

	}
	header["Content-Type"] = contentType
	header["Authorization"] = rdc.AuthToken

	if debug {
		fmt.Printf("[DEBUG] Authentication token = %v\n", rdc.AuthToken)
	}

	resData, statusCode, err = call(method, reqUrl.String(), bodyContent, header)
	if err != nil {
		printStdError("\nUnable to connect to server '%s'.\n", rdc.BaseUrl)
		printStdError("%v\n\n", err)
		os.Exit(1)
	}
	if (statusCode != 200 && statusCode != 400) || (statusCode == 400 && check400) {
		printStdError("\nUnable to connect to server '%s'.\n", rdc.BaseUrl)
		printStdError("Server returned response code %v: %v\n\n", statusCode, http.StatusText(statusCode))
		if statusCode == 401 {
			printStdError("Please, perform a new login before requesting any action.\n\n")
		}
		os.Exit(1)
	}
	return
}
