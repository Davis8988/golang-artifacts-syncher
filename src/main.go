package main

import (
	"golang-artifacts-syncher/src/helpers"
	"golang-artifacts-syncher/src/nexus3_adapter"
	"flag"
	"sync"
	"strings"
)

var (
	userToUse string
	passToUse string
	srcServersUrlsStr string
	destServersUrlsStr string
	reposNamesStr string
	packagesNamesStr string
	packagesVersionsStr string
	httpRequestHeadersStr string
	downloadPkgsDirPath string
	httpRequestTimeoutSecondsInt int

	srcServersUrlsArr [] string
	reposNamesArr [] string
	packagesNamesArr [] string
	packagesVersionsArr [] string
	httpRequestHeadersMap map[string] string
	packagesToDownloadMap sync.Map

)

func initVars() {
	helpers.Init()
	helpers.LogInfo.Print("Initializing from envs vars")
	userToUse = helpers.Getenv("USER_TO_USE", "")
	passToUse = helpers.Getenv("PASS_TO_USE", "")
	srcServersUrlsStr = helpers.Getenv("SRC_SERVERS_URLS_STR", "")
	destServersUrlsStr = helpers.Getenv("DEST_SERVERS_URLS_STR", "")
	reposNamesStr = helpers.Getenv("REPOS_NAMES_STR", "")
	packagesNamesStr = helpers.Getenv("PACKAGES_NAMES_STR", "")
	packagesVersionsStr = helpers.Getenv("PACKAGES_VERSIONS_STR", "")
	httpRequestHeadersStr = helpers.Getenv("HTTP_REQUEST_HEADERS_STR", "")  // Example: "key=value;key1=value1;key2=value2"
	downloadPkgsDirPath = helpers.Getenv("DOWNLOAD_PKGS_DIR_PATH", helpers.GetCurrentProgramDir())
	httpRequestTimeoutSecondsInt = helpers.StrToInt(helpers.Getenv("HTTP_REQUEST_TIMEOUT_SECONDS_INT", "45"))
}

func printVars() {
	helpers.LogInfo.Printf("SRC_SERVERS_URLS_STR: '%s'", srcServersUrlsStr)
	helpers.LogInfo.Printf("DEST_SERVERS_URLS_STR: '%s'", destServersUrlsStr)
	helpers.LogInfo.Printf("REPOS_NAMES_STR: '%s'", reposNamesStr)
	helpers.LogInfo.Printf("PACKAGES_NAMES_STR: '%s'", packagesNamesStr)
	helpers.LogInfo.Printf("PACKAGES_VERSIONS_STR: '%s'", packagesVersionsStr)
	helpers.LogInfo.Printf("HTTP_REQUEST_HEADERS_STR: '%s'", httpRequestHeadersStr)
	helpers.LogInfo.Printf("DOWNLOAD_PKGS_DIR_PATH: '%s'", downloadPkgsDirPath)
	helpers.LogInfo.Printf("HTTP_REQUEST_TIMEOUT_SECONDS_INT: '%d'", httpRequestTimeoutSecondsInt)
	
	helpers.LogInfo.Printf("srcServersUrlsArr: %v", srcServersUrlsArr)
	helpers.LogInfo.Printf("reposNamesArr: %v", reposNamesArr)
	helpers.LogInfo.Printf("packagesNamesArr: %v", packagesNamesArr)
	helpers.LogInfo.Printf("packagesVersionsArr: %v", packagesVersionsArr)
	packagesToDownloadMapStr := helpers.Synched_ConvertSyncedMapToString(packagesToDownloadMap)
	helpers.LogInfo.Printf("packagesToDownloadMap: \n%v", packagesToDownloadMapStr)
}

