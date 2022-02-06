package main

import (
	"flag"
	"fmt"
	"golang-artifacts-syncher/src/helpers"
	"golang-artifacts-syncher/src/nexus3_adapter"
	"path/filepath"
	"strings"
	"sync"
)

var (
	srcServersUserToUse          string
	srcServersPassToUse          string
	srcServersUrlsStr            string
	srcReposNamesStr             string
	destServersUrlsStr           string
	destReposNamesStr            string
	destServersUserToUse         string
	destServersPassToUse         string
	packagesNamesStr             string
	packagesVersionsStr          string
	httpRequestHeadersStr        string
	downloadPkgsDirPath          string
	httpRequestTimeoutSecondsInt int

	srcServersUrlsArr     []string
	srcReposNamesArr      []string
	destServersUrlsArr    []string
	destReposNamesArr     []string
	packagesNamesArr      []string
	packagesVersionsArr   []string
	httpRequestHeadersMap map[string]string
	packagesToDownloadMap sync.Map
)

func initVars() {
	helpers.Init()
	helpers.LogInfo.Print("Initializing from envs vars")
	srcServersUserToUse = helpers.Getenv("SRC_SERVERS_USER_TO_USE", "")
	srcServersPassToUse = helpers.Getenv("SRC_SERVERS_PASS_TO_USE", "")
	srcServersUrlsStr = helpers.Getenv("SRC_SERVERS_URLS_STR", "")
	srcReposNamesStr = helpers.Getenv("SRC_REPOS_NAMES_STR", "")
	destServersUrlsStr = helpers.Getenv("DEST_SERVERS_URLS_STR", "")
	destReposNamesStr = helpers.Getenv("DEST_REPOS_NAMES_STR", "")
	destServersUserToUse = helpers.Getenv("DEST_SERVERS_USER_TO_USE", "")
	destServersPassToUse = helpers.Getenv("DEST_SERVERS_PASS_TO_USE", "")
	packagesNamesStr = helpers.Getenv("PACKAGES_NAMES_STR", "")
	packagesVersionsStr = helpers.Getenv("PACKAGES_VERSIONS_STR", "")
	httpRequestHeadersStr = helpers.Getenv("HTTP_REQUEST_HEADERS_STR", "") // Example: "key=value;key1=value1;key2=value2"
	downloadPkgsDirPath = helpers.Getenv("DOWNLOAD_PKGS_DIR_PATH", helpers.GetCurrentProgramDir())
	httpRequestTimeoutSecondsInt = helpers.StrToInt(helpers.Getenv("HTTP_REQUEST_TIMEOUT_SECONDS_INT", "45"))
}

func printVars() {
	helpers.LogInfo.Printf("SRC_SERVERS_URLS_STR: '%s'", srcServersUrlsStr)
	helpers.LogInfo.Printf("SRC_REPOS_NAMES_STR: '%s'", srcReposNamesStr)
	helpers.LogInfo.Printf("SRC_SERVERS_USER_TO_USE: '%s'", srcServersUserToUse)
	helpers.LogInfo.Printf("SRC_SERVERS_PASS_TO_USE: '%s'", strings.Repeat("*", len(srcServersPassToUse)))
	helpers.LogInfo.Printf("DEST_SERVERS_URLS_STR: '%s'", destServersUrlsStr)
	helpers.LogInfo.Printf("DEST_REPOS_NAMES_STR: '%s'", destReposNamesStr)
	helpers.LogInfo.Printf("DEST_SERVERS_USER_TO_USE: '%s'", destServersUserToUse)
	helpers.LogInfo.Printf("DEST_SERVERS_PASS_TO_USE: '%s'", strings.Repeat("*", len(destServersPassToUse)))
	helpers.LogInfo.Printf("PACKAGES_NAMES_STR: '%s'", packagesNamesStr)
	helpers.LogInfo.Printf("PACKAGES_VERSIONS_STR: '%s'", packagesVersionsStr)
	helpers.LogInfo.Printf("HTTP_REQUEST_HEADERS_STR: '%s'", httpRequestHeadersStr)
	helpers.LogInfo.Printf("DOWNLOAD_PKGS_DIR_PATH: '%s'", downloadPkgsDirPath)
	helpers.LogInfo.Printf("HTTP_REQUEST_TIMEOUT_SECONDS_INT: '%d'", httpRequestTimeoutSecondsInt)

	helpers.LogInfo.Printf("srcServersUrlsArr: %v", srcServersUrlsArr)
	helpers.LogInfo.Printf("destServersUrlsArr: %v", destServersUrlsArr)
	helpers.LogInfo.Printf("srcReposNamesArr: %v", srcReposNamesArr)
	helpers.LogInfo.Printf("packagesNamesArr: %v", packagesNamesArr)
	helpers.LogInfo.Printf("packagesVersionsArr: %v", packagesVersionsArr)
	packagesToDownloadMapStr := helpers.Synched_ConvertSyncedMapToString(packagesToDownloadMap)
	helpers.LogInfo.Printf("packagesToDownloadMap: \n%v", packagesToDownloadMapStr)
}

