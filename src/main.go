package main

import (
	"flag"
	"golang-artifacts-syncher/src/global_structs"
	"golang-artifacts-syncher/src/helper_funcs"
	"golang-artifacts-syncher/src/mylog"
	"golang-artifacts-syncher/src/nuget_cli"
)

func initialize() {
	StartTimer()
	mylog.InitLogger()
	helper_funcs.InitVars()
	helper_funcs.PrintVars()
}

func validateEnvBeforeRun() {
	helper_funcs.ValidateEnvironment()
}

func parseArgs()() {
	mylog.Logger.Info("Parsing args")
	flag.Parse()
	helper_funcs.UpdateVars()
	helper_funcs.PrintVars()
}

func searchSrcServersForAvailableVersionsOfSpecifiedPackages() []global_structs.NugetPackageDetailsStruct {
	return nuget_cli.SearchSrcServersForAvailableVersionsOfSpecifiedPackages()
}

func searchDestServersForAvailableNugetPackages() map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct {
	return nuget_cli.SearchDestServersForAvailableNugetPackages()
}

func downloadFoundPackages(foundPackagesArr []global_structs.NugetPackageDetailsStruct) []global_structs.DownloadPackageDetailsStruct {
	return nuget_cli.DownloadFoundPackages(foundPackagesArr)
}

func uploadDownloadedPackages(downloadedPkgsArr []global_structs.DownloadPackageDetailsStruct) {
	nuget_cli.UploadDownloadedPackages(downloadedPkgsArr)
}

// Remove all packages that were downloaded but not uploaded - no need for them anymore
func deleteLocalUnuploadedPackages(uploadedPkgsArr []global_structs.DownloadPackageDetailsStruct) {
	helper_funcs.DeleteLocalUnuploadedPackages(uploadedPkgsArr)
}

// Remove all packages from remote that were downloaded but not uploaded - no need for them anymore
func deleteRemoteUnuploadedPackages(uploadedPkgsArr []global_structs.DownloadPackageDetailsStruct, destServersFoundPackagesMap map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct) {
	nuget_cli.DeleteRemoteUnuploadedPackages(uploadedPkgsArr, destServersFoundPackagesMap)
}

func StartTimer() {
	helper_funcs.StartTimer()
}

func Finish(uploadedPkgsArr []global_structs.DownloadPackageDetailsStruct) {
	mylog.Logger.Info("")
	mylog.Logger.Info("Summary:")
	mylog.Logger.Info(" Downloaded & Uploaded the following packages:")
	for i, pkgDetailsStruct := range uploadedPkgsArr {
		mylog.Logger.Infof(" %d) %s", i, pkgDetailsStruct.PkgDetailsStruct.HashCode())
	}
	mylog.Logger.Info("")
	mylog.Logger.Info("Finished")
	duration := helper_funcs.EndTimer()
	mylog.Logger.Infof("Time: %v", duration)
	mylog.Logger.Info("")
}



func main() {
	println("Program Started")
	initialize()
	parseArgs()
	validateEnvBeforeRun()
	filteredFoundPackagesDetailsList := searchSrcServersForAvailableVersionsOfSpecifiedPackages()
	downloadedPkgsArr := downloadFoundPackages(filteredFoundPackagesDetailsList)
	uploadDownloadedPackages(downloadedPkgsArr)
	deleteLocalUnuploadedPackages(downloadedPkgsArr)  // Remove all packages that were not uploaded (maybe from previous runs..)
	destServersFoundPackagesMap := searchDestServersForAvailableNugetPackages()
	deleteRemoteUnuploadedPackages(downloadedPkgsArr, destServersFoundPackagesMap)  // Remove all packages that were not uploaded (maybe from previous runs..)
	Finish(downloadedPkgsArr)
}
