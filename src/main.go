package main

import (
	"flag"
	"fmt"
	"golang-artifacts-syncher/src/helpers"
	// "golang-artifacts-syncher/src/nuget_cli"
	"path/filepath"
	"strings"
	"sync"
)


func initVars() {
	helpers.Init()
}

func printVars() {
	helpers.PrintVars()
}

func validateEnv() {
	helpers.ValidateEnvironment()
}

func parseArgs() {
	helpers.LogInfo.Print("Parsing args")
	flag.Parse()
}

func updateVars() {
	helpers.UpdateVars()
}

func searchAvailableVersionsOfSpecifiedPackages() {
	helpers.SearchForAvailableNugetPackages()
}


func uploadDownloadedPackages(downloadedPkgsArr []helpers.DownloadPackageDetailsStruct) {
	helpers.LogInfo.Printf("Uploading %d downloaded packages to servers: %v", len(downloadedPkgsArr), helpers.DestServersUrlsArr)
	if len(helpers.DestServersUrlsArr) == 0 {
		helpers.LogWarning.Printf("No servers to upload to were given - skipping uploading of: %d packages", len(downloadedPkgsArr))
		return
	}
	for _, downloadedPkgStruct := range downloadedPkgsArr {
		helpers.UploadDownloadedPackage(helpers.UploadPackageDetailsStruct{
			PkgDetailsStruct:       downloadedPkgStruct.PkgDetailsStruct,
			UploadFilePath:         downloadedPkgStruct.DownloadFilePath,
			UploadFileChecksum:     downloadedPkgStruct.DownloadFileChecksum,
			UploadFileChecksumType: downloadedPkgStruct.DownloadFileChecksumType,
		})
	}
}

func main() {
	helpers.LogInfo.Print("Started")
	initVars()
	parseArgs()
	updateVars()
	printVars()
	validateEnv()
	foundPackagesArr := searchAvailableVersionsOfSpecifiedPackages()
	downloadedPkgsArr := downloadSpecifiedPackages(foundPackagesArr)
	uploadDownloadedPackages(downloadedPkgsArr)
	helpers.LogInfo.Print("Finished")
}
