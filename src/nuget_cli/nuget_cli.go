package nuget_cli

import (
	"golang-artifacts-syncher/src/nuget_packages_xml"
	"golang-artifacts-syncher/src/global_structs"
	"golang-artifacts-syncher/src/global_vars"
	"golang-artifacts-syncher/src/helper_funcs"
	"golang-artifacts-syncher/src/mylog"
	"regexp"
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

func ParseXmlDataToSinglePkgDetailsStruct(entryStruct nuget_packages_xml.SinglePackagesDetailsXmlStruct) global_structs.NugetPackageDetailsStruct {
    var pkgDetailsStruct global_structs.NugetPackageDetailsStruct
    pkgDetailsStruct.Checksum = entryStruct.Properties.PackageHash
    pkgDetailsStruct.ChecksumType = entryStruct.Properties.PackageHashAlgorithm
    pkgDetailsStruct.PkgDetailsUrl = entryStruct.ID
    pkgDetailsStruct.PkgFileUrl = entryStruct.Content.Src
    pkgDetailsStruct.Name = ""
    pkgDetailsStruct.Version = ""
    parsedNameAndVersionArr := ParsePkgNameAndVersionFromFileURL(pkgDetailsStruct.PkgDetailsUrl)
    if parsedNameAndVersionArr != nil {
        pkgDetailsStruct.Name = parsedNameAndVersionArr[0]
        pkgDetailsStruct.Version = parsedNameAndVersionArr[1]
    }
    return pkgDetailsStruct
}

func ParsePkgNameAndVersionFromFileURL(pkgDetailsUrl string) [] string {
    mylog.LogDebug.Printf("Parsing URL for Name & Version: \"%s\"", pkgDetailsUrl)
    re := regexp.MustCompile("'(.*?)'")  // Find values in between quotes
    resultArr := re.FindAllString(pkgDetailsUrl, -1)  // -1 = find ALL available matches
    if len(resultArr) != 2 {
        mylog.LogError.Printf("Failed to parse URL for pkg Name & Version:  \"%s\"", pkgDetailsUrl)
        mylog.LogError.Printf("Found regex result count is: %d different from 2", len(resultArr))
        return nil
    }
    // Trim
    for i, value := range resultArr {resultArr[i] = helper_funcs.TrimQuotes(value)}
    return resultArr
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
	parsedPackagesDetailsArr := make([] global_structs.NugetPackageDetailsStruct, 0)
	skipGroupCount := global_vars.SearchPackagesUrlSkipGroupCount;
	origUrlAddr := httpRequestArgs.UrlAddress;
	currentSkipValue := 0;
	foundPackagesCount := skipGroupCount + 1;  // Start with dummy found packages of more than group count: skipGroupCount - Meaning there are more packages to search..
	mylog.LogDebug.Printf("Attempting to query for all packages in groups of: %d", skipGroupCount)
	for foundPackagesCount > skipGroupCount {
		httpRequestArgs.UrlAddress = helper_funcs.FmtSprintf("%s&$skip=%d&$top=%d", origUrlAddr, currentSkipValue, skipGroupCount)  // Adding &$skip=%d&$top=%d  to url
		responseBody := helper_funcs.MakeHttpRequest(httpRequestArgs)
		if len(responseBody) == 0 {return [] global_structs.NugetPackageDetailsStruct {}}
		currentParsedPackagesDetailsArr := ParseHttpRequestResponseForPackagesVersions(responseBody)
		foundPackagesCount = len(currentParsedPackagesDetailsArr);
		mylog.LogDebug.Printf("Current found packages count: %d", foundPackagesCount)
		parsedPackagesDetailsArr = append(parsedPackagesDetailsArr, currentParsedPackagesDetailsArr...)  // Add 2 slices
	}

	
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

func UploadDownloadedPackage(uploadPkgStruct global_structs.UploadPackageDetailsStruct) global_structs.UploadPackageDetailsStruct {
	pkgPrintStr := helper_funcs.FmtSprintf("%s==%s", uploadPkgStruct.PkgDetailsStruct.Name, uploadPkgStruct.PkgDetailsStruct.Version)
	pkgName := uploadPkgStruct.PkgDetailsStruct.Name
	pkgVersion := uploadPkgStruct.PkgDetailsStruct.Version

	// Check if package already exists. If so, then compare it's checksum and skip on matching
	for _, destServerUrl := range global_vars.DestServersUrlsArr {
		for _, repoName := range global_vars.DestReposNamesArr {
			destServerRepo := destServerUrl + "/" + repoName
			mylog.LogInfo.Printf("Checking if pkg: '%s' already exists at dest server: %s", pkgPrintStr, destServerRepo)
			checkDestServerPkgExistUrl := destServerRepo + "/" + "Packages(Id='" + pkgName + "',Version='" + pkgVersion + "')"
			httpRequestArgs := global_structs.HttpRequestArgsStruct{
				UrlAddress: checkDestServerPkgExistUrl,
				HeadersMap: global_vars.HttpRequestHeadersMap,
				UserToUse:  global_vars.DestServersUserToUse,
				PassToUse:  global_vars.DestServersPassToUse,
				TimeoutSec: global_vars.HttpRequestTimeoutSecondsInt,
				Method:     "GET",
			}

			foundPackagesDetailsArr := SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs)
			mylog.LogInfo.Printf("Found: %s", foundPackagesDetailsArr)

			emptyNugetPackageDetailsStruct := global_structs.NugetPackageDetailsStruct{}
			shouldCompareChecksum := true
			if len(foundPackagesDetailsArr) != 1 {
				mylog.LogInfo.Printf("Found multiple or no packages: \"%d\" - Should be only 1. Skipping checksum comparison. Continuing with the upload..", len(foundPackagesDetailsArr))
				shouldCompareChecksum = false
			} else if len(foundPackagesDetailsArr) == 1 && foundPackagesDetailsArr[0] == emptyNugetPackageDetailsStruct {
				mylog.LogInfo.Print("No package found. Continuing with the upload..")
				shouldCompareChecksum = false
			}
			
			if shouldCompareChecksum {
				// Check the checksum:
				mylog.LogInfo.Printf("Comparing found package's checksum to know if should upload to: %s or not", destServerRepo)
				foundPackageChecksum := foundPackagesDetailsArr[0].Checksum
				fileToUploadChecksum := uploadPkgStruct.UploadFileChecksum
				if foundPackageChecksum == fileToUploadChecksum {
				fileName := helper_funcs.GetFileNameFromPath(uploadPkgStruct.UploadFilePath)
				mylog.LogWarning.Printf("Checksum match: upload target file already exists in dest server: '%s' \n"+
					"Skipping upload of pkg: \"%s\"", destServerRepo, fileName)
				return uploadPkgStruct
				}
			}
			
			if len(destServerRepo) > 1 {
				lastChar := destServerRepo[len(destServerRepo)-1:]
				mylog.LogInfo.Printf("Adding '/' char to dest server repo url: \"%s\"", destServerRepo)
				if lastChar != "/" {destServerRepo += "/"}
			}
			httpRequestArgs.UrlAddress = destServerRepo
			// Upload the package file
			UploadPkg(uploadPkgStruct, httpRequestArgs)
		}
	}

	return uploadPkgStruct
}

func UploadPkg(uploadPkgStruct global_structs.UploadPackageDetailsStruct, httpRequestArgsStruct global_structs.HttpRequestArgsStruct) {
    pkgPrintStr := helper_funcs.FmtSprintf("%s==%s", uploadPkgStruct.PkgDetailsStruct.Name, uploadPkgStruct.PkgDetailsStruct.Version)
	mylog.LogInfo.Printf("Uploading package: \"%s\" from: %s", pkgPrintStr, uploadPkgStruct.UploadFilePath)
    httpRequestArgsStruct.Method = "PUT"
    httpRequestArgsStruct.UploadFilePath = uploadPkgStruct.UploadFilePath
    helper_funcs.MakeHttpRequest(httpRequestArgsStruct)

}