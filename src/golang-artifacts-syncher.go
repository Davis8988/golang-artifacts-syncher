package main

import (
	"flag"
	"golang-artifacts-syncher/src/global_structs"
	"golang-artifacts-syncher/src/helper_funcs"
	"golang-artifacts-syncher/src/mylog"
	"golang-artifacts-syncher/src/nuget_cli"
)

var (    
    versionArg *bool
	AppVersion string = ""
)

func start() {
	println(helper_funcs.Fmt_Sprintf("Syncher Version: %s", AppVersion))
	println("Program Started")
}

func initSuccessIndicatorFilePath() {
	helper_funcs.InitSuccessIndicatorFilePath()
}

func initialize() {
	StartTimer()
	mylog.InitLogger()
	helper_funcs.InitVars()
	helper_funcs.UpdateVars()
	helper_funcs.PrintVars()
	helper_funcs.CreateRequiredFiles()
}

func validateEnvBeforeRun() {
	helper_funcs.ValidateEnvironment()
}

func parseArgs() {
	versionArg = flag.Bool("version", false, "prints syncher version")
	flag.Parse()
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

// Remove all checksum files of packages that weren't even found - not relevant anymore..
func deleteLocalUnfoundPackagesChecksumFiles(foundPackagesArr []global_structs.NugetPackageDetailsStruct) {
	helper_funcs.DeleteLocalUnfoundPackagesChecksumFiles(foundPackagesArr)
}

// Remove all packages from remote that were downloaded but not uploaded - no need for them anymore
func deleteRemoteUnuploadedPackages(uploadedPkgsArr []global_structs.NugetPackageDetailsStruct, destServersFoundPackagesMap map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct) {
	nuget_cli.DeleteRemoteUnuploadedPackages(uploadedPkgsArr, destServersFoundPackagesMap)
}

func StartTimer() {
	helper_funcs.StartTimer()
}

func printFinishRunInfo(filteredFoundPackagesDetailsList []global_structs.NugetPackageDetailsStruct, downloadedPkgsArr []global_structs.DownloadPackageDetailsStruct, uploadedPkgsArr []global_structs.UploadPackageDetailsStruct) {
	helper_funcs.PrintFinishSummary(filteredFoundPackagesDetailsList, downloadedPkgsArr, uploadedPkgsArr)
	mylog.Logger.Infof("Syncher Version: %s", AppVersion)
}

func finish() {
	helper_funcs.WriteSuccessIndicatorFile()  // Success finished run file indicator
}

func printVersionAndExit() {
	helper_funcs.Fmt_Print(AppVersion)
	finish()
	helper_funcs.Exit(0)
}

func runProgram() {
	initSuccessIndicatorFilePath()

	// Errors handling function
	defer helper_funcs.HandlePanicErrors()

	parseArgs()

	// Check if wanted only to print app version
	if *versionArg {printVersionAndExit()}

	// Start
	start()
	initialize()
	validateEnvBeforeRun()
	filteredFoundPackagesDetailsList := searchSrcServersForAvailableVersionsOfSpecifiedPackages()
	destServersFoundPackagesMap := searchDestServersForAvailableNugetPackages()
	downloadedPkgsArr := downloadFoundPackages(filteredFoundPackagesDetailsList, destServersFoundPackagesMap)
	uploadedPkgsArr := uploadDownloadedPackages(downloadedPkgsArr)
	deleteLocalUploadedPackages(uploadedPkgsArr)  // Remove all packages that were uploaded (maybe from previous runs..)
	deleteLocalUnfoundPackagesChecksumFiles(filteredFoundPackagesDetailsList)  // Remove all checksum files of packages that weren't even found - not relevant anymore..
	deleteRemoteUnuploadedPackages(filteredFoundPackagesDetailsList, destServersFoundPackagesMap)  // Remove all packages that were not uploaded (maybe from previous runs..)
	printFinishRunInfo(filteredFoundPackagesDetailsList, downloadedPkgsArr, uploadedPkgsArr)
	finish()
}


func main() {
	runProgram()
}
