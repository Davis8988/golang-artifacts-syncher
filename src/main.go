package main

import (
	"flag"
	"golang-artifacts-syncher/src/global_structs"
	"golang-artifacts-syncher/src/global_vars"
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


func searchAvailableVersionsOfSpecifiedPackages() []global_structs.NugetPackageDetailsStruct {
	return nuget_cli.SearchForAvailableNugetPackages()
}

func sortFoundNugetPackagesArray(nugetPackagesDetailsStructArr []global_structs.NugetPackageDetailsStruct) {
	helper_funcs.SortNugetPackageDetailsStructArr(nugetPackagesDetailsStructArr)
}

func FilterFoundPackages(nugetPackagesDetailsStructArr []global_structs.NugetPackageDetailsStruct) []global_structs.NugetPackageDetailsStruct {
	foundPackagesDetailsArr := helper_funcs.FilterFoundPackagesByRequestedVersion(nugetPackagesDetailsStructArr) // Filter by requested version - if any version is specified..
	return helper_funcs.FilterLastNPackages(foundPackagesDetailsArr, global_vars.PackagesDownloadLimitCount)
}

func downloadFoundPackages(foundPackagesArr []global_structs.NugetPackageDetailsStruct) []global_structs.DownloadPackageDetailsStruct {
	return nuget_cli.DownloadFoundPackages(foundPackagesArr)
}

func uploadDownloadedPackages(downloadedPkgsArr []global_structs.DownloadPackageDetailsStruct) {
	nuget_cli.UploadDownloadedPackages(downloadedPkgsArr)
}

// Remove all packages that were downloaded but not uploaded - no need for them anymore
func deleteUnuploadedPackages(uploadedPkgsArr []global_structs.DownloadPackageDetailsStruct) {
	nuget_cli.UploadDownloadedPackages(uploadedPkgsArr)
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
	initialize()
	parseArgs()
	validateEnvBeforeRun()
	totalFoundPackagesArr := searchAvailableVersionsOfSpecifiedPackages()
	mylog.Logger.Infof("Total found packages count: %d", len(totalFoundPackagesArr))
	sortFoundNugetPackagesArray(totalFoundPackagesArr)
	targetPackagesArr := FilterFoundPackages(totalFoundPackagesArr)
	downloadedPkgsArr := downloadFoundPackages(targetPackagesArr)
	uploadDownloadedPackages(downloadedPkgsArr)
	deleteUnuploadedPackages(downloadedPkgsArr)  // Remove all packages that were not uploaded (maybe from previous runs..)
	Finish()
}
