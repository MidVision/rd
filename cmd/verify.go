// Copyright © 2024 Rafael Ruiz Palacios <support@midvision.com>

package cmd

import (
	"archive/zip"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	systemInfoFilename = "general-info.txt"
	propertiesFilename = "rapiddeploy.properties"
	logsFilename       = "logs.zip"
)

var outputFile string

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifies the RapidDeploy installation and retrieves useful information.",
	Long: `Verifies the RapidDeploy installation and retrieves the system information, configuration and 
logs of the server in a ZIP file for further investigation.`,
	Run: func(cmd *cobra.Command, args []string) {
		if quiet {
			os.Stdout = nil
		}
		retcode := 0
		defer func() {
			os.Exit(retcode)
		}()
		// Load the login session file - initialize the rdClient struct
		if err := rdClient.loadLoginFile(); err != nil {
			printStdError("\n%v\n\n", err)
			os.Exit(1)
		}

		/*************** Retrieve system info ***************/
		resData, _, _ := rdClient.call(http.MethodGet, "system/general-info", nil, "text/xml")
		systemInfoFilePath := filepath.Join(os.TempDir(), systemInfoFilename)
		err := os.WriteFile(systemInfoFilePath, resData, 0644)
		if err != nil {
			printStdError("\nUnable to create file: %s\n", systemInfoFilePath)
			printStdError("%v\n\n", err)
			os.Exit(1)
		}

		/*************** Retrive properties ***************/
		resData, _, _ = rdClient.call(http.MethodGet, "system/configuration", nil, "text/xml")
		propertiesFilePath := filepath.Join(os.TempDir(), propertiesFilename)
		err = os.WriteFile(propertiesFilePath, resData, 0644)
		if err != nil {
			printStdError("\nUnable to create file: %s\n", propertiesFilePath)
			printStdError("%v\n\n", err)
			os.Exit(1)
		}

		/*************** Retrieve application logs ***************/
		resData, _, _ = rdClient.call(http.MethodGet, "system/application-logs", nil, "application/zip")
		logsFilePath := filepath.Join(os.TempDir(), logsFilename)
		err = os.WriteFile(logsFilePath, resData, 0644)
		if err != nil {
			printStdError("\nUnable to create file: %s\n", logsFilePath)
			printStdError("%v\n\n", err)
			os.Exit(1)
		}

		/*************** Create ZIP file ***************/
		archiveName := getDatedFilename("rd-info", "zip")
		if outputFile != "" {
			archiveName = outputFile
		}
		archiveAbsPath, err := filepath.Abs(archiveName)
		if err != nil {
			printStdError("\n%v\n\n", err)
			retcode = 1
			return
		}
		if debug {
			fmt.Printf("[DEBUG] Creating ZIP archive: %v\n", archiveAbsPath)
		}
		archive, err := os.Create(archiveName)
		defer archive.Close()
		if err != nil {
			printStdError("\n%v\n\n", err)
			retcode = 1
			return
		}
		archiveWriter := zip.NewWriter(archive)
		defer archiveWriter.Close()

		// ===> Include the system info file
		sysInfoFile, err := os.Open(systemInfoFilePath)
		defer sysInfoFile.Close()
		if err != nil {
			printStdError("\n%v\n\n", err)
			retcode = 1
			return
		}
		if debug {
			fmt.Printf("[DEBUG] => Including the system info file into the archive: %v\n", sysInfoFile.Name())
		}
		sysInfoZipWriter, err := archiveWriter.Create(systemInfoFilename)
		if err != nil {
			printStdError("\n%v\n\n", err)
			retcode = 1
			return
		}
		if _, err := io.Copy(sysInfoZipWriter, sysInfoFile); err != nil {
			printStdError("\n%v\n\n", err)
			retcode = 1
			return
		}
		if debug {
			fmt.Printf("[DEBUG]    Removing the system info file: %v\n", systemInfoFilePath)
		}
		os.Remove(systemInfoFilePath)

		// ===> Include the properties file
		propsFile, err := os.Open(propertiesFilePath)
		defer propsFile.Close()
		if err != nil {
			printStdError("\n%v\n\n", err)
			retcode = 1
			return
		}
		if debug {
			fmt.Printf("[DEBUG] => Including the properties file into the archive: %v\n", propsFile.Name())
		}
		propsZipWriter, err := archiveWriter.Create(propertiesFilename)
		if err != nil {
			printStdError("\n%v\n\n", err)
			retcode = 1
			return
		}
		if _, err := io.Copy(propsZipWriter, propsFile); err != nil {
			printStdError("\n%v\n\n", err)
			retcode = 1
			return
		}
		if debug {
			fmt.Printf("[DEBUG]    Removing the properties file: %v\n", propertiesFilePath)
		}
		os.Remove(propertiesFilePath)

		// ===> Include the logs file
		zipReader, err := zip.OpenReader(logsFilePath)
		defer zipReader.Close()
		if err != nil {
			printStdError("\n%v\n\n", err)
			retcode = 1
			return
		}
		if debug {
			fmt.Printf("[DEBUG] Opening the logs ZIP file: %v\n", logsFilePath)
		}
		for _, zipItem := range zipReader.File {
			if debug {
				fmt.Printf("[DEBUG] => Including the logs file into the archive: %v\n", zipItem.Name)
			}
			zipItemReader, err := zipItem.Open()
			defer zipItemReader.Close()
			if err != nil {
				printStdError("\n%v\n\n", err)
				retcode = 1
				return
			}
			header, err := zip.FileInfoHeader(zipItem.FileInfo())
			if err != nil {
				printStdError("\n%v\n\n", err)
				retcode = 1
				return
			}
			header.Name = zipItem.Name
			header.Method = zip.Deflate
			targetItem, err := archiveWriter.CreateHeader(header)
			if err != nil {
				printStdError("\n%v\n\n", err)
				retcode = 1
				return
			}
			_, err = io.Copy(targetItem, zipItemReader)
			if err != nil {
				printStdError("\n%v\n\n", err)
				retcode = 1
				return
			}
		}
		if debug {
			fmt.Printf("[DEBUG]    Removing the logs ZIP file: %v\n", logsFilePath)
		}
		os.Remove(logsFilePath)

		// Show resulting ZIP file
		fmt.Println()
		fmt.Println("RapidDeploy information file: " + archiveAbsPath)
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().StringVarP(&outputFile, "output", "o", "", "The absolute or relative to current directory path for the output information ZIP file.")
}

func getDatedFilename(prefix, extension string) string {
	const layout = "20060102150405"
	t := time.Now()
	return prefix + "-" + t.Format(layout) + "." + extension
}
