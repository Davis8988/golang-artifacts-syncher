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

func searchAvailableVersionsOfSpecifiedPackages() []helpers.NugetPackageDetailsStruct {
	return helpers.SearchForAvailableNugetPackages()
}

func downloadSpecifiedPackages(foundPackagesArr []helpers.NugetPackageDetailsStruct) []helpers.DownloadPackageDetailsStruct {
	helpers.LogInfo.Printf("Downloading found %d packages", len(foundPackagesArr))
	var totalDownloadedPackagesDetailsArr []helpers.DownloadPackageDetailsStruct

	wg := sync.WaitGroup{}
	// Ensure all routines finish before returning
	defer wg.Wait()

	for _, pkgDetailsStruct := range foundPackagesArr {
		if len(pkgDetailsStruct.Name) == 0 || len(pkgDetailsStruct.Version) == 0 {
			helpers.LogInfo.Print("Skipping downloading of an unnamed/unversioned pkg")
			continue
		}

		wg.Add(1)
		fileName := pkgDetailsStruct.Name + "." + pkgDetailsStruct.Version + ".nupkg"
		downloadFilePath := filepath.Join(helpers.downloadPkgsDirPath, fileName) // downloadPkgsDirPath == global var
		downloadPkgDetailsStruct := helpers.DownloadPackageDetailsStruct{
			PkgDetailsStruct:         pkgDetailsStruct,
			DownloadFilePath:         downloadFilePath,
			DownloadFileChecksum:     helpers.CalculateFileChecksum(downloadFilePath), // Can by empty if file doesn't exist yet
			DownloadFileChecksumType: "SHA512",                                        // Default checksum algorithm for Nuget pkgs
		}

		go func(downloadPkgDetailsStruct helpers.DownloadPackageDetailsStruct) {
			defer wg.Done()
			helpers.DownloadPkg(downloadPkgDetailsStruct)
			helpers.Synched_AppendDownloadedPkgDetailsObj(&totalDownloadedPackagesDetailsArr, downloadPkgDetailsStruct)
		}(downloadPkgDetailsStruct)
	}
	wg.Wait()

	return totalDownloadedPackagesDetailsArr
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
