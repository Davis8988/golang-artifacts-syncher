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

func DownloadNugetPackage(downloadPkgDetailsStruct global_structs.DownloadPackageDetailsStruct) global_structs.DownloadPackageDetailsStruct {
	mylog.Logger.Infof("Downloading package: %s==%s", downloadPkgDetailsStruct.PkgDetailsStruct.Name, downloadPkgDetailsStruct.PkgDetailsStruct.Version)
    fileUrl := downloadPkgDetailsStruct.PkgDetailsStruct.PkgFileUrl
    downloadFilePath := downloadPkgDetailsStruct.DownloadFilePath
    downloadFileChecksum := downloadPkgDetailsStruct.DownloadFileChecksum
    fileChecksum := downloadPkgDetailsStruct.PkgDetailsStruct.Checksum
    if fileChecksum == downloadFileChecksum {
        fileName := helper_funcs.GetFileName(downloadFilePath)
        mylog.Logger.Warnf("Checksum match: download target file already exists. Skipping download of: \"%s\"", fileName)
		downloadPkgDetailsStruct.IsSuccessful = true
        return downloadPkgDetailsStruct
    }
    httpResponsePtr := helper_funcs.MakeHttpRequest(
        global_structs.HttpRequestArgsStruct{
            UrlAddress: fileUrl,
            Method: "GET",
            DownloadFilePath: downloadFilePath,
            TimeoutSec: global_vars.AppConfig.HttpRequestDownloadTimeoutSecondsInt,
        },
    )

	if httpResponsePtr == nil {return downloadPkgDetailsStruct}
	httpResponse := *httpResponsePtr
	if httpResponse.StatusCode < 400 {downloadPkgDetailsStruct.IsSuccessful = true}
	return downloadPkgDetailsStruct
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
		helper_funcs.Synched_ErrorsDetected(true)
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
	skipGroupCount := global_vars.AppConfig.SearchPackagesUrlSkipGroupCount;
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
		helper_funcs.Synched_ErrorsDetected(true)
		return [] global_structs.NugetPackageDetailsStruct {}
    }

	if (len(responseBody) == 0) {return [] global_structs.NugetPackageDetailsStruct {}}
    
    parsedPackagesDetailsArr := ParseHttpRequestResponseForPackagesVersions(responseBody)

    return parsedPackagesDetailsArr
}

