package nuget_cli

import (
	"golang-artifacts-syncher/src/global_structs"
	"golang-artifacts-syncher/src/helpers_funcs"
	"sync"
)

func PushNugetPackage() {

}

func DownloadNugetPackage() {

}

func SearchForAvailableNugetPackages() []global_structs.NugetPackageDetailsStruct {
	var totalFoundPackagesDetailsArr []global_structs.NugetPackageDetailsStruct
	searchUrlsArr := helpers_funcs.PrepareSrcSearchAllPkgsVersionsUrlsArray()

	wg := sync.WaitGroup{}

	// Ensure all routines finish before returning
	defer wg.Wait()

	if len(searchUrlsArr) > 0 {
		helpers_funcs.LogInfo.Printf("Checking %d src URL addresses for pkgs versions", len(searchUrlsArr))
		for _, urlToCheck := range searchUrlsArr {
			wg.Add(1)
			go func(urlToCheck string) {
				defer wg.Done()
				httpRequestArgs := global_structs.HttpRequestArgsStruct{
					UrlAddress: urlToCheck,
					HeadersMap: helpers_funcs.HttpRequestHeadersMap,
					UserToUse:  helpers_funcs.SrcServersUserToUse,
					PassToUse:  helpers_funcs.SrcServersPassToUse,
					TimeoutSec: helpers_funcs.HttpRequestTimeoutSecondsInt,
					Method:     "GET",
				}
				foundPackagesDetailsArr := helpers_funcs.SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs)
				foundPackagesDetailsArr = helpers_funcs.FilterFoundPackagesByRequestedVersion(foundPackagesDetailsArr) // Filter by requested version - if any version is specified..
				helpers_funcs.Synched_AppendPkgDetailsObj(&totalFoundPackagesDetailsArr, foundPackagesDetailsArr)
			}(urlToCheck)
		}
	}
	wg.Wait()

	return totalFoundPackagesDetailsArr
}