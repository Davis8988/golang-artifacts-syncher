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
		downloadFilePath := filepath.Join(downloadPkgsDirPath, fileName) // downloadPkgsDirPath == global var
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

func uploadDownloadedPackage(uploadPkgStruct helpers.UploadPackageDetailsStruct) helpers.UploadPackageDetailsStruct {
	pkgPrintStr := fmt.Sprintf("%s==%s", uploadPkgStruct.PkgDetailsStruct.Name, uploadPkgStruct.PkgDetailsStruct.Version)
	pkgName := uploadPkgStruct.PkgDetailsStruct.Name
	pkgVersion := uploadPkgStruct.PkgDetailsStruct.Version

	// Check if package already exists. If so, then compare it's checksum and skip on matching
	for _, destServerUrl := range destServersUrlsArr {
		for _, repoName := range destReposNamesArr {
			destServerRepo := destServerUrl + "/" + repoName
			helpers.LogInfo.Printf("Checking if pkg: '%s' already exists at dest server: %s", pkgPrintStr, destServerRepo)
			checkDestServerPkgExistUrl := destServerRepo + "/" + "Packages(Id='" + pkgName + "',Version='" + pkgVersion + "')"
			httpRequestArgs := helpers.HttpRequestArgsStruct{
				UrlAddress: checkDestServerPkgExistUrl,
				HeadersMap: httpRequestHeadersMap,
				UserToUse:  destServersUserToUse,
				PassToUse:  destServersPassToUse,
				TimeoutSec: httpRequestTimeoutSecondsInt,
				Method:     "GET",
			}

			foundPackagesDetailsArr := helpers.SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs)
			helpers.LogInfo.Printf("Found: %s", foundPackagesDetailsArr)

			emptyNugetPackageDetailsStruct := helpers.NugetPackageDetailsStruct{}
			shouldCompareChecksum := true
			if len(foundPackagesDetailsArr) != 1 {
				helpers.LogInfo.Printf("Found multiple or no packages: \"%d\" - Should be only 1. Skipping checksum comparison. Continuing with the upload..", len(foundPackagesDetailsArr))
				shouldCompareChecksum = false
			} else if len(foundPackagesDetailsArr) == 1 && foundPackagesDetailsArr[0] == emptyNugetPackageDetailsStruct {
				helpers.LogInfo.Print("No package found. Continuing with the upload..")
				shouldCompareChecksum = false
			}
			
			if shouldCompareChecksum {
				// Check the checksum:
				helpers.LogInfo.Printf("Comparing found package's checksum to know if should upload to: %s or not", destServerRepo)
				foundPackageChecksum := foundPackagesDetailsArr[0].Checksum
				fileToUploadChecksum := uploadPkgStruct.UploadFileChecksum
				if foundPackageChecksum == fileToUploadChecksum {
				fileName := filepath.Base(uploadPkgStruct.UploadFilePath)
				helpers.LogWarning.Printf("Checksum match: upload target file already exists in dest server: '%s' \n"+
					"Skipping upload of pkg: \"%s\"", destServerRepo, fileName)
				return uploadPkgStruct
				}
			}
			
			if len(destServerRepo) > 1 {
				lastChar := destServerRepo[len(destServerRepo)-1:]
				helpers.LogInfo.Printf("Adding '/' char to dest server repo url: \"%s\"", destServerRepo)
				if lastChar != "/" {destServerRepo += "/"}
			}
			httpRequestArgs.UrlAddress = destServerRepo
			// Upload the package file
			helpers.UploadPkg(uploadPkgStruct, httpRequestArgs)
		}
	}

	return uploadPkgStruct
}

func uploadDownloadedPackages(downloadedPkgsArr []helpers.DownloadPackageDetailsStruct) {
	helpers.LogInfo.Printf("Uploading %d downloaded packages to servers: %v", len(downloadedPkgsArr), destServersUrlsArr)
	if len(destServersUrlsArr) == 0 {
		helpers.LogWarning.Printf("No servers to upload to were given - skipping uploading of: %d packages", len(downloadedPkgsArr))
		return
	}
	for _, downloadedPkgStruct := range downloadedPkgsArr {
		uploadDownloadedPackage(helpers.UploadPackageDetailsStruct{
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