func SearchSrcServersForAvailableVersionsOfSpecifiedPackages() []global_structs.NugetPackageDetailsStruct {
	var totalFoundPackagesDetailsArr []global_structs.NugetPackageDetailsStruct
	
	wg := sync.WaitGroup{}
	for _, pkgName := range global_vars.AppConfig.PackagesNamesArr {
		wg.Add(1)
        go func(packageNameToSearch string) {
			defer wg.Done()
			searchUrlsArr := helper_funcs.PrepareSrcSearchUrlsForPackageArray(packageNameToSearch)
			threadFoundPackagesDetailsMap := make(map[string] global_structs.NugetPackageDetailsStruct)
			for _, urlToCheck := range searchUrlsArr {
				getRequestArgs := global_structs.HttpRequestArgsStruct{
					UrlAddress: urlToCheck,  // Only the url is different each time
					HeadersMap: global_vars.AppConfig.HttpRequestHeadersMap,
					UserToUse:  global_vars.AppConfig.SrcServersUserToUse,
					PassToUse:  global_vars.AppConfig.SrcServersPassToUse,
					TimeoutSec: global_vars.AppConfig.HttpRequestGlobalDefaultTimeoutSecondsInt,
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
			mylog.Logger.Debugf("Keeping last: %d of found '%s' packages", global_vars.AppConfig.PackagesDownloadLimitCount, packageNameToSearch)
			threadFoundPackagesDetailsArr = helper_funcs.FilterLastNPackages(threadFoundPackagesDetailsArr, global_vars.AppConfig.PackagesDownloadLimitCount)
			mylog.Logger.Infof("Targeted %d of '%s' packages", len(threadFoundPackagesDetailsArr), packageNameToSearch)
			helper_funcs.Synched_JoinTwoPkgDetailsSlices(&totalFoundPackagesDetailsArr, threadFoundPackagesDetailsArr)
		} (pkgName)
	}

	mylog.Logger.Debug("Waiting for searching threads to finish searching for pkgs")
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
	for _, destServerUrl := range global_vars.AppConfig.DestServersUrlsArr {
		wg.Add(1)
        go func(destServerUrlToSearch string) {
			defer wg.Done()
			threadFoundPackagesDetailsMap := make(map[string] map[string] global_structs.NugetPackageDetailsStruct)
			foundPackagesCount := 0
			for _, pkgName := range global_vars.AppConfig.PackagesNamesArr {
				urlToCheck := destServerUrlToSearch + "Packages()?$filter=tolower(Id)%20eq%20'"+pkgName+"'"
				getRequestArgs := global_structs.HttpRequestArgsStruct{
					UrlAddress: urlToCheck,  // Only the url is different each time
					HeadersMap: global_vars.AppConfig.HttpRequestHeadersMap,
					UserToUse:  global_vars.AppConfig.SrcServersUserToUse,
					PassToUse:  global_vars.AppConfig.SrcServersPassToUse,
					TimeoutSec: global_vars.AppConfig.HttpRequestGlobalDefaultTimeoutSecondsInt,
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
			
			mylog.Logger.Infof("Found total %d of '%s' packages in dest server: %s", foundPackagesCount, global_vars.AppConfig.PackagesNamesStr, destServerUrlToSearch)
			helper_funcs.Synched_AddPkgDetailsStructMapToMap(totalFoundDestServersPackagesDetailsMap, destServerUrlToSearch, threadFoundPackagesDetailsMap)
		} (destServerUrl)
	}

	mylog.Logger.Debug("Waiting for searching threads to finish searching for pkgs")
	wg.Wait()
	pkgsCount := 0
	for _, destServerFoundPkgsMap := range totalFoundDestServersPackagesDetailsMap {
		for _, foundPkgsMap := range destServerFoundPkgsMap {
			pkgsCount += len(foundPkgsMap)
		}
	}
	mylog.Logger.Infof("Total of %d packages found at dest servers: %s", pkgsCount, global_vars.AppConfig.DestServersUrlsArr)
	
	return totalFoundDestServersPackagesDetailsMap
}


func DownloadFoundPackages(foundPackagesArr []global_structs.NugetPackageDetailsStruct, destServersFoundPackagesMap map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct) []global_structs.DownloadPackageDetailsStruct {
	foundPackagesCount := len(foundPackagesArr)
	if foundPackagesCount == 0 {
		mylog.Logger.Warn("No packages to download found")
		return []global_structs.DownloadPackageDetailsStruct {}
	}
	totalDownloadedPackagesDetailsArr := make([]global_structs.DownloadPackageDetailsStruct, 0, foundPackagesCount)
	mylog.Logger.Infof("Downloading %d found packages simultaneously in groups of: %d", len(foundPackagesArr), global_vars.AppConfig.PackagesMaxConcurrentDownloadCount)
	mylog.Logger.Infof(" to dir: %s", global_vars.AppConfig.DownloadPkgsDirPath)

	wg := new(sync.WaitGroup)
	
	concurrentCountGuard := make(chan int, global_vars.AppConfig.PackagesMaxConcurrentDownloadCount) // Set an array of size: 'global_vars.PackagesMaxConcurrentDownloadCount'

	// Download concurrently with threads
	for _, pkgDetailsStruct := range foundPackagesArr {
		if len(pkgDetailsStruct.Name) == 0 || len(pkgDetailsStruct.Version) == 0 {
			mylog.Logger.Info("Skipping downloading of an unnamed/no-versioned pkg")
			continue
		}

		if doAllDestServersContainPackage(pkgDetailsStruct, destServersFoundPackagesMap) {
			mylog.Logger.Warnf("Checksum match: remote target package: %s already exists in all dest servers: '%s'. Skipping download of it", pkgDetailsStruct.HashCode(), global_vars.AppConfig.DestServersUrlsStr)
			continue
		}

		fileName := pkgDetailsStruct.Name + "." + pkgDetailsStruct.Version + ".nupkg"
		downloadFilePath := helper_funcs.Filepath_Join(global_vars.AppConfig.DownloadPkgsDirPath, fileName) // 'downloadPkgsDirPath' is a global var
		pkgFileChecksum := helper_funcs.CalculateFileChecksum(downloadFilePath);
		downloadPkgDetailsStruct := global_structs.DownloadPackageDetailsStruct{
			PkgDetailsStruct:         pkgDetailsStruct,
			DownloadFilePath:         downloadFilePath,
			DownloadFileChecksum:     pkgFileChecksum, // Can by empty if file doesn't exist yet
			IsSuccessful:             false,
			DownloadFileChecksumType: "SHA512",                                        // Default checksum algorithm for Nuget pkgs
		}
		
		// Record 1 thread
		wg.Add(1)
		concurrentCountGuard <- 1; // Add 1 to concurrent threads count - Would block if array is filled. Can only be freed by thread executing: '<- concurrentCountGuard' below
		go func(downloadPkgDetailsStruct global_structs.DownloadPackageDetailsStruct) {
			defer wg.Done()
			downloadPkgDetailsStruct = DownloadNugetPackage(downloadPkgDetailsStruct)
			helper_funcs.Synched_AppendDownloadedPkgDetailsObj(&totalDownloadedPackagesDetailsArr, downloadPkgDetailsStruct)
			<- concurrentCountGuard  // Remove 1 from 'concurrentCountGuard'
		}(downloadPkgDetailsStruct)
	}

	mylog.Logger.Debugf("Waiting for downloading threads")
	wg.Wait()
	mylog.Logger.Debugf("Finished waiting for downloading threads")

	return totalDownloadedPackagesDetailsArr
}

func doAllDestServersContainPackage(pkgDetailsStruct global_structs.NugetPackageDetailsStruct, destServersFoundPackagesMap map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct) bool {
	pkgName := pkgDetailsStruct.Name
	pkgHash := pkgDetailsStruct.HashCode()
	pkgChecksum := pkgDetailsStruct.Checksum
	
	// Skipping download of already dest-servers-existing pkgs
	for destServerUrl, destServerPackagesMap := range destServersFoundPackagesMap {
		mylog.Logger.Debugf("Checking if dest server: %s already contains pkg: %s", destServerUrl, pkgHash)
		packagesMap, isDestServerContainsPkgsMap := destServerPackagesMap[pkgName]
		if ! isDestServerContainsPkgsMap {return false}

		// Found map of versions of the package - checking if contains the specific pkg version
		packageDetailsStruct, isPkgsMapContainsPackage := packagesMap[pkgHash]
		if ! isPkgsMapContainsPackage {return false}
		
		// Found the specific pkg version - comparing checksum
		destServerPkgChecksum := packageDetailsStruct.Checksum
		if destServerPkgChecksum != pkgChecksum {return false}  // If found at least 1 dest server that doesn't contain the pkg - return false!
	}
	return true
}

func UploadDownloadedPackage(uploadPkgStruct global_structs.UploadPackageDetailsStruct) global_structs.UploadPackageDetailsStruct{
	pkgPrintStr := helper_funcs.Fmt_Sprintf("%s==%s", uploadPkgStruct.PkgDetailsStruct.Name, uploadPkgStruct.PkgDetailsStruct.Version)
	pkgName := uploadPkgStruct.PkgDetailsStruct.Name
	pkgVersion := uploadPkgStruct.PkgDetailsStruct.Version

	// Check if package already exists. If so, then compare it's checksum and skip on matching
	for _, destServerUrl := range global_vars.AppConfig.DestServersUrlsArr {
		mylog.Logger.Infof("Checking if pkg: '%s' already exists at dest server: %s", pkgPrintStr, destServerUrl)
		checkDestServerPkgExistUrl := destServerUrl + "/" + "Packages(Id='" + pkgName + "',Version='" + pkgVersion + "')"
		uploadHttpRequestArgs := global_structs.HttpRequestArgsStruct{
			UrlAddress: checkDestServerPkgExistUrl,
			HeadersMap: global_vars.AppConfig.HttpRequestHeadersMap,
			UserToUse:  global_vars.AppConfig.DestServersUserToUse,
			PassToUse:  global_vars.AppConfig.DestServersPassToUse,
			TimeoutSec: global_vars.AppConfig.HttpRequestUploadTimeoutSecondsInt,
			Method:     "GET",
		}

		foundPackagesDetailsArr := SearchSpecificPackageVersionByURLRequest(uploadHttpRequestArgs)
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
				uploadPkgStruct.IsSuccessful = true
				continue
			}
		}
		
		if len(destServerUrl) > 1 {
			lastChar := destServerUrl[len(destServerUrl)-1:]
			if lastChar != "/" {
				mylog.Logger.Debugf("Adding '/' char to dest server repo url: \"%s\"", destServerUrl)
				destServerUrl += "/"
			}
		}
		uploadHttpRequestArgs.UrlAddress = destServerUrl
		// Upload the package file
		uploadPkgStruct = UploadPkg(uploadPkgStruct, uploadHttpRequestArgs)
	}
	return uploadPkgStruct
}

func UploadPkg(uploadPkgStruct global_structs.UploadPackageDetailsStruct, uploadHttpRequestArgs global_structs.HttpRequestArgsStruct) global_structs.UploadPackageDetailsStruct{
    pkgPrintStr := helper_funcs.Fmt_Sprintf("%s==%s", uploadPkgStruct.PkgDetailsStruct.Name, uploadPkgStruct.PkgDetailsStruct.Version)
	mylog.Logger.Infof("Uploading package: \"%s\" from: %s", pkgPrintStr, uploadPkgStruct.UploadFilePath)
    uploadHttpRequestArgs.Method = "PUT"
    uploadHttpRequestArgs.UploadFilePath = uploadPkgStruct.UploadFilePath
    httpResponsePtr := helper_funcs.MakeHttpRequest(uploadHttpRequestArgs)

	if httpResponsePtr == nil {return uploadPkgStruct}
	httpResponse := *httpResponsePtr
	if httpResponse.StatusCode <= 400 {uploadPkgStruct.IsSuccessful = true}
	return uploadPkgStruct
}

func UploadDownloadedPackages(downloadedPkgsArr []global_structs.DownloadPackageDetailsStruct) []global_structs.UploadPackageDetailsStruct {
	downloadedPkgsCount := len(downloadedPkgsArr)
	totalUploadedPackagesDetailsArr := make([] global_structs.UploadPackageDetailsStruct, 0 , downloadedPkgsCount)
	if downloadedPkgsCount == 0 {
		mylog.Logger.Warnf("No packages to upload given")
		return totalUploadedPackagesDetailsArr
	}
	mylog.Logger.Infof("Uploading %d downloaded packages to servers: %v  in groups of: %d", downloadedPkgsCount, global_vars.AppConfig.DestServersUrlsArr, global_vars.AppConfig.PackagesMaxConcurrentUploadCount)
	if len(global_vars.AppConfig.DestServersUrlsArr) == 0 {
		mylog.Logger.Warnf("No servers to upload to were given - skipping uploading of: %d packages", downloadedPkgsCount)
		return totalUploadedPackagesDetailsArr
	}

	wg := sync.WaitGroup{}
	concurrentCountGuard := make(chan int, global_vars.AppConfig.PackagesMaxConcurrentUploadCount) // Set an array of size: 'global_vars.PackagesMaxConcurrentUploadCount'

	// Upload concurrently with threads
	for _, downloadedPkgStruct := range downloadedPkgsArr {
		wg.Add(1)
		concurrentCountGuard <- 1; // Add 1 to concurrent threads count - Would block if array is filled. Can only be freed by thread executing: '<- concurrentCountGuard' below
		go func(downloadedPkgDetails global_structs.DownloadPackageDetailsStruct) {
			defer wg.Done()
			uploadPkgStruct := global_structs.UploadPackageDetailsStruct{
				PkgDetailsStruct       : downloadedPkgDetails.PkgDetailsStruct,
				UploadFilePath         : downloadedPkgDetails.DownloadFilePath,
				UploadFileChecksum     : downloadedPkgDetails.DownloadFileChecksum,
				UploadFileChecksumType : downloadedPkgDetails.DownloadFileChecksumType,
				IsSuccessful           : false,
			}
			uploadPkgStruct = UploadDownloadedPackage(uploadPkgStruct)
			helper_funcs.Synched_AppendUploadedPkgDetailsObj(&totalUploadedPackagesDetailsArr, uploadPkgStruct)
			<- concurrentCountGuard  // Remove 1 from 'concurrentCountGuard'
		}(downloadedPkgStruct)
	}
	
	mylog.Logger.Debugf("Waiting for uploading threads")
	wg.Wait()
	mylog.Logger.Debugf("Finished waiting for uploading threads")

	return totalUploadedPackagesDetailsArr
	
}

func DeleteRemoteUnuploadedPackages(uploadedPkgsArr []global_structs.NugetPackageDetailsStruct, destServersFoundPackagesMap map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct) {
	if len(uploadedPkgsArr) == 0 {return}
	remoteServersStr := global_vars.AppConfig.DestServersUrlsStr
    mylog.Logger.Infof("Removing all unuploaded packages from remote servers: %s", remoteServersStr)
	mylog.Logger.Info("")
    
	var destServersPackagesToRemoveMap map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct
	destServersPackagesToRemoveMap = destServersFoundPackagesMap
	
	// Remove packages that shouldn't be deleted:
	for _, pkgDetailsStruct := range uploadedPkgsArr {
		pkgName := pkgDetailsStruct.Name
		pkgHash := pkgDetailsStruct.HashCode()
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
	mylog.Logger.Infof("Removing total of %d packages from dest servers: %s", pkgsCount, global_vars.AppConfig.DestServersUrlsArr)
	if pkgsCount == 0 {
		mylog.Logger.Info(" ** No packages to remove from dest servers")
		return
	}

	wg := sync.WaitGroup{}
	concurrentCountGuard := make(chan int, global_vars.AppConfig.PackagesMaxConcurrentDeleteCount) // Set an array of size: 'global_vars.PackagesMaxConcurrentDeleteCount'

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
		HeadersMap: global_vars.AppConfig.HttpRequestHeadersMap,
		UserToUse:  global_vars.AppConfig.DestServersUserToUse,
		PassToUse:  global_vars.AppConfig.DestServersPassToUse,
		TimeoutSec: global_vars.AppConfig.HttpRequestGlobalDefaultTimeoutSecondsInt,
		Method:     "DELETE",
	}
	httpResponsePtr := helper_funcs.MakeHttpRequest(delRequestArgs)
	if httpResponsePtr == nil {
		mylog.Logger.Errorf("'httpResponsePtr' is null. Failed to delete package: %s", delRequestArgs.UrlAddress)
		helper_funcs.Synched_ErrorsDetected(true)
		return
	}
	httpResponse := *httpResponsePtr
	responseBody := httpResponse.BodyStr
	httpCode_NotFound := 404
	if (httpResponse.StatusCode == httpCode_NotFound) {return} // <- OK

	if httpResponse.StatusCode >= 400 {
		mylog.Logger.Errorf("%s %s - Failed to delete package: %s", httpResponse.StatusStr, responseBody, delRequestArgs.UrlAddress)
		helper_funcs.Synched_ErrorsDetected(true)
	}
}