func validateEnv() {
	helpers.LogInfo.Print("Validating envs")

	// Validate len(packagesVersionsArr) == len(packagesNamesArr)  (Only when packagesVersionsArr is defined)
	if !nexus3_adapter.IsStrArrayEmpty(packagesVersionsArr) {
		helpers.LogInfo.Print("Comparing packages names & versions arrays lengths")
		if len(packagesVersionsArr) != len(packagesNamesArr) {
			errMsg := "Packages Versions to search count is different from Packages Names to search count\n"
			errMsg += "Can't search for packages versions & names which are not of the same count.\n"
			errMsg += "When passing packages versions to search - the versions count must be of the same count of packages names to search.\n"
			errMsg += "A version for each package name to search"
			helpers.LogError.Fatal(errMsg)
		}
	}

	helpers.LogInfo.Print("All Good")
}

func parseArgs() {
	helpers.LogInfo.Print("Parsing args")
	flag.Parse()
}

func updateVars() {
	helpers.LogInfo.Print("Updating vars")
	srcServersUrlsArr = make([]string, 0, 4)
	destServersUrlsArr = make([]string, 0, 4)
	srcReposNamesArr = make([]string, 0, 4)
	packagesNamesArr = make([]string, 0, 10)
	packagesVersionsArr = make([]string, 0, 10)
	if len(srcServersUrlsStr) > 1 {srcServersUrlsArr = strings.Split(srcServersUrlsStr, ";")}
	if len(srcReposNamesStr) > 1 {srcReposNamesArr = strings.Split(srcReposNamesStr, ";")}
	if len(destServersUrlsStr) > 1 {destServersUrlsArr = strings.Split(destServersUrlsStr, ";")}
	if len(destReposNamesStr) > 1 {destReposNamesArr = strings.Split(destReposNamesStr, ";")}
	if len(packagesNamesStr) > 1 {packagesNamesArr = strings.Split(packagesNamesStr, ";")}
	if len(packagesVersionsStr) > 1 {packagesVersionsArr = strings.Split(packagesVersionsStr, ";")}
	httpRequestHeadersMap = helpers.ParseHttpHeadersStrToMap(httpRequestHeadersStr)

	for i, pkgName := range packagesNamesArr {
		// If map doesn't contain value at: 'pkgName' - add one to point to empty string array: []
		packagesToDownloadMap.LoadOrStore(pkgName, make([]string, 0, 10))
		// If received a version array for it - add it to the list
		if len(packagesVersionsArr) > i {
			pkgVersion := packagesVersionsArr[i]
			currentVersionsArr := helpers.LoadStringArrValueFromSynchedMap(packagesToDownloadMap, pkgName)
			packagesToDownloadMap.Store(pkgName, append(currentVersionsArr, pkgVersion))
		}
	}
}

func prepareSrcSearchAllPkgsVersionsUrlsArray() []string {
	var searchUrlsArr = make([]string, 0, 10) // Create a slice with length=0 and capacity=10

	helpers.LogInfo.Print("Preparing src search packages urls array")
	for _, srcServerUrl := range srcServersUrlsArr {
		for _, repoName := range srcReposNamesArr {
			for _, pkgName := range packagesNamesArr {
				versionsToSearchArr := helpers.LoadStringArrValueFromSynchedMap(packagesToDownloadMap, pkgName)
				if len(versionsToSearchArr) == 0 { // Either use search
					searchUrlsArr = append(searchUrlsArr, srcServerUrl+"/"+repoName+"/"+"Packages()?$filter=tolower(Id)%20eq%20'"+pkgName+"'")
					continue
				} // Or specific package details request for each specified requested version
				for _, pkgVersion := range versionsToSearchArr {
					searchUrlsArr = append(searchUrlsArr, srcServerUrl+"/"+repoName+"/"+"Packages(Id='"+pkgName+"',Version='"+pkgVersion+"')")
				}

			}
		}
	}
	return searchUrlsArr
}

