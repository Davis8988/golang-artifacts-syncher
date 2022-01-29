package main

import (
	"golang-artifacts-syncher/src/helpers"
	"golang-artifacts-syncher/src/nexus3_adapter"
	"errors"
	"flag"
	"os"
	"sync"
	"strings"
)

var (
	userToUse string
	passToUse string
	serversUrlsStr string
	reposNamesStr string
	packagesNamesStr string
	packagesVersionsStr string
	httpRequestHeadersStr string
	httpRequestTimeoutSeconds int

	serversUrlsArr [] string
	reposNamesArr [] string
	packagesNamesArr [] string
	packagesVersionsArr [] string
	httpRequestHeadersMap map[string] string
	packagesToDownloadMap sync.Map

	lock sync.RWMutex
)

func synched_convertSyncedMapToString(synchedMap sync.Map) string {
	lock.Lock()
	result := helpers.ConvertSyncedMapToString(synchedMap)
	lock.Unlock()
	return result
}


func initVars() {
	helpers.LogInfo.Print("Initializing from envs vars")
	userToUse = helpers.Getenv("USER_TO_USE", "")
	passToUse = helpers.Getenv("PASS_TO_USE", "")
	serversUrlsStr = helpers.Getenv("SERVERS_URLS_STR", "")
	reposNamesStr = helpers.Getenv("REPOS_NAMES_STR", "")
	packagesNamesStr = helpers.Getenv("PACKAGES_NAMES_STR", "")
	packagesVersionsStr = helpers.Getenv("PACKAGES_VERSIONS_STR", "")
	httpRequestHeadersStr = helpers.Getenv("HTTP_REQUEST_HEADERS_STR", "")  // Example: "key=value;key1=value1;key2=value2"
	httpRequestTimeoutStr = helpers.Getenv("HTTP_REQUEST_TIMEOUT_SECONDS_STR", "45")  
	// packagesToDownloadMap = make(map[string][] string)
	lock = sync.RWMutex{}
}

func printVars() {
	helpers.LogInfo.Printf("SERVERS_URLS_STR: '%s'", serversUrlsStr)
	helpers.LogInfo.Printf("REPOS_NAMES_STR: '%s'", reposNamesStr)
	helpers.LogInfo.Printf("PACKAGES_NAMES_STR: '%s'", packagesNamesStr)
	helpers.LogInfo.Printf("PACKAGES_VERSIONS_STR: '%s'", packagesVersionsStr)
	helpers.LogInfo.Printf("HTTP_REQUEST_HEADERS_STR: '%s'", httpRequestHeadersStr)
	helpers.LogInfo.Printf("HTTP_REQUEST_TIMEOUT_SECONDS_STR: '%s'", httpRequestTimeoutStr)
	
	helpers.LogInfo.Printf("serversUrlsArr: %v", serversUrlsArr)
	helpers.LogInfo.Printf("reposNamesArr: %v", reposNamesArr)
	helpers.LogInfo.Printf("packagesNamesArr: %v", packagesNamesArr)
	helpers.LogInfo.Printf("packagesVersionsArr: %v", packagesVersionsArr)
	packagesToDownloadMapStr := synched_convertSyncedMapToString(packagesToDownloadMap)
	helpers.LogInfo.Printf("packagesToDownloadMap: \n%v", packagesToDownloadMapStr)
}

func abortWithError(errMsg string, exitCode int) {
	e := errors.New(errMsg)
	helpers.LogError.Printf("%s", e)
	helpers.LogError.Printf("Aborting with exit code: %d", exitCode)
	os.Exit(exitCode)
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
			abortWithError(errMsg, 1)
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
	serversUrlsArr = make([]string, 0, 4)
	reposNamesArr = make([]string, 0, 4)
	packagesNamesArr = make([]string, 0, 10)
	packagesVersionsArr = make([]string, 0, 10)
	if len(serversUrlsStr) > 1 {serversUrlsArr = strings.Split(serversUrlsStr, ";")}
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
	
	helpers.LogInfo.Print("Preparing search packages urls array")
	searchOptionsUrl := "Search()?"
	for _, serverUrl := range serversUrlsArr {
		for _, repoName := range reposNamesArr {
			for _, pkgName := range packagesNamesArr {
				versionsToSearchArr := helpers.LoadStringArrValueFromSynchedMap(packagesToDownloadMap, pkgName)
				if len(versionsToSearchArr) > 0 {continue}
				searchUrlsArr = append(searchUrlsArr, serverUrl + "/" + repoName + "/" + searchOptionsUrl + "id='" + pkgName + "'")
			}
		}
	}
	return searchUrlsArr
} 

func searchAvailableVersionsOfSpecifiedPackages() []string {
	var foundPackagesNamesArr []string
	searchUrlsArr := prepareSearchAllPkgsVersionsUrlsArray()
	if len(searchUrlsArr) > 0 {
		helpers.LogInfo.Printf("Checking %d URL addresses for pkgs versions", len(searchUrlsArr))
		for _, urlToCheck := range searchUrlsArr {
			httpRequestArgs := helpers.HttpRequestArgsStruct {
				UrlAddress: urlToCheck,
				HeadersMap: httpRequestHeadersMap,
				UserToUse: userToUse,
				PassToUse: passToUse,
			}
			helpers.SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs)
		}
	}

	return foundPackagesNamesArr 
}

func downloadSpecifiedPackages() {
	foundPackagesArr := searchAvailableVersionsOfSpecifiedPackages()
	helpers.LogInfo.Printf("Found packages: %v", foundPackagesArr)
}

func uploadDownloadedPackages() {

}

func main() {
	helpers.LogInfo.Print("Started")
	initVars()
	parseArgs()
	updateVars()
	printVars()
	validateEnv()
	downloadSpecifiedPackages()
	uploadDownloadedPackages()
	helpers.LogInfo.Print("Finished")
}
