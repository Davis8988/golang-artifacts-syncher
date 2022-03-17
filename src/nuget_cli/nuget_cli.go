package nuget_cli

import (
	"golang-artifacts-syncher/src/global_structs"
	"golang-artifacts-syncher/src/global_vars"
	"golang-artifacts-syncher/src/helper_funcs"
	"golang-artifacts-syncher/src/mylog"
	"golang-artifacts-syncher/src/nuget_packages_xml"
	"regexp"
	"strings"
	"sync"
)

func PushNugetPackage() {

}

func DownloadNugetPackage(downloadPkgDetailsStruct global_structs.DownloadPackageDetailsStruct) {
	mylog.Logger.Infof("Downloading package: %s==%s", downloadPkgDetailsStruct.PkgDetailsStruct.Name, downloadPkgDetailsStruct.PkgDetailsStruct.Version)
    fileUrl := downloadPkgDetailsStruct.PkgDetailsStruct.PkgFileUrl
    downloadFilePath := downloadPkgDetailsStruct.DownloadFilePath
    downloadFileChecksum := downloadPkgDetailsStruct.DownloadFileChecksum
    fileChecksum := downloadPkgDetailsStruct.PkgDetailsStruct.Checksum
    if fileChecksum == downloadFileChecksum {
        fileName := helper_funcs.GetFileName(downloadFilePath)
        mylog.Logger.Warnf("Checksum match: download target file already exists. Skipping download of: \"%s\"", fileName)
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

func ParseXmlDataToSinglePkgDetailsStruct(entryStruct nuget_packages_xml.SinglePackagesDetailsXmlStruct) *global_structs.NugetPackageDetailsStruct {
    var pkgDetailsStruct global_structs.NugetPackageDetailsStruct
    pkgDetailsStruct.PkgDetailsUrl = entryStruct.ID
	parsedNameAndVersionArr := ParsePkgNameAndVersionFromFileURL(pkgDetailsStruct.PkgDetailsUrl)
	if parsedNameAndVersionArr == nil {return nil}
	pkgDetailsStruct.Name = parsedNameAndVersionArr[0]
	pkgDetailsStruct.Version = parsedNameAndVersionArr[1]
	pkgDetailsStruct.Checksum = entryStruct.Properties.PackageHash
    pkgDetailsStruct.ChecksumType = entryStruct.Properties.PackageHashAlgorithm
    pkgDetailsStruct.PkgFileUrl = entryStruct.Content.Src
    return &pkgDetailsStruct
}

func ParsePkgNameAndVersionFromFileURL(pkgDetailsUrl string) [] string {
	if len(pkgDetailsUrl) == 0 {return nil}
    mylog.Logger.Debugf("Parsing URL for Name & Version: \"%s\"", pkgDetailsUrl)
    re := regexp.MustCompile("'(.*?)'")  // Find values in between quotes
    resultArr := re.FindAllString(pkgDetailsUrl, -1)  // -1 = find ALL available matches
    if len(resultArr) != 2 {
        mylog.Logger.Errorf("\nFailed to parse URL for pkg Name & Version:  \"%s\"", pkgDetailsUrl)
        mylog.Logger.Errorf("Found regex result count is: %d different from 2\n", len(resultArr))
        return nil
    }

    for i, value := range resultArr {resultArr[i] = helper_funcs.TrimQuotes(value)}  // Trim
    return resultArr
}

func ParseHttpRequestResponseForPackagesVersions(responseBody string) [] global_structs.NugetPackageDetailsStruct {
    parsedPackagesVersionsArr := make([] global_structs.NugetPackageDetailsStruct, 0)
    mylog.Logger.Infof("Parsing http request response for packages details array")
    parsedPackagesDetailsStruct := nuget_packages_xml.ParseMultipleNugetPackagesXmlData(responseBody)
    if len(parsedPackagesDetailsStruct.Entry) == 0 {  // If failed to parse entries, it might be only a single entry and in that case attempt to parse it
        entryStruct := nuget_packages_xml.ParseSingleNugetPackagesXmlData(responseBody)
        pkgDetailsStructPtr := ParseXmlDataToSinglePkgDetailsStruct(entryStruct)
		if (pkgDetailsStructPtr == nil) {return parsedPackagesVersionsArr} // Failed to parse - return empty
		pkgDetailsStruct := *pkgDetailsStructPtr
        parsedPackagesVersionsArr = append(parsedPackagesVersionsArr, pkgDetailsStruct)
        return parsedPackagesVersionsArr  
    }
    for _, entryStruct := range parsedPackagesDetailsStruct.Entry {
        pkgDetailsStructPtr := ParseXmlDataToSinglePkgDetailsStruct(entryStruct)
        if (pkgDetailsStructPtr == nil) {continue}
		pkgDetailsStruct := *pkgDetailsStructPtr
        parsedPackagesVersionsArr = append(parsedPackagesVersionsArr, pkgDetailsStruct)
    }
    return parsedPackagesVersionsArr
}

func ParseHttpRequestResponseForPackagesVersions_ToMap(responseBody string) map[string] global_structs.NugetPackageDetailsStruct {
    parsedPackagesVersionsMap := map[string] global_structs.NugetPackageDetailsStruct {}
    mylog.Logger.Infof("Parsing http request response for packages details map")
    parsedPackagesDetailsStruct := nuget_packages_xml.ParseMultipleNugetPackagesXmlData(responseBody)
    if len(parsedPackagesDetailsStruct.Entry) == 0 {  // If failed to parse entries, it might be only a single entry and in that case attempt to parse it
        entryStruct := nuget_packages_xml.ParseSingleNugetPackagesXmlData(responseBody)
        pkgDetailsStructPtr := ParseXmlDataToSinglePkgDetailsStruct(entryStruct)
		if (pkgDetailsStructPtr == nil) {return parsedPackagesVersionsMap} // Failed to parse - return empty
		pkgDetailsStruct := *pkgDetailsStructPtr
		parsedPackagesVersionsMap[pkgDetailsStruct.HashCode()] = pkgDetailsStruct
        return parsedPackagesVersionsMap  
    }
    for _, entryStruct := range parsedPackagesDetailsStruct.Entry {
        pkgDetailsStructPtr := ParseXmlDataToSinglePkgDetailsStruct(entryStruct)
        if (pkgDetailsStructPtr == nil) {continue}
		pkgDetailsStruct := *pkgDetailsStructPtr
        parsedPackagesVersionsMap[pkgDetailsStruct.HashCode()] = pkgDetailsStruct
    }
    return parsedPackagesVersionsMap
}


func SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs global_structs.HttpRequestArgsStruct) [] global_structs.NugetPackageDetailsStruct {
	parsedPackagesDetailsArr := [] global_structs.NugetPackageDetailsStruct {}
	queryUrlAddr := httpRequestArgs.UrlAddress;
	skipGroupCount := global_vars.SearchPackagesUrlSkipGroupCount;
	currentSkipValue := 0;
	foundPackagesCount := skipGroupCount + 1;  // Start with dummy found packages of more than group count: skipGroupCount - Meaning there are more packages to search..
	
	mylog.Logger.Debugf("Searching for all packages: '%s' at: %s", queryUrlAddr)
	mylog.Logger.Debugf("Attempting to query for all packages in groups of: %d", skipGroupCount)
	for foundPackagesCount >= skipGroupCount { // <-- While there are may still packages to query for
		httpRequestArgs.UrlAddress = helper_funcs.Fmt_Sprintf("%s&$skip=%d&$top=%d", queryUrlAddr, currentSkipValue, skipGroupCount)  // Adding &$skip=%d&$top=%d  to url
		httpResponsePtr := helper_funcs.MakeHttpRequest(httpRequestArgs)
		if httpResponsePtr == nil {return parsedPackagesDetailsArr}
		httpResponse := *httpResponsePtr
		responseBody := httpResponse.BodyStr
		if len(responseBody) == 0 || httpResponse.StatusCode >= 400 {return parsedPackagesDetailsArr}
		currentParsedPackagesDetailsArr := ParseHttpRequestResponseForPackagesVersions(responseBody)
		foundPackagesCount = len(currentParsedPackagesDetailsArr);
		parsedPackagesDetailsArr = append(parsedPackagesDetailsArr, currentParsedPackagesDetailsArr...)  // Add 2 slices
		currentSkipValue += skipGroupCount;  // Skip another group for the next query
	}
	
    return parsedPackagesDetailsArr
}

func SearchPackagesAvailableVersionsByURLRequest_ToMap_NoSkipPkgs(httpRequestArgs global_structs.HttpRequestArgsStruct) map[string] global_structs.NugetPackageDetailsStruct {
	parsedPackagesDetailsMap := map[string] global_structs.NugetPackageDetailsStruct {}
	queryUrlAddr := httpRequestArgs.UrlAddress;
	
	mylog.Logger.Debugf("Searching for all packages: '%s' at: %s", queryUrlAddr)
	httpResponsePtr := helper_funcs.MakeHttpRequest(httpRequestArgs)
	if httpResponsePtr == nil {return parsedPackagesDetailsMap}
	httpResponse := *httpResponsePtr
	responseBody := httpResponse.BodyStr
	if len(responseBody) == 0 || httpResponse.StatusCode >= 400 {return parsedPackagesDetailsMap}
	parsedPackagesDetailsMap = ParseHttpRequestResponseForPackagesVersions_ToMap(responseBody)
	
    return parsedPackagesDetailsMap
}

func SearchSpecificPackageVersionByURLRequest(httpRequestArgs global_structs.HttpRequestArgsStruct) [] global_structs.NugetPackageDetailsStruct {
	httpResponsePtr := helper_funcs.MakeHttpRequest(httpRequestArgs)
    if httpResponsePtr == nil {return [] global_structs.NugetPackageDetailsStruct {}}
	httpResponse := *httpResponsePtr
	
	urlAddress := httpResponse.UrlAddress
	responseBody := httpResponse.BodyStr
	responseCode := httpResponse.StatusCode

	httpCode_NotFound := 404
	if (httpResponse.StatusCode == httpCode_NotFound) {return [] global_structs.NugetPackageDetailsStruct {}}

	if responseCode >= 400 {
        mylog.Logger.Errorf("\n%d  %s", responseCode, responseBody)
		mylog.Logger.Errorf("Returned code: %d. HTTP request failure: %s\n", responseCode, urlAddress)
		return [] global_structs.NugetPackageDetailsStruct {}
    }

	if (len(responseBody) == 0) {return [] global_structs.NugetPackageDetailsStruct {}}
    
    parsedPackagesDetailsArr := ParseHttpRequestResponseForPackagesVersions(responseBody)

    return parsedPackagesDetailsArr
}

func SearchSrcServersForAvailableVersionsOfSpecifiedPackages() []global_structs.NugetPackageDetailsStruct {
	var totalFoundPackagesDetailsArr []global_structs.NugetPackageDetailsStruct
	
	wg := sync.WaitGroup{}
	for _, pkgName := range global_vars.PackagesNamesArr {
		wg.Add(1)
        go func(packageNameToSearch string) {
			defer wg.Done()
			searchUrlsArr := helper_funcs.PrepareSrcSearchUrlsForPackageArray(packageNameToSearch)
			threadFoundPackagesDetailsMap := make(map[string] global_structs.NugetPackageDetailsStruct)
			for _, urlToCheck := range searchUrlsArr {
				getRequestArgs := global_structs.HttpRequestArgsStruct{
					UrlAddress: urlToCheck,  // Only the url is different each time
					HeadersMap: global_vars.HttpRequestHeadersMap,
					UserToUse:  global_vars.SrcServersUserToUse,
					PassToUse:  global_vars.SrcServersPassToUse,
					TimeoutSec: global_vars.HttpRequestTimeoutSecondsInt,
					Method:     "GET",
				}
				var foundPackagesDetailsArr []global_structs.NugetPackageDetailsStruct
				if (strings.Contains(urlToCheck, ",Version='")) {
					mylog.Logger.Debugf("Searching for specific package: '%s' version at: %s", getRequestArgs.UrlAddress)
					foundPackagesDetailsArr = SearchSpecificPackageVersionByURLRequest(getRequestArgs)
				} else {
					foundPackagesDetailsArr = SearchPackagesAvailableVersionsByURLRequest(getRequestArgs)
				}
				
				foundPackagesCount := len(foundPackagesDetailsArr)
				if foundPackagesCount == 0 {
					mylog.Logger.Warnf(" *** No packages '%s' found at server: %s", packageNameToSearch, urlToCheck)
				} else {
					mylog.Logger.Debugf("Found %d packages '%s'  at server: %s", foundPackagesCount, packageNameToSearch, urlToCheck)
				}
				helper_funcs.AppendPkgDetailsArrayToMap(threadFoundPackagesDetailsMap, foundPackagesDetailsArr)  // Append to existing map
			}
			
			threadFoundPackagesDetailsArr := helper_funcs.ConvertPkgDetailsMapToArray(threadFoundPackagesDetailsMap)
			
			// Sort
			helper_funcs.SortNugetPackageDetailsStructArr(threadFoundPackagesDetailsArr)

			// Filter total found pkgs count of package: ${packageNameToSearch}
			mylog.Logger.Debugf("Filtering thread found '%s' pkgs by requested versions", packageNameToSearch)
			threadFoundPackagesDetailsArr = helper_funcs.FilterFoundPackagesByRequestedVersion(threadFoundPackagesDetailsArr) // Filter by requested version - if any version is specified..
			mylog.Logger.Debugf("Keeping last: %d of found '%s' packages", global_vars.PackagesDownloadLimitCount, packageNameToSearch)
			threadFoundPackagesDetailsArr = helper_funcs.FilterLastNPackages(threadFoundPackagesDetailsArr, global_vars.PackagesDownloadLimitCount)
			mylog.Logger.Infof("Targeted %d of '%s' packages", len(threadFoundPackagesDetailsArr), packageNameToSearch)
			helper_funcs.Synched_JoinTwoPkgDetailsSlices(&totalFoundPackagesDetailsArr, threadFoundPackagesDetailsArr)
		} (pkgName)
	}

	mylog.Logger.Debug("Waiting for threads to finish searching for pkgs")
	wg.Wait()
	mylog.Logger.Infof("Total packages found count: %d", len(totalFoundPackagesDetailsArr))
	
	return totalFoundPackagesDetailsArr
}

// Dest servers search:
//  return map has the following structure:
//  destServer(str) -> pkgName(str) -> pkgVersion(str) -> pkgObj(Struct)
func SearchDestServersForAvailableNugetPackages() map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct {
	totalFoundDestServersPackagesDetailsMap := make(map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct)
	
	wg := sync.WaitGroup{}
	for _, destServerUrl := range global_vars.DestServersUrlsArr {
		wg.Add(1)
        go func(destServerUrlToSearch string) {
			defer wg.Done()
			threadFoundPackagesDetailsMap := make(map[string] map[string] global_structs.NugetPackageDetailsStruct)
			foundPackagesCount := 0
			for _, pkgName := range global_vars.PackagesNamesArr {
				urlToCheck := destServerUrlToSearch + "Packages()?$filter=tolower(Id)%20eq%20'"+pkgName+"'"
				getRequestArgs := global_structs.HttpRequestArgsStruct{
					UrlAddress: urlToCheck,  // Only the url is different each time
					HeadersMap: global_vars.HttpRequestHeadersMap,
					UserToUse:  global_vars.SrcServersUserToUse,
					PassToUse:  global_vars.SrcServersPassToUse,
					TimeoutSec: global_vars.HttpRequestTimeoutSecondsInt,
					Method:     "GET",
				}
				
				foundPackagesDetailsMap := SearchPackagesAvailableVersionsByURLRequest_ToMap_NoSkipPkgs(getRequestArgs)
				
				currentFoundPackagesCount := len(foundPackagesDetailsMap)
				if currentFoundPackagesCount == 0 {
					mylog.Logger.Warnf(" *** No packages '%s' found at dest server: %s", pkgName, urlToCheck)
				} else {
					mylog.Logger.Debugf("Found %d packages '%s'  at dest server: %s", currentFoundPackagesCount, pkgName, urlToCheck)
				}
				foundPackagesCount += currentFoundPackagesCount
				threadFoundPackagesDetailsMap[pkgName] = foundPackagesDetailsMap  // Append to existing map
			}
			
			mylog.Logger.Infof("Found total %d of '%s' packages in dest server: %s", foundPackagesCount, global_vars.PackagesNamesStr, destServerUrlToSearch)
			helper_funcs.Synched_AddPkgDetailsStructMapToMap(totalFoundDestServersPackagesDetailsMap, destServerUrlToSearch, threadFoundPackagesDetailsMap)
		} (destServerUrl)
	}

	mylog.Logger.Debug("Waiting for threads to finish searching for pkgs")
	wg.Wait()
	pkgsCount := 0
	for _, destServerFoundPkgsMap := range totalFoundDestServersPackagesDetailsMap {
		for _, foundPkgsMap := range destServerFoundPkgsMap {
			pkgsCount += len(foundPkgsMap)
		}
	}
	mylog.Logger.Infof("Total of %d packages found at dest servers: %s", pkgsCount, global_vars.DestServersUrlsArr)
	
	return totalFoundDestServersPackagesDetailsMap
}


func DownloadFoundPackages(foundPackagesArr []global_structs.NugetPackageDetailsStruct) []global_structs.DownloadPackageDetailsStruct {
	foundPackagesCount := len(foundPackagesArr)
	if foundPackagesCount == 0 {
		mylog.Logger.Warn("No packages to download found")
		return []global_structs.DownloadPackageDetailsStruct {}
	}
	totalDownloadedPackagesDetailsArr := make([]global_structs.DownloadPackageDetailsStruct, 0, foundPackagesCount)
	mylog.Logger.Infof("Downloading found %d packages simultaneously in groups of: %d", len(foundPackagesArr), global_vars.PackagesMaxConcurrentDownloadCount)
	mylog.Logger.Infof(" to dir: %s", global_vars.DownloadPkgsDirPath)

	wg := sync.WaitGroup{}
	// Ensure all routines finish before returning
	defer wg.Wait()
	concurrentCountGuard := make(chan int, global_vars.PackagesMaxConcurrentDownloadCount) // Set an array of size: 'global_vars.PackagesMaxConcurrentDownloadCount'

	// Download concurrently with threads
	for _, pkgDetailsStruct := range foundPackagesArr {
		if len(pkgDetailsStruct.Name) == 0 || len(pkgDetailsStruct.Version) == 0 {
			mylog.Logger.Info("Skipping downloading of an unnamed/no-versioned pkg")
			continue
		}
		
		wg.Add(1)
		fileName := pkgDetailsStruct.Name + "." + pkgDetailsStruct.Version + ".nupkg"
		downloadFilePath := helper_funcs.Filepath_Join(global_vars.DownloadPkgsDirPath, fileName) // 'downloadPkgsDirPath' is a global var
		downloadPkgDetailsStruct := global_structs.DownloadPackageDetailsStruct{
			PkgDetailsStruct:         pkgDetailsStruct,
			DownloadFilePath:         downloadFilePath,
			DownloadFileChecksum:     helper_funcs.CalculateFileChecksum(downloadFilePath), // Can by empty if file doesn't exist yet
			DownloadFileChecksumType: "SHA512",                                        // Default checksum algorithm for Nuget pkgs
		}
		
		concurrentCountGuard <- 1; // Add 1 to concurrent threads count - Would block if array is filled. Can only be freed by thread executing: '<- concurrentCountGuard' below
		go func(downloadPkgDetailsStruct global_structs.DownloadPackageDetailsStruct) {
			defer wg.Done()
			DownloadNugetPackage(downloadPkgDetailsStruct)
			helper_funcs.Synched_AppendDownloadedPkgDetailsObj(&totalDownloadedPackagesDetailsArr, downloadPkgDetailsStruct)
			<- concurrentCountGuard  // Remove 1 from 'concurrentCountGuard'
		}(downloadPkgDetailsStruct)
	}
	wg.Wait()

	return totalDownloadedPackagesDetailsArr
}

func UploadDownloadedPackage(uploadPkgStruct global_structs.UploadPackageDetailsStruct) global_structs.UploadPackageDetailsStruct {
	pkgPrintStr := helper_funcs.Fmt_Sprintf("%s==%s", uploadPkgStruct.PkgDetailsStruct.Name, uploadPkgStruct.PkgDetailsStruct.Version)
	pkgName := uploadPkgStruct.PkgDetailsStruct.Name
	pkgVersion := uploadPkgStruct.PkgDetailsStruct.Version

	// Check if package already exists. If so, then compare it's checksum and skip on matching
	for _, destServerUrl := range global_vars.DestServersUrlsArr {
		mylog.Logger.Infof("Checking if pkg: '%s' already exists at dest server: %s", pkgPrintStr, destServerUrl)
		checkDestServerPkgExistUrl := destServerUrl + "/" + "Packages(Id='" + pkgName + "',Version='" + pkgVersion + "')"
		httpRequestArgs := global_structs.HttpRequestArgsStruct{
			UrlAddress: checkDestServerPkgExistUrl,
			HeadersMap: global_vars.HttpRequestHeadersMap,
			UserToUse:  global_vars.DestServersUserToUse,
			PassToUse:  global_vars.DestServersPassToUse,
			TimeoutSec: global_vars.HttpRequestTimeoutSecondsInt,
			Method:     "GET",
		}

		foundPackagesDetailsArr := SearchSpecificPackageVersionByURLRequest(httpRequestArgs)
		foundPackagesCount := len(foundPackagesDetailsArr)
		mylog.Logger.Debugf("Found: %s", foundPackagesDetailsArr)
		
		emptyNugetPackageDetailsStruct := global_structs.NugetPackageDetailsStruct{}
		shouldCompareChecksum := true
		if foundPackagesCount != 1 {
			mylog.Logger.Infof("Found multiple or no packages: \"%d\" - Should be only 1. Skipping checksum comparison. Continuing with the upload..", foundPackagesCount)
			shouldCompareChecksum = false
		} else if foundPackagesCount == 1 && foundPackagesDetailsArr[0] == emptyNugetPackageDetailsStruct {
			mylog.Logger.Info("No package found. Continuing with the upload..")
			shouldCompareChecksum = false
		}
		
		if shouldCompareChecksum {
			// Check the checksum:
			mylog.Logger.Infof("Comparing found package's checksum to know if should upload to: %s or not", destServerUrl)
			foundPackageChecksum := foundPackagesDetailsArr[0].Checksum
			fileToUploadChecksum := uploadPkgStruct.UploadFileChecksum
			if foundPackageChecksum == fileToUploadChecksum {
			fileName := helper_funcs.Filepath_GetFileNameFromPath(uploadPkgStruct.UploadFilePath)
			mylog.Logger.Warnf("Checksum match: upload target file already exists in dest server: '%s' \n"+
				"Skipping upload of pkg: \"%s\"", destServerUrl, fileName)
			return uploadPkgStruct
			}
		}
		
		if len(destServerUrl) > 1 {
			lastChar := destServerUrl[len(destServerUrl)-1:]
			if lastChar != "/" {
				mylog.Logger.Debugf("Adding '/' char to dest server repo url: \"%s\"", destServerUrl)
				destServerUrl += "/"
			}
		}
		httpRequestArgs.UrlAddress = destServerUrl
		// Upload the package file
		UploadPkg(uploadPkgStruct, httpRequestArgs)
	
	}

	return uploadPkgStruct
}

func UploadPkg(uploadPkgStruct global_structs.UploadPackageDetailsStruct, httpRequestArgsStruct global_structs.HttpRequestArgsStruct) {
    pkgPrintStr := helper_funcs.Fmt_Sprintf("%s==%s", uploadPkgStruct.PkgDetailsStruct.Name, uploadPkgStruct.PkgDetailsStruct.Version)
	mylog.Logger.Infof("Uploading package: \"%s\" from: %s", pkgPrintStr, uploadPkgStruct.UploadFilePath)
    httpRequestArgsStruct.Method = "PUT"
    httpRequestArgsStruct.UploadFilePath = uploadPkgStruct.UploadFilePath
    helper_funcs.MakeHttpRequest(httpRequestArgsStruct)

}

func UploadDownloadedPackages(downloadedPkgsArr []global_structs.DownloadPackageDetailsStruct) {
	if len(downloadedPkgsArr) == 0 {
		mylog.Logger.Warnf("No packages to upload given")
		return
	}
	mylog.Logger.Infof("Uploading %d downloaded packages to servers: %v  in groups of: %d", len(downloadedPkgsArr), global_vars.DestServersUrlsArr, global_vars.PackagesMaxConcurrentUploadCount)
	if len(global_vars.DestServersUrlsArr) == 0 {
		mylog.Logger.Warnf("No servers to upload to were given - skipping uploading of: %d packages", len(downloadedPkgsArr))
		return
	}

	wg := sync.WaitGroup{}
	concurrentCountGuard := make(chan int, global_vars.PackagesMaxConcurrentUploadCount) // Set an array of size: 'global_vars.PackagesMaxConcurrentUploadCount'

	// Upload concurrently with threads
	for _, downloadedPkgStruct := range downloadedPkgsArr {
		wg.Add(1)
		concurrentCountGuard <- 1; // Add 1 to concurrent threads count - Would block if array is filled. Can only be freed by thread executing: '<- concurrentCountGuard' below
		go func(downloadedPkgDetails global_structs.DownloadPackageDetailsStruct) {
			defer wg.Done()
			UploadDownloadedPackage(global_structs.UploadPackageDetailsStruct{
				PkgDetailsStruct       : downloadedPkgDetails.PkgDetailsStruct,
				UploadFilePath         : downloadedPkgDetails.DownloadFilePath,
				UploadFileChecksum     : downloadedPkgDetails.DownloadFileChecksum,
				UploadFileChecksumType : downloadedPkgDetails.DownloadFileChecksumType,
			})
			<- concurrentCountGuard  // Remove 1 from 'concurrentCountGuard'
		}(downloadedPkgStruct)
	}
	
	wg.Wait()
	
}

func DeleteRemoteUnuploadedPackages(uploadedPkgsArr []global_structs.DownloadPackageDetailsStruct, destServersFoundPackagesMap map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct) {
	if len(uploadedPkgsArr) == 0 {return}
	remoteServersStr := global_vars.DestServersUrlsStr
    mylog.Logger.Infof("Removing all unuploaded packages from remote servers: %s", remoteServersStr)
	mylog.Logger.Info("")
    
	var destServersPackagesToRemoveMap map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct
	destServersPackagesToRemoveMap = destServersFoundPackagesMap
	
	// Remove packages that shouldn't be deleted:
	for _, uploadedPkgStruct := range uploadedPkgsArr {
		pkgName := uploadedPkgStruct.PkgDetailsStruct.Name
		pkgHash := uploadedPkgStruct.PkgDetailsStruct.HashCode()
		for _, destServerPackagesMap := range destServersPackagesToRemoveMap {
			if packagesMap, isMapContainsKey := destServerPackagesMap[pkgName]; isMapContainsKey {
				mylog.Logger.Warnf("Skip delete remote pkg: %s", pkgHash)
				delete(packagesMap, pkgHash)
			}
		}
	}

	pkgsCount := 0
	for _, destServerPkgsMap := range destServersPackagesToRemoveMap {
		for _, foundPkgsMap := range destServerPkgsMap {
			pkgsCount += len(foundPkgsMap)
		}
	}
	mylog.Logger.Infof("Removing total of %d packages from dest servers: %s", pkgsCount, global_vars.DestServersUrlsArr)
	if pkgsCount == 0 {
		mylog.Logger.Info(" ** No packages to remove from dest servers")
		return
	}

	wg := sync.WaitGroup{}
	concurrentCountGuard := make(chan int, global_vars.PackagesMaxConcurrentDeleteCount) // Set an array of size: 'global_vars.PackagesMaxConcurrentDeleteCount'

	// Upload concurrently with threads
	for destServerUrl, destServerPkgsMap := range destServersPackagesToRemoveMap {
		for _, foundPkgsMap := range destServerPkgsMap {
			for packageHash, packageDetailsStruct := range foundPkgsMap {
				wg.Add(1)
				concurrentCountGuard <- 1; // Add 1 to concurrent threads count - Would block if array is filled. Can only be freed by thread executing: '<- concurrentCountGuard' below
				go func(pkgHash string, pkgDetailsStruct global_structs.NugetPackageDetailsStruct) {
					defer wg.Done()
					mylog.Logger.Debugf("Deleting pkg: %s from: %s", pkgHash, destServerUrl)
					DeleteRemotePackage(pkgDetailsStruct)
					<- concurrentCountGuard  // Remove 1 from 'concurrentCountGuard'
				}(packageHash, packageDetailsStruct)
			}
		}
	}

	wg.Wait()
}

func DeleteRemotePackage(pkgToDeleteStruct global_structs.NugetPackageDetailsStruct) {
	delRequestArgs := global_structs.HttpRequestArgsStruct{
		UrlAddress: pkgToDeleteStruct.PkgFileUrl,  // Only the url is different each time
		HeadersMap: global_vars.HttpRequestHeadersMap,
		UserToUse:  global_vars.DestServersUserToUse,
		PassToUse:  global_vars.DestServersPassToUse,
		TimeoutSec: global_vars.HttpRequestTimeoutSecondsInt,
		Method:     "DELETE",
	}
	httpResponsePtr := helper_funcs.MakeHttpRequest(delRequestArgs)
	if httpResponsePtr == nil {
		mylog.Logger.Errorf("'httpResponsePtr' is null. Failed to delete package: %s", delRequestArgs.UrlAddress)
		return
	}
	httpResponse := *httpResponsePtr
	responseBody := httpResponse.BodyStr
	httpCode_NotFound := 404
	if (httpResponse.StatusCode == httpCode_NotFound) {return} // <- OK

	if httpResponse.StatusCode >= 400 {
		mylog.Logger.Errorf("%s %s - Failed to delete package: %s", httpResponse.StatusStr, responseBody, delRequestArgs.UrlAddress)
	}
}
