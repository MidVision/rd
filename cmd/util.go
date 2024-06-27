package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func encodeToken(username, password string) string {
	tokenStr := username + ":" + password
	msg := []byte(tokenStr)
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(msg)))
	base64.StdEncoding.Encode(encoded, msg)
	return string(encoded)
}

func decodeToken(tokenStr string) (string, string, error) {
	decLen := base64.StdEncoding.DecodedLen(len(tokenStr))
	decoded := make([]byte, decLen)
	tokenByte := []byte(tokenStr)
	n, err := base64.StdEncoding.Decode(decoded, tokenByte)
	if err != nil {
		return "", "", err
	}
	if n > decLen {
		return "", "", fmt.Errorf("Something went wrong decoding the authentication token.")
	}
	arr := strings.SplitN(string(decoded), ":", 2)
	if len(arr) != 2 {
		return "", "", fmt.Errorf("Invalid authentication token.")
	}
	username := arr[0]
	password := strings.Trim(arr[1], "\x00")
	return username, password, nil
}

func getHome() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return home
}

func call(method string, reqUrlStr string, bodyContent interface{}, header map[string]string) ([]byte, int, error) {

	httpClient := &http.Client{Timeout: 5 * time.Second}

	// Parse the URL for the request
	reqUrl, err := url.Parse(reqUrlStr)
	if err != nil {
		if debug {
			fmt.Printf("[ERROR] Parse URL: %v\n", err)
		}
		return nil, -1, err
	}

	// Prepare the body for the request
	var reqData []byte
	if bodyContent != nil {
		reqData, err = json.Marshal(bodyContent)
		if err != nil {
			if debug {
				fmt.Printf("[ERROR] Marshal body: %v\n", err)
			}
			return nil, -1, err
		}
	}

	if debug {
		fmt.Printf("[DEBUG] Request body = %v\n", string(reqData))
	}

	// Create the HTTP request
	req, err := http.NewRequest(method, reqUrl.String(), bytes.NewBuffer(reqData))
	if err != nil {
		return nil, -1, err
	}

	// Add header to the request
	for key, value := range header {
		req.Header.Add(key, value)
	}

	if debug {
		fmt.Printf("[DEBUG] Request URL = %v\n", req.URL)
		fmt.Printf("[DEBUG] Request method = %v\n", req.Method)
		fmt.Printf("[DEBUG] Request header = %v\n", req.Header)
	}

	// Perform the request
	res, err := httpClient.Do(req)
	if err != nil {
		if debug {
			fmt.Printf("[ERROR] HTTP Client: %v\n", err)
		}
		return nil, -1, err
	}

	// Read the response
	resData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		if debug {
			fmt.Printf("[ERROR] Read response: %v\n", err)
		}
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
