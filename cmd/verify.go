// Copyright © 2024 Rafael Ruiz Palacios <support@midvision.com>

// TODO: provide output path as a flag.

package cmd

import (
	"archive/zip"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	systemInfoFilename = "general-info.txt"
	propertiesFilename = "rapiddeploy.properties"
	logsFilename       = "logs.zip"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifies the RapidDeploy installation and retrieves useful information.",
	Long: `Verifies the RapidDeploy installation and retrieves the system information, configuration and 
logs of the server in a ZIP file for further investigation.`,
	// TODO
	//The -mvcloud option retrieves also information about the
	//MidVision Cloud product installed along with RapidDeploy.`,
	Run: func(cmd *cobra.Command, args []string) {
		retcode := 0
		defer func() {
			os.Exit(retcode)
		}()
		fmt.Println()
		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			fmt.Println(err.Error())
			retcode = 1
			return
		}

		/*************** Retrieve system info ***************/
		resData, statusCode, err := rdClient.call("GET", "system/general-info", nil, "text/xml")
		if err != nil {
			fmt.Printf("Unable to connect to server '%s'.\n", rdClient.BaseUrl)
			fmt.Printf("%v\n\n", err.Error())
			retcode = 1
			return
		}
		if statusCode != 200 {
			fmt.Printf("Unable to connect to server '%s'.\n", rdClient.BaseUrl)
			fmt.Printf("Please, perform a new login before requesting any action.\n\n")
			retcode = 1
			return
		}
		err = os.WriteFile(os.TempDir()+systemInfoFilename, resData, 0644)
		if err != nil {
			fmt.Println("Unable to create file:", os.TempDir()+logsFilename)
			fmt.Printf("%v\n\n", err)
			retcode = 1
			return
		}

		/*************** Retrive properties ***************/
		resData, statusCode, err = rdClient.call("GET", "system/configuration", nil, "text/xml")
		if err != nil {
			fmt.Printf("Unable to connect to server '%s'.\n", rdClient.BaseUrl)
			fmt.Printf("%v\n\n", err.Error())
			retcode = 1
			return
		}
		if statusCode != 200 {
			fmt.Printf("Unable to connect to server '%s'.\n", rdClient.BaseUrl)
			fmt.Printf("Please, perform a new login before requesting any action.\n\n")
			retcode = 1
			return
		}
		err = os.WriteFile(os.TempDir()+propertiesFilename, resData, 0644)
		if err != nil {
			fmt.Println("Unable to create file:", os.TempDir()+logsFilename)
			fmt.Printf("%v\n\n", err)
			retcode = 1
			return
		}

		/*************** Retrieve application logs ***************/
		resData, statusCode, err = rdClient.call("GET", "system/application-logs", nil, "application/zip")
		if err != nil {
			fmt.Printf("Unable to connect to server '%s'.\n", rdClient.BaseUrl)
			fmt.Printf("%v\n\n", err.Error())
			retcode = 1
			return
		}
		if statusCode != 200 {
			fmt.Printf("Unable to connect to server '%s'.\n", rdClient.BaseUrl)
			fmt.Printf("Please, perform a new login before requesting any action.\n\n")
			retcode = 1
			return
		}
		err = os.WriteFile(os.TempDir()+logsFilename, resData, 0644)
		if err != nil {
			fmt.Println("Unable to create file:", os.TempDir()+logsFilename)
			fmt.Printf("%v\n\n", err)
			retcode = 1
			return
		}

		/*************** Create ZIP file ***************/
		archiveName := getDatedFilename("rd-info", "zip")
		archiveAbsPath, err := filepath.Abs(archiveName)
		if err != nil {
			fmt.Printf("%v\n\n", err)
			retcode = 1
			return
		}
		if debug {
			fmt.Printf("[DEBUG] Creating ZIP archive: %v\n", archiveAbsPath)
		}
		archive, err := os.Create(archiveName)
		defer archive.Close()
		if err != nil {
			fmt.Printf("%v\n\n", err)
			retcode = 1
			return
		}
		archiveWriter := zip.NewWriter(archive)
		defer archiveWriter.Close()

		// ===> Include the system info file
		sysInfoFile, err := os.Open(os.TempDir() + systemInfoFilename)
		defer sysInfoFile.Close()
		if err != nil {
			fmt.Printf("%v\n\n", err)
			retcode = 1
			return
		}
		if debug {
			fmt.Printf("[DEBUG] => Including the system info file into the archive: %v\n", sysInfoFile.Name())
		}
		sysInfoZipWriter, err := archiveWriter.Create(systemInfoFilename)
		if err != nil {
			fmt.Printf("%v\n\n", err)
			retcode = 1
			return
		}
		if _, err := io.Copy(sysInfoZipWriter, sysInfoFile); err != nil {
			fmt.Printf("%v\n\n", err)
			retcode = 1
			return
		}
		if debug {
			fmt.Printf("[DEBUG]    Removing the system info file: %v\n", sysInfoFile.Name())
		}
		os.Remove(os.TempDir() + systemInfoFilename)

		// ===> Include the properties file
		propsFile, err := os.Open(os.TempDir() + propertiesFilename)
		defer propsFile.Close()
		if err != nil {
			fmt.Printf("%v\n\n", err)
			retcode = 1
			return
		}
		if debug {
			fmt.Printf("[DEBUG] => Including the properties file into the archive: %v\n", propsFile.Name())
		}
		propsZipWriter, err := archiveWriter.Create(propertiesFilename)
		if err != nil {
			fmt.Printf("%v\n\n", err)
			retcode = 1
			return
		}
		if _, err := io.Copy(propsZipWriter, propsFile); err != nil {
			fmt.Printf("%v\n\n", err)
			retcode = 1
			return
		}
		if debug {
			fmt.Printf("[DEBUG]    Removing the properties file: %v\n", propsFile.Name())
		}
		os.Remove(os.TempDir() + propertiesFilename)

		// ===> Include the logs file
		zipReader, err := zip.OpenReader(os.TempDir() + logsFilename)
		defer zipReader.Close()
		if err != nil {
			fmt.Printf("%v\n\n", err)
			retcode = 1
			return
		}
		if debug {
			fmt.Printf("[DEBUG] Opening the logs ZIP file: %v\n", os.TempDir()+logsFilename)
		}
		for _, zipItem := range zipReader.File {
			if debug {
				fmt.Printf("[DEBUG] => Including the logs file into the archive: %v\n", zipItem.Name)
			}
			zipItemReader, err := zipItem.Open()
			defer zipItemReader.Close()
			if err != nil {
				fmt.Printf("%v\n\n", err)
				retcode = 1
				return
			}
			header, err := zip.FileInfoHeader(zipItem.FileInfo())
			if err != nil {
				fmt.Printf("%v\n\n", err)
				retcode = 1
				return
			}
			header.Name = zipItem.Name
			header.Method = zip.Deflate
			targetItem, err := archiveWriter.CreateHeader(header)
			if err != nil {
				fmt.Printf("%v\n\n", err)
				retcode = 1
				return
			}
			_, err = io.Copy(targetItem, zipItemReader)
			if err != nil {
				fmt.Printf("%v\n\n", err)
				retcode = 1
				return
			}
		}
		if debug {
			fmt.Printf("[DEBUG]    Removing the logs ZIP file: %v\n", os.TempDir()+logsFilename)
		}
		os.Remove(os.TempDir() + logsFilename)

		// Show resulting ZIP file
		fmt.Println("RapidDeploy information file: " + archiveAbsPath)
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(verifyCmd)
}

func getDatedFilename(prefix, extension string) string {
	const layout = "20060102150405"
	t := time.Now()
	return prefix + "-" + t.Format(layout) + "." + extension
}