func validateEnv() {
	helpers.LogInfo.Print("Validating envs")

	// Validate len(packagesVersionsArr) == len(packagesNamesArr)  (Only when packagesVersionsArr is defined)
	if ! nexus3_adapter.IsStrArrayEmpty(packagesVersionsArr) {
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
	reposNamesArr = make([]string, 0, 4)
	packagesNamesArr = make([]string, 0, 10)
	packagesVersionsArr = make([]string, 0, 10)
	if len(srcServersUrlsStr) > 1 {srcServersUrlsArr = strings.Split(srcServersUrlsStr, ";")}
	if len(reposNamesStr) > 1 {reposNamesArr = strings.Split(reposNamesStr, ";")}
	if len(packagesNamesStr) > 1 {packagesNamesArr = strings.Split(packagesNamesStr, ";")}
	if len(packagesVersionsStr) > 1 {packagesVersionsArr = strings.Split(packagesVersionsStr, ";")}
	httpRequestHeadersMap = helpers.ParseHttpHeadersStrToMap(httpRequestHeadersStr)

	for i, pkgName := range packagesNamesArr {
		// If map doesn't contain value at: 'pkgName' - add one to point to empty string array: []
		packagesToDownloadMap.LoadOrStore(pkgName, make([] string, 0, 10))
		// If received a version array for it - add it to the list
		if len(packagesVersionsArr) > i {
			pkgVersion := packagesVersionsArr[i]
			currentVersionsArr := helpers.LoadStringArrValueFromSynchedMap(packagesToDownloadMap, pkgName)
			packagesToDownloadMap.Store(pkgName, append(currentVersionsArr, pkgVersion))
		}
	}
}

func prepareSearchAllPkgsVersionsUrlsArray() []string {
	var searchUrlsArr = make([]string, 0, 10)  // Create a slice with length=0 and capacity=10
	
	helpers.LogInfo.Print("Preparing src search packages urls array")
	for _, srcServerUrl := range srcServersUrlsArr {
		for _, repoName := range reposNamesArr {
			for _, pkgName := range packagesNamesArr {
				versionsToSearchArr := helpers.LoadStringArrValueFromSynchedMap(packagesToDownloadMap, pkgName)
				if len(versionsToSearchArr) == 0 {  // Either use search
					searchUrlsArr = append(searchUrlsArr, srcServerUrl + "/" + repoName + "/" + "Search()?id='" + pkgName + "'")
					continue
				}                   // Or specific package details request for each specified requested version
				for _, pkgVersion := range versionsToSearchArr {
					searchUrlsArr = append(searchUrlsArr, srcServerUrl + "/" + repoName + "/" + "Packages(Id='" + pkgName + "',Version='" + pkgVersion + "')")
				}
				
			}
		}
	}
	return searchUrlsArr
} 

func filterFoundPackagesByRequestedVersion(foundPackagesDetailsArr [] helpers.NugetPackageDetailsStruct) [] helpers.NugetPackageDetailsStruct {
	helpers.LogInfo.Printf("Filtering found pkgs by requested versions")
	var filteredPackagesDetailsArr [] helpers.NugetPackageDetailsStruct
	for _, pkgDetailStruct := range foundPackagesDetailsArr {
		pkgVersion := pkgDetailStruct.Version
		pkgName := pkgDetailStruct.Name
		versionsToSearchArr := helpers.LoadStringArrValueFromSynchedMap(packagesToDownloadMap, pkgName)  // Use global var: packagesToDownloadMap
		if len(versionsToSearchArr) == 0 {
			filteredPackagesDetailsArr = append(filteredPackagesDetailsArr, pkgDetailStruct)
			continue
		}
		for _, requestedVersion := range versionsToSearchArr {
			if pkgVersion == requestedVersion {filteredPackagesDetailsArr = append(filteredPackagesDetailsArr, pkgDetailStruct)}  // This version is requested - Add pkg details obj to the result filtered array
		}
	}
	return filteredPackagesDetailsArr
}

func searchAvailableVersionsOfSpecifiedPackages() [] helpers.NugetPackageDetailsStruct {
	var totalFoundPackagesDetailsArr [] helpers.NugetPackageDetailsStruct
	searchUrlsArr := prepareSearchAllPkgsVersionsUrlsArray()
	
	wg := sync.WaitGroup{}
	
	// Ensure all routines finish before returning
	defer wg.Wait()

	if len(searchUrlsArr) > 0 {
		helpers.LogInfo.Printf("Checking %d URL addresses for pkgs versions", len(searchUrlsArr))
		for _, urlToCheck := range searchUrlsArr {
			wg.Add(1)
			go func(urlToCheck string) {
				defer wg.Done()
				httpRequestArgs := helpers.HttpRequestArgsStruct {
					UrlAddress: urlToCheck,
					HeadersMap: httpRequestHeadersMap,
					UserToUse: userToUse,
					PassToUse: passToUse,
					TimeoutSec: httpRequestTimeoutSecondsInt,
					Method: "GET",
				}
				foundPackagesDetailsArr := helpers.SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs)
				foundPackagesDetailsArr = filterFoundPackagesByRequestedVersion(foundPackagesDetailsArr)  // Filter by requested version - if any version is specified..
				helpers.Synched_AppendPkgDetailsObj(&totalFoundPackagesDetailsArr, foundPackagesDetailsArr)
			}(urlToCheck)
		}
	}
	wg.Wait()

	return totalFoundPackagesDetailsArr 
}

func downloadSpecifiedPackages(foundPackagesArr [] helpers.NugetPackageDetailsStruct) [] helpers.DownloadPackageDetailsStruct {
	helpers.LogInfo.Printf("Downloading found %d packages", len(foundPackagesArr))
	var totalDownloadedPackagesDetailsArr [] helpers.DownloadPackageDetailsStruct
	
	wg := sync.WaitGroup{}
	// Ensure all routines finish before returning
	defer wg.Wait()

	for _, pkgDetailsStruct := range foundPackagesArr {
		if len(pkgDetailsStruct.Name) == 0 || len(pkgDetailsStruct.Version) == 0 {
			helpers.LogInfo.Print("Skipping downloading of an unnamed/unversioned pkg")
			continue
		}

		wg.Add(1)
		downloadPkgDetailsStruct := helpers.DownloadPackageDetailsStruct {
			PkgDetailsStruct: pkgDetailsStruct,
			DownloadPath: downloadPkgsDirPath,
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

func uploadDownloadedPackage(downloadedPkgStruct helpers.DownloadPackageDetailsStruct) helpers.DownloadPackageDetailsStruct {

}

func uploadDownloadedPackages(downloadedPkgsArr [] helpers.DownloadPackageDetailsStruct) {
	helpers.LogInfo.Printf("Uploading downloaded %d packages", len(downloadedPkgsArr))
	for _, downloadedPkgStruct := range downloadedPkgsArr {
		
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
