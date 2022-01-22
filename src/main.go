package main

import (
	"errors"
	"flag"
	"os"
	"log"
	"strings"
)

var (
	// Info writes logs in the color blue with "INFO: " as prefix
	Info = log.New(os.Stdout, "\u001b[34mINFO: \u001B[0m", log.LstdFlags|log.Lshortfile)

	// Warning writes logs in the color yellow with "WARNING: " as prefix
	Warning = log.New(os.Stdout, "\u001b[33mWARNING: \u001B[0m", log.LstdFlags|log.Lshortfile)

	// Error writes logs in the color red with "ERROR: " as prefix
	Error = log.New(os.Stdout, "\u001b[31mERROR: \u001b[0m", log.LstdFlags|log.Lshortfile)

	// Debug writes logs in the color cyan with "DEBUG: " as prefix
	Debug = log.New(os.Stdout, "\u001b[36mDEBUG: \u001B[0m", log.LstdFlags|log.Lshortfile)

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
)

// Attempts to resolve an environment variable, 
//  with a default value if it's empty
func getenv(key, fallback string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        return fallback
    }
    return value
}

func initVars() {
	log.Print("Initializing from envs vars")
	userToUse = getenv("USER_TO_USE", "")
	passToUse = getenv("PASS_TO_USE", "")
	serversUrlsStr = getenv("SERVERS_URLS_STR", "")
	reposNamesStr = getenv("REPOS_NAMES_STR", "")
	packagesNamesStr = getenv("PACKAGES_NAMES_STR", "")
	packagesVersionsStr = getenv("PACKAGES_VERSIONS_STR", "")
}

func printVars() {
	log.Printf("SERVERS_URLS_STR: '%s'", serversUrlsStr)
	log.Printf("REPOS_NAMES_STR: '%s'", reposNamesStr)
	log.Printf("PACKAGES_NAMES_STR: '%s'", packagesNamesStr)
	log.Printf("PACKAGES_VERSIONS_STR: '%s'", packagesVersionsStr)
	
	log.Printf("serversUrlsArr: %v", serversUrlsArr)
	log.Printf("reposNamesArr: %v", reposNamesArr)
	log.Printf("packagesNamesArr: %v", packagesNamesArr)
	log.Printf("packagesVersionsArr: %v", packagesVersionsArr)
}

func abortWithError(errMsg string, exitCode int) {
	e := errors.New(errMsg)
	log.Printf("Error")
	log.Printf("%s", e)
	log.Printf("Aborting with exit code: %d", exitCode)
	os.Exit(exitCode)
}

func validateEnv() {
	log.Print("Validating envs")

	// Validate len(packagesVersionsArr) == len(packagesNamesArr)  (Only when packagesVersionsArr is defined)
	if len(packagesVersionsArr) > 0 {
		log.Print("Comparing packages names & versions arrays lengths")
		if len(packagesVersionsArr) != len(packagesNamesArr) {
			errMsg := "Packages Versions to search count is different from Packages Names to search count\n"
			errMsg += "Can't search for packages versions & names which are not of the same count.\n"
			errMsg += "When passing packages versions to search - the versions count must be of the same count of packages names to search.\n"
			errMsg += "A version for each package name to search"
			abortWithError(errMsg, 1)
		}
	}
	
	log.Print("Done. OK")
}

func parseArgs() {
	log.Print("Parsing args")
	flag.Parse()
}

func updateVars() {
	log.Print("Updating vars")
	serversUrlsArr = strings.Split(serversUrlsStr, ";")
	reposNamesArr = strings.Split(reposNamesStr, ";")
	packagesNamesArr = strings.Split(packagesNamesStr, ";")
	packagesVersionsArr = strings.Split(packagesVersionsStr, ";")
}

/* func prepareSearchUrlsArray() []string {
	log.Printf("Preparing search packages urls")
	searchOptionsUrl := "Search()?"
	for _, serverUrl := range serversUrlsArr {
		for _, repoName := range reposNamesArr {
			for _, pkgName := range packagesNamesArr {
				searchUrlsArr = append(searchUrlsArr, serverUrl + "/" + repoName + "/" + searchOptionsUrl + "id='" + pkgName + "'")
			}
		}
	}
} */

func searchSpecifiedPackages() []string {
	var foundPackagesArr []string
	var searchUrlsArr []string
	
	log.Printf("Search array: %v", searchUrlsArr)
	return foundPackagesArr 
}

func downloadSpecifiedPackages() {
	foundPackagesArr := searchSpecifiedPackages()
	log.Printf("Found packages: %v", foundPackagesArr)
}

func uploadDownloadedPackages() {

}

func main() {
	log.Print("Started")
	initVars()
	parseArgs()
	updateVars()
	printVars()
	validateEnv()
	downloadSpecifiedPackages()
	uploadDownloadedPackages()
	log.Print("Finished")
}
