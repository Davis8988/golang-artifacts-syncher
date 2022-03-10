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


func searchAvailableVersionsOfSpecifiedPackages() []global_structs.NugetPackageDetailsStruct {
	return nuget_cli.SearchForAvailableNugetPackages()
}

func SortFoundNugetPackagesArray(nugetPackagesDetailsStructArr []global_structs.NugetPackageDetailsStruct) {
	mylog.Logger.Infof("Sorting found nuget packages array")
	helper_funcs.SortNugetPackageDetailsStructArr(nugetPackagesDetailsStructArr)
	mylog.Logger.Infof("Done")
}

func FilterFoundPackages(nugetPackagesDetailsStructArr []global_structs.NugetPackageDetailsStruct) {
	mylog.Logger.Infof("Filtering found nuget packages")
	helper_funcs.SortNugetPackageDetailsStructArr(nugetPackagesDetailsStructArr)
	mylog.Logger.Infof("Done")
}

func downloadFoundPackages(foundPackagesArr []global_structs.NugetPackageDetailsStruct) []global_structs.DownloadPackageDetailsStruct {
	return nuget_cli.DownloadFoundPackages(foundPackagesArr)
}

func uploadDownloadedPackages(downloadedPkgsArr []global_structs.DownloadPackageDetailsStruct) {
	nuget_cli.UploadDownloadedPackages(downloadedPkgsArr)
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
	foundPackagesArr := searchAvailableVersionsOfSpecifiedPackages()
	SortFoundNugetPackagesArray(foundPackagesArr)
	FilterFoundPackages(foundPackagesArr)
	downloadedPkgsArr := downloadFoundPackages(foundPackagesArr)
	uploadDownloadedPackages(downloadedPkgsArr)
	Finish()
}
