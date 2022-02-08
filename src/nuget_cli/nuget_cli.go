package nuget_cli

import (
	"golang-artifacts-syncher/src/global_structs"
	"golang-artifacts-syncher/src/helpers"
	"sync"
)

func PushNugetPackage() {

}

func DownloadNugetPackage() {

}

func SearchForAvailableNugetPackages() []global_structs.NugetPackageDetailsStruct {
	var totalFoundPackagesDetailsArr []global_structs.NugetPackageDetailsStruct
	searchUrlsArr := helpers.PrepareSrcSearchAllPkgsVersionsUrlsArray()

	wg := sync.WaitGroup{}

	// Ensure all routines finish before returning
	defer wg.Wait()

	if len(searchUrlsArr) > 0 {
		helpers.LogInfo.Printf("Checking %d src URL addresses for pkgs versions", len(searchUrlsArr))
		for _, urlToCheck := range searchUrlsArr {
			wg.Add(1)
			go func(urlToCheck string) {
				defer wg.Done()
				httpRequestArgs := global_structs.HttpRequestArgsStruct{
					UrlAddress: urlToCheck,
					HeadersMap: httpRequestHeadersMap,
					UserToUse:  srcServersUserToUse,
					PassToUse:  srcServersPassToUse,
					TimeoutSec: httpRequestTimeoutSecondsInt,
					Method:     "GET",
				}
				foundPackagesDetailsArr := SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs)
				foundPackagesDetailsArr = filterFoundPackagesByRequestedVersion(foundPackagesDetailsArr) // Filter by requested version - if any version is specified..
				Synched_AppendPkgDetailsObj(&totalFoundPackagesDetailsArr, foundPackagesDetailsArr)
			}(urlToCheck)
		}
	}
	wg.Wait()

	return totalFoundPackagesDetailsArr
}