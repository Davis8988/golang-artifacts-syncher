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
	return nuget_cli.DownloadSpecifiedPackages(foundPackagesArr)
}

func uploadDownloadedPackages(downloadedPkgsArr []global_structs.DownloadPackageDetailsStruct) {
	if len(downloadedPkgsArr) == 0 {
		mylog.Logger.Warnf("No packages to upload given")
		return
	}
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
