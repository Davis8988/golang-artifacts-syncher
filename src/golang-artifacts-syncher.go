package main

import (
	"flag"
	"golang-artifacts-syncher/src/global_structs"
	"golang-artifacts-syncher/src/helper_funcs"
	"golang-artifacts-syncher/src/mylog"
	"golang-artifacts-syncher/src/nuget_cli"
)

var (    
    BuildVersion string = ""
)

func Start() {
	println(helper_funcs.Fmt_Sprintf("Version: %s", BuildVersion))
	println("Program Started")
}

func initialize() {
	StartTimer()
	mylog.InitLogger()
	helper_funcs.InitVars()
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

func downloadFoundPackages(foundPackagesArr []global_structs.NugetPackageDetailsStruct, destServersFoundPackagesMap map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct) []global_structs.DownloadPackageDetailsStruct {
	return nuget_cli.DownloadFoundPackages(foundPackagesArr, destServersFoundPackagesMap)
}

func uploadDownloadedPackages(downloadedPkgsArr []global_structs.DownloadPackageDetailsStruct) []global_structs.UploadPackageDetailsStruct {
	return nuget_cli.UploadDownloadedPackages(downloadedPkgsArr)
}

// Remove all packages that were downloaded but not uploaded - no need for them anymore
func deleteLocalUploadedPackages(uploadedPkgsArr []global_structs.UploadPackageDetailsStruct) {
	helper_funcs.DeleteLocalUploadedPackages(uploadedPkgsArr)
}

// Remove all packages from remote that were downloaded but not uploaded - no need for them anymore
func deleteRemoteUnuploadedPackages(uploadedPkgsArr []global_structs.NugetPackageDetailsStruct, destServersFoundPackagesMap map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct) {
	nuget_cli.DeleteRemoteUnuploadedPackages(uploadedPkgsArr, destServersFoundPackagesMap)
}

func StartTimer() {
	helper_funcs.StartTimer()
}

func Finish(filteredFoundPackagesDetailsList []global_structs.NugetPackageDetailsStruct, downloadedPkgsArr []global_structs.DownloadPackageDetailsStruct, uploadedPkgsArr []global_structs.UploadPackageDetailsStruct) {
	helper_funcs.PrintFinishSummary(filteredFoundPackagesDetailsList, downloadedPkgsArr, uploadedPkgsArr)
	mylog.Logger.Infof("Version: %s", BuildVersion)
}



func main() {
	Start()
	initialize()
	parseArgs()
	validateEnvBeforeRun()
	filteredFoundPackagesDetailsList := searchSrcServersForAvailableVersionsOfSpecifiedPackages()
	destServersFoundPackagesMap := searchDestServersForAvailableNugetPackages()
	downloadedPkgsArr := downloadFoundPackages(filteredFoundPackagesDetailsList, destServersFoundPackagesMap)
	uploadedPkgsArr := uploadDownloadedPackages(downloadedPkgsArr)
	deleteLocalUploadedPackages(uploadedPkgsArr)  // Remove all packages that were uploaded (maybe from previous runs..)
	deleteRemoteUnuploadedPackages(filteredFoundPackagesDetailsList, destServersFoundPackagesMap)  // Remove all packages that were not uploaded (maybe from previous runs..)
	Finish(filteredFoundPackagesDetailsList, downloadedPkgsArr, uploadedPkgsArr)
}