func filterFoundPackagesByRequestedVersion(foundPackagesDetailsArr []helpers.NugetPackageDetailsStruct) []helpers.NugetPackageDetailsStruct {
	helpers.LogInfo.Printf("Filtering found pkgs by requested versions")
	var filteredPackagesDetailsArr []helpers.NugetPackageDetailsStruct
	for _, pkgDetailStruct := range foundPackagesDetailsArr {
		pkgVersion := pkgDetailStruct.Version
		pkgName := pkgDetailStruct.Name
		versionsToSearchArr := helpers.LoadStringArrValueFromSynchedMap(packagesToDownloadMap, pkgName) // Use global var: packagesToDownloadMap
		if len(versionsToSearchArr) == 0 {
			filteredPackagesDetailsArr = append(filteredPackagesDetailsArr, pkgDetailStruct)
			continue
		}
		for _, requestedVersion := range versionsToSearchArr {
			if pkgVersion == requestedVersion {filteredPackagesDetailsArr = append(filteredPackagesDetailsArr, pkgDetailStruct)} // This version is requested - Add pkg details obj to the result filtered array
		}
	}
	return filteredPackagesDetailsArr
}

func searchAvailableVersionsOfSpecifiedPackages() []helpers.NugetPackageDetailsStruct {
	var totalFoundPackagesDetailsArr []helpers.NugetPackageDetailsStruct
	searchUrlsArr := prepareSrcSearchAllPkgsVersionsUrlsArray()

	wg := sync.WaitGroup{}

	// Ensure all routines finish before returning
	defer wg.Wait()

	if len(searchUrlsArr) > 0 {
		helpers.LogInfo.Printf("Checking %d src URL addresses for pkgs versions", len(searchUrlsArr))
		for _, urlToCheck := range searchUrlsArr {
			wg.Add(1)
			go func(urlToCheck string) {
				defer wg.Done()
				httpRequestArgs := helpers.HttpRequestArgsStruct{
					UrlAddress: urlToCheck,
					HeadersMap: httpRequestHeadersMap,
					UserToUse:  srcServersUserToUse,
					PassToUse:  srcServersPassToUse,
					TimeoutSec: httpRequestTimeoutSecondsInt,
					Method:     "GET",
				}
				foundPackagesDetailsArr := helpers.SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs)
				foundPackagesDetailsArr = filterFoundPackagesByRequestedVersion(foundPackagesDetailsArr) // Filter by requested version - if any version is specified..
				helpers.Synched_AppendPkgDetailsObj(&totalFoundPackagesDetailsArr, foundPackagesDetailsArr)
			}(urlToCheck)
		}
	}
	wg.Wait()

	return totalFoundPackagesDetailsArr
}

func downloadSpecifiedPackages(foundPackagesArr []helpers.NugetPackageDetailsStruct) []helpers.DownloadPackageDetailsStruct {
	helpers.LogInfo.Printf("Downloading found %d packages", len(foundPackagesArr))
	var totalDownloadedPackagesDetailsArr []helpers.DownloadPackageDetailsStruct

	wg := sync.WaitGroup{}
	// Ensure all routines finish before returning
	defer wg.Wait()

	for _, pkgDetailsStruct := range foundPackagesArr {
		if len(pkgDetailsStruct.Name) == 0 || len(pkgDetailsStruct.Version) == 0 {
			helpers.LogInfo.Print("Skipping downloading of an unnamed/unversioned pkg")
			continue
		}

		wg.Add(1)
		fileName := pkgDetailsStruct.Name + "." + pkgDetailsStruct.Version + ".nupkg"
		downloadFilePath := filepath.Join(downloadPkgsDirPath, fileName) // downloadPkgsDirPath == global var
		downloadPkgDetailsStruct := helpers.DownloadPackageDetailsStruct{
			PkgDetailsStruct:         pkgDetailsStruct,
			DownloadFilePath:         downloadFilePath,
			DownloadFileChecksum:     helpers.CalculateFileChecksum(downloadFilePath), // Can by empty if file doesn't exist yet
			DownloadFileChecksumType: "SHA512",                                        // Default checksum algorithm for Nuget pkgs
		}

		go func(downloadPkgDetailsStruct helpers.DownloadPackageDetailsStruct) {
			defer wg.Done()
			helpers.DownloadPkg(downloadPkgDetailsStruct)
			helpers.Synched_AppendDownloadedPkgDetailsObj(&totalDownloadedPackagesDetailsArr, downloadPkgDetailsStruct)
		}(downloadPkgDetailsStruct)
	}
	wg.Wait()

	return totalDownloadedPackagesDetailsArr
}

