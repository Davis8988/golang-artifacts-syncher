package main

import (
	"golang-artifacts-syncher/src/helpers"
	"golang-artifacts-syncher/src/nexus3_adapter"
	"errors"
	"flag"
	"os"
	"sync"
	"log"
	"strings"
)

var (
	// Info writes logs in the color white
	LogInfo = log.New(os.Stdout, "\u001b[37m", log.LstdFlags)

	// Warning writes logs in the color yellow with "WARNING: " as prefix
	LogWarning = log.New(os.Stdout, "\u001b[33mWARNING: ", log.LstdFlags)

	// Error writes logs in the color red with " Error: " as prefix
	LogError = log.New(os.Stdout, "\u001b[35m Error: \u001B[31m", log.LstdFlags)

	// Debug writes logs in the color cyan with "Debug: " as prefix
	LogDebug = log.New(os.Stdout, "\u001b[36mDebug: ", log.LstdFlags)

	userToUse string
	passToUse string
	serversUrlsStr string
	reposNamesStr string
	packagesNamesStr string
	packagesVersionsStr string

	serversUrlsArr [] string
	reposNamesArr [] string
	packagesNamesArr [] string
	packagesVersionsArr [] string
	// packagesToDownloadMap map[string][] string
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
	LogInfo.Print("Initializing from envs vars")
	userToUse = helpers.Getenv("USER_TO_USE", "")
	passToUse = helpers.Getenv("PASS_TO_USE", "")
	serversUrlsStr = helpers.Getenv("SERVERS_URLS_STR", "")
	reposNamesStr = helpers.Getenv("REPOS_NAMES_STR", "")
	packagesNamesStr = helpers.Getenv("PACKAGES_NAMES_STR", "")
	packagesVersionsStr = helpers.Getenv("PACKAGES_VERSIONS_STR", "")
	// packagesToDownloadMap = make(map[string][] string)
	lock = sync.RWMutex{}
}

func printVars() {
	LogInfo.Printf("SERVERS_URLS_STR: '%s'", serversUrlsStr)
	LogInfo.Printf("REPOS_NAMES_STR: '%s'", reposNamesStr)
	LogInfo.Printf("PACKAGES_NAMES_STR: '%s'", packagesNamesStr)
	LogInfo.Printf("PACKAGES_VERSIONS_STR: '%s'", packagesVersionsStr)
	
	LogInfo.Printf("serversUrlsArr: %v", serversUrlsArr)
	LogInfo.Printf("reposNamesArr: %v", reposNamesArr)
	LogInfo.Printf("packagesNamesArr: %v", packagesNamesArr)
	LogInfo.Printf("packagesVersionsArr: %v", packagesVersionsArr)
	packagesToDownloadMapStr := synched_convertSyncedMapToString(packagesToDownloadMap)
	LogInfo.Printf("packagesToDownloadMap: \n%v", packagesToDownloadMapStr)
}

func abortWithError(errMsg string, exitCode int) {
	e := errors.New(errMsg)
	LogError.Printf("%s", e)
	LogError.Printf("Aborting with exit code: %d", exitCode)
	os.Exit(exitCode)
}

func validateEnv() {
	LogInfo.Print("Validating envs")

	// Validate len(packagesVersionsArr) == len(packagesNamesArr)  (Only when packagesVersionsArr is defined)
	if ! nexus3_adapter.IsStrArrayEmpty(packagesVersionsArr) {
		LogInfo.Print("Comparing packages names & versions arrays lengths")
		if len(packagesVersionsArr) != len(packagesNamesArr) {
			errMsg := "Packages Versions to search count is different from Packages Names to search count\n"
			errMsg += "Can't search for packages versions & names which are not of the same count.\n"
			errMsg += "When passing packages versions to search - the versions count must be of the same count of packages names to search.\n"
			errMsg += "A version for each package name to search"
			abortWithError(errMsg, 1)
		}
	}
	
	LogInfo.Print("All Good")
}

func parseArgs() {
	LogInfo.Print("Parsing args")
	flag.Parse()
}

func updateVars() {
	LogInfo.Print("Updating vars")
	serversUrlsArr = strings.Split(serversUrlsStr, ";")
	reposNamesArr = strings.Split(reposNamesStr, ";")
	packagesNamesArr = strings.Split(packagesNamesStr, ";")
	packagesVersionsArr = strings.Split(packagesVersionsStr, ";")

	for i, pkgName := range packagesNamesArr {
		// If map doesn't contain value at: 'pkgName' - add one to point to empty string array: []
		packagesToDownloadMap.LoadOrStore(pkgName, make([] string, 0))
		// If received a version array for it - add it to the list
		if len(packagesVersionsArr) >= i {
			pkgVersion := packagesVersionsArr[i]
			currentVersionsArr := helpers.LoadStringArrValueFromSynchedMap(packagesToDownloadMap, pkgName)
			packagesToDownloadMap.Store(pkgName, append(currentVersionsArr, pkgVersion))
		}
	}
}

// func prepareSearchUrlsArray() []string {
	
// } 

func searchSpecifiedPackages() []string {
	var foundPackagesArr []string
	var searchUrlsArr = make([]string, 0, 10)  // Create a slice with length=0 and capacity=10
	
	LogInfo.Print("Preparing search packages urls array")
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



	return foundPackagesArr 
}

func downloadSpecifiedPackages() {
	foundPackagesArr := searchSpecifiedPackages()
	LogInfo.Printf("Found packages: %v", foundPackagesArr)
}

func uploadDownloadedPackages() {

}

func main() {
	LogInfo.Print("Started")
	initVars()
	parseArgs()
	updateVars()
	printVars()
	validateEnv()
	downloadSpecifiedPackages()
	uploadDownloadedPackages()
	LogInfo.Print("Finished")
}
