// Copyright © 2017 Rafael Ruiz Palacios <support@midvision.com>

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
		// The HTTP client to perform the REST calls.
		client *http.Client

		// URL and authentication token used to perform the calls.
		// These parameters will be saved as a JSON file in the home folder.
		BaseUrl   *url.URL `json:"url"`
		AuthToken string   `json:"token"`
	}

	// Declared here as it is used in different entity structs.
	PluginDataSet struct {
		Id         int16  `xml:" id,omitempty"`
		PluginData string `xml:" pluginData,omitempty"`
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

// ONLY for development purposes.
// Should always be 'false' by default.
var genXML bool = false

func (rdc *RDClient) loadLoginFile() error {
	loginFilePath := path.Join(getHome(), loginFile)

	if _, err := os.Stat(loginFilePath); err != nil {
		return fmt.Errorf("No login session found!\nPlease, perform a login before requesting any action.\n")
	}

	content, err := ioutil.ReadFile(loginFilePath)
	if err != nil {
		return fmt.Errorf("Invalid login session found!\nPlease, perform a new login before requesting any action.\n")
	}

	if err := json.Unmarshal(content, rdc); err != nil {
		return fmt.Errorf("Invalid login session found!\nPlease, perform a new login before requesting any action.\n")
	} else {
		// TODO: encrypt the authentication token in the login file!!!
		// rdc.AuthToken = decodeAuthToken(rdc.AuthToken)
		if err != nil {
			return fmt.Errorf("Invalid login session found!\nPlease, perform a new login before requesting any action.\n")
		}
	}
	return nil
}

func (rdc *RDClient) saveLoginFile() error {
	loginFilePath := path.Join(getHome(), loginFile)

	// TODO: encrypt the authentication token in the login file!!!
	// rdc.AuthToken = encodeAuthToken(rdc.AuthToken)

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

func (rdc *RDClient) call(method string, relUrl string, bodyContent interface{}) ([]byte, int, error) {

	if rdc.BaseUrl == nil {
		return nil, -1, fmt.Errorf("No URL found in login session.\nPlease, perform a new login before requesting any action.\n")
	}

	if rdc.AuthToken == "" {
		return nil, -1, fmt.Errorf("No authentication token found in login session.\nPlease, perform a new login before requesting any action.\n")
	}

	// Resolve the absolute URL for the request
	reqUrl, err := rdc.BaseUrl.Parse(rdc.BaseUrl.EscapedPath() + "/ws/" + relUrl)
	if err != nil {
		return nil, -1, err
	}

	// Create the HTTP request
	// TODO: the body may be required for some services!!
	req, err := http.NewRequest(method, reqUrl.String(), nil)
	if err != nil {
		return nil, -1, err
	}

	// Set the headers of the request
	req.Header.Add("Content-Type", "text/xml")
	req.Header.Add("Authorization", rdc.AuthToken)

	if debug {
		fmt.Printf("[DEBUG] Request URL = %v\n", req.URL)
		fmt.Printf("[DEBUG] Authentication token = %v\n", rdc.AuthToken)
	}

	// Perform the request
	res, err := rdc.client.Do(req)
	if err != nil {
		return nil, -1, err
	}

	// Read the response
	resData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, -1, err
	}

	if debug {
		fmt.Printf("[DEBUG] Response code = %v\n", res.StatusCode)
		fmt.Printf("[DEBUG] Response body = %v\n", string(resData))
	}
	// This is just for development purposes.
	// It will genenerate a file with the XML response
	// to use afterwards to generate the Go lang structs.
	// Make sure 'genXML' is always 'false' by default!
	if genXML {
		ioutil.WriteFile("aux.xml", resData, 0600)
	}

	res.Body.Close()

	return resData, res.StatusCode, nil
}
