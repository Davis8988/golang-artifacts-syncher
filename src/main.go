package main

import (
	"flag"
	"golang-artifacts-syncher/src/global_structs"
	"golang-artifacts-syncher/src/global_vars"
	"golang-artifacts-syncher/src/helper_funcs"
	"golang-artifacts-syncher/src/mylog"
	"golang-artifacts-syncher/src/nuget_cli"

	"path/filepath"
	"sync"
)

func initLogger() {
	mylog.InitLogger()
}

func initVars() {
	helper_funcs.InitVars()
}

func printVars() {
	helper_funcs.PrintVars()
}

func validateEnv() {
	helper_funcs.ValidateEnvironment()
}

func parseArgs() {
	mylog.Logger.Info("Parsing args")
	flag.Parse()
}

func updateVars() {
	helper_funcs.UpdateVars()
}

func searchAvailableVersionsOfSpecifiedPackages() []global_structs.NugetPackageDetailsStruct {
	return nuget_cli.SearchForAvailableNugetPackages()
}

func downloadSpecifiedPackages(foundPackagesArr []global_structs.NugetPackageDetailsStruct) []global_structs.DownloadPackageDetailsStruct {
	var totalDownloadedPackagesDetailsArr []global_structs.DownloadPackageDetailsStruct
	if len(foundPackagesArr) == 0 {return totalDownloadedPackagesDetailsArr}
	mylog.Logger.Infof("Downloading found %d packages simultaneously in groups of: %d", len(foundPackagesArr), global_vars.PackagesMaxConcurrentDownloadCount)

	wg := sync.WaitGroup{}
	// Ensure all routines finish before returning
	defer wg.Wait()
	concurrentCountGuard := make(chan int, global_vars.PackagesMaxConcurrentDownloadCount) // Set an array of size: 'global_vars.PackagesMaxConcurrentDownloadCount'

	for _, pkgDetailsStruct := range foundPackagesArr {
		if len(pkgDetailsStruct.Name) == 0 || len(pkgDetailsStruct.Version) == 0 {
			mylog.Logger.Info("Skipping downloading of an unnamed/unversioned pkg")
			continue
		}

		wg.Add(1)
		fileName := pkgDetailsStruct.Name + "." + pkgDetailsStruct.Version + ".nupkg"
		downloadFilePath := filepath.Join(global_vars.DownloadPkgsDirPath, fileName) // downloadPkgsDirPath == global var
		downloadPkgDetailsStruct := global_structs.DownloadPackageDetailsStruct{
			PkgDetailsStruct:         pkgDetailsStruct,
			DownloadFilePath:         downloadFilePath,
			DownloadFileChecksum:     helper_funcs.CalculateFileChecksum(downloadFilePath), // Can by empty if file doesn't exist yet
			DownloadFileChecksumType: "SHA512",                                        // Default checksum algorithm for Nuget pkgs
		}
		
		concurrentCountGuard <- 1; // Add 1 to concurrent threads count - Would block if array is filled. Can only be freed by thread executing: '<- concurrentCountGuard' below
		go func(downloadPkgDetailsStruct global_structs.DownloadPackageDetailsStruct) {
			defer wg.Done()
			nuget_cli.DownloadNugetPackage(downloadPkgDetailsStruct)
			helper_funcs.Synched_AppendDownloadedPkgDetailsObj(&totalDownloadedPackagesDetailsArr, downloadPkgDetailsStruct)
			<- concurrentCountGuard  // Remove 1 from 'concurrentCountGuard'
		}(downloadPkgDetailsStruct)
	}
	wg.Wait()

	return totalDownloadedPackagesDetailsArr
}

func uploadDownloadedPackages(downloadedPkgsArr []global_structs.DownloadPackageDetailsStruct) {
	mylog.Logger.Infof("Uploading %d downloaded packages to servers: %v", len(downloadedPkgsArr), global_vars.DestServersUrlsArr)
	if len(global_vars.DestServersUrlsArr) == 0 {
		mylog.Logger.Warnf("No servers to upload to were given - skipping uploading of: %d packages", len(downloadedPkgsArr))
		return
	}
	for _, downloadedPkgStruct := range downloadedPkgsArr {
		nuget_cli.UploadDownloadedPackage(global_structs.UploadPackageDetailsStruct{
			PkgDetailsStruct:       downloadedPkgStruct.PkgDetailsStruct,
			UploadFilePath:         downloadedPkgStruct.DownloadFilePath,
			UploadFileChecksum:     downloadedPkgStruct.DownloadFileChecksum,
			UploadFileChecksumType: downloadedPkgStruct.DownloadFileChecksumType,
		})
	}
}

func StartTimer() {
	helper_funcs.StartTimer()
}

func Finish() {
	mylog.Logger.Info("Finished")
	mylog.Logger.Info("")
	duration := helper_funcs.EndTimer()
	mylog.Logger.Infof("Time: %v", duration)
	mylog.Logger.Info("")
}

func main() {
	println("Program Started")
	StartTimer()
	initLogger()
	initVars()
	parseArgs()
	updateVars()
	printVars()
	validateEnv()
	foundPackagesArr := searchAvailableVersionsOfSpecifiedPackages()
	downloadedPkgsArr := downloadSpecifiedPackages(foundPackagesArr)
	uploadDownloadedPackages(downloadedPkgsArr)
	Finish()
}
