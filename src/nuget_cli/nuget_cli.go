package nuget_cli

import (
	"golang-artifacts-syncher/src/nuget_packages_xml"
	"golang-artifacts-syncher/src/global_structs"
	"golang-artifacts-syncher/src/global_vars"
	"golang-artifacts-syncher/src/helper_funcs"
	"golang-artifacts-syncher/src/mylog"
	"sync"
)

func PushNugetPackage() {

}

func DownloadNugetPackage(downloadPkgDetailsStruct global_structs.DownloadPackageDetailsStruct) {
	mylog.LogInfo.Printf("Downloading package: %s==%s", downloadPkgDetailsStruct.PkgDetailsStruct.Name, downloadPkgDetailsStruct.PkgDetailsStruct.Version)
    fileUrl := downloadPkgDetailsStruct.PkgDetailsStruct.PkgFileUrl
    downloadFilePath := downloadPkgDetailsStruct.DownloadFilePath
    downloadFileChecksum := downloadPkgDetailsStruct.DownloadFileChecksum
    fileChecksum := downloadPkgDetailsStruct.PkgDetailsStruct.Checksum
    if fileChecksum == downloadFileChecksum {
        fileName := helper_funcs.GetFileName(downloadFilePath)
        mylog.LogWarning.Printf("Checksum match: download target file already exists. Skipping download of: \"%s\"", fileName)
        return
    }
    helper_funcs.MakeHttpRequest(
        global_structs.HttpRequestArgsStruct{
            UrlAddress: fileUrl,
            Method: "GET",
            DownloadFilePath: downloadFilePath,
        },
    )
}

func ParseHttpRequestResponseForPackagesVersions(responseBody string) [] global_structs.NugetPackageDetailsStruct {
    parsedPackagesVersionsArr := make([] global_structs.NugetPackageDetailsStruct, 0)
    mylog.LogInfo.Printf("Parsing http request response for packages details")
    parsedPackagesDetailsStruct := nuget_packages_xml.ParseMultipleNugetPackagesXmlData(responseBody)
    if len(parsedPackagesDetailsStruct.Entry) == 0 {  // If failed to parse entries, it might be only a single entry and in that case attempt to parse it
        entryStruct := nuget_packages_xml.ParseSingleNugetPackagesXmlData(responseBody)
        pkgDetailsStruct := ParseXmlDataToSinglePkgDetailsStruct(entryStruct)
        parsedPackagesVersionsArr = append(parsedPackagesVersionsArr, pkgDetailsStruct)
        return parsedPackagesVersionsArr  
    }
    for _, entryStruct := range parsedPackagesDetailsStruct.Entry {
        pkgDetailsStruct := ParseXmlDataToSinglePkgDetailsStruct(entryStruct)
        if len(pkgDetailsStruct.Name) == 0 || len(pkgDetailsStruct.Version) == 0 {continue}
        parsedPackagesVersionsArr = append(parsedPackagesVersionsArr, pkgDetailsStruct)
    }
    return parsedPackagesVersionsArr
}


func SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs global_structs.HttpRequestArgsStruct) [] global_structs.NugetPackageDetailsStruct {
    responseBody := helper_funcs.MakeHttpRequest(httpRequestArgs)
    if len(responseBody) == 0 {return [] global_structs.NugetPackageDetailsStruct {}}
    parsedPackagesDetailsArr := ParseHttpRequestResponseForPackagesVersions(responseBody)

    return parsedPackagesDetailsArr
}


func SearchForAvailableNugetPackages() []global_structs.NugetPackageDetailsStruct {
	var totalFoundPackagesDetailsArr []global_structs.NugetPackageDetailsStruct
	searchUrlsArr := helper_funcs.PrepareSrcSearchAllPkgsVersionsUrlsArray()

	wg := sync.WaitGroup{}

	// Ensure all routines finish before returning
	defer wg.Wait()

	if len(searchUrlsArr) > 0 {
		mylog.LogInfo.Printf("Checking %d src URL addresses for pkgs versions", len(searchUrlsArr))
		for _, urlToCheck := range searchUrlsArr {
			wg.Add(1)
			go func(urlToCheck string) {
				defer wg.Done()
				httpRequestArgs := global_structs.HttpRequestArgsStruct{
					UrlAddress: urlToCheck,
					HeadersMap: global_vars.HttpRequestHeadersMap,
					UserToUse:  global_vars.SrcServersUserToUse,
					PassToUse:  global_vars.SrcServersPassToUse,
					TimeoutSec: global_vars.HttpRequestTimeoutSecondsInt,
					Method:     "GET",
				}
				foundPackagesDetailsArr := SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs)
				foundPackagesDetailsArr = helper_funcs.FilterFoundPackagesByRequestedVersion(foundPackagesDetailsArr) // Filter by requested version - if any version is specified..
				helper_funcs.Synched_AppendPkgDetailsObj(&totalFoundPackagesDetailsArr, foundPackagesDetailsArr)
			}(urlToCheck)
		}
	}
	wg.Wait()

	return totalFoundPackagesDetailsArr
}