func uploadDownloadedPackage(uploadPkgStruct helpers.UploadPackageDetailsStruct) helpers.UploadPackageDetailsStruct {
	pkgPrintStr := fmt.Sprintf("%s==%s", uploadPkgStruct.PkgDetailsStruct.Name, uploadPkgStruct.PkgDetailsStruct.Version)
	pkgName := uploadPkgStruct.PkgDetailsStruct.Name
	pkgVersion := uploadPkgStruct.PkgDetailsStruct.Version

	// Check if package already exists. If so, then compare it's checksum and skip on matching
	for _, destServerUrl := range destServersUrlsArr {
		for _, repoName := range destReposNamesArr {
			destServerRepo := destServerUrl + "/" + repoName
			helpers.LogInfo.Printf("Checking if pkg: '%s' already exists at dest server: %s", pkgPrintStr, destServerRepo)
			checkDestServerPkgExistUrl := destServerRepo + "/" + "Packages(Id='" + pkgName + "',Version='" + pkgVersion + "')"
			httpRequestArgs := helpers.HttpRequestArgsStruct{
				UrlAddress: checkDestServerPkgExistUrl,
				HeadersMap: httpRequestHeadersMap,
				UserToUse:  destServersUserToUse,
				PassToUse:  destServersPassToUse,
				TimeoutSec: httpRequestTimeoutSecondsInt,
				Method:     "GET",
			}

			foundPackagesDetailsArr := helpers.SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs)
			helpers.LogInfo.Printf("Found: %s", foundPackagesDetailsArr)

			emptyNugetPackageDetailsStruct := helpers.NugetPackageDetailsStruct{}
			shouldCompareChecksum := true
			if len(foundPackagesDetailsArr) != 1 {
				helpers.LogInfo.Printf("Found multiple or no packages: \"%d\" - Should be only 1. Skipping checksum comparison. Continuing with the upload..", len(foundPackagesDetailsArr))
				shouldCompareChecksum = false
			} else if len(foundPackagesDetailsArr) == 1 && foundPackagesDetailsArr[0] == emptyNugetPackageDetailsStruct {
				helpers.LogInfo.Print("No package found. Continuing with the upload..")
				shouldCompareChecksum = false
			}
			
			if shouldCompareChecksum {
				// Check the checksum:
				helpers.LogInfo.Printf("Comparing found package's checksum to know if should upload to: %s or not", destServerRepo)
				foundPackageChecksum := foundPackagesDetailsArr[0].Checksum
				fileToUploadChecksum := uploadPkgStruct.UploadFileChecksum
				if foundPackageChecksum == fileToUploadChecksum {
				fileName := filepath.Base(uploadPkgStruct.UploadFilePath)
				helpers.LogWarning.Printf("Checksum match: upload target file already exists in dest server: '%s' \n"+
					"Skipping upload of pkg: \"%s\"", destServerRepo, fileName)
				return uploadPkgStruct
				}
			}
			
			httpRequestArgs.UrlAddress = destServerRepo
			// Upload the package file
			helpers.UploadPkg(uploadPkgStruct, httpRequestArgs)
		}
	}

	return uploadPkgStruct
}

func uploadDownloadedPackages(downloadedPkgsArr []helpers.DownloadPackageDetailsStruct) {
	helpers.LogInfo.Printf("Uploading %d downloaded packages to servers: %v", len(downloadedPkgsArr), destServersUrlsArr)
	if len(destServersUrlsArr) == 0 {
		helpers.LogWarning.Printf("No servers to upload to were given - skipping uploading of: %d packages", len(downloadedPkgsArr))
		return
	}
	for _, downloadedPkgStruct := range downloadedPkgsArr {
		uploadDownloadedPackage(helpers.UploadPackageDetailsStruct{
			PkgDetailsStruct:       downloadedPkgStruct.PkgDetailsStruct,
			UploadFilePath:         downloadedPkgStruct.DownloadFilePath,
			UploadFileChecksum:     downloadedPkgStruct.DownloadFileChecksum,
			UploadFileChecksumType: downloadedPkgStruct.DownloadFileChecksumType,
		})
	}
}

func main() {
	helpers.LogInfo.Print("Started")
	initVars()
	parseArgs()
	updateVars()
	printVars()
	validateEnv()
	foundPackagesArr := searchAvailableVersionsOfSpecifiedPackages()
	downloadedPkgsArr := downloadSpecifiedPackages(foundPackagesArr)
	uploadDownloadedPackages(downloadedPkgsArr)
	helpers.LogInfo.Print("Finished")
}
