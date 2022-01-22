package main

import (
	"flag"
	"os"
	"log"
	"strings"
)

var (
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

func validateEnv() {
	log.Print("Validating envs")
	log.Print("Comparing packages names & versions arrays lengths")
	
	
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
}

func prepareSearchUrlsArray() []string {
	log.Printf("Preparing search packages urls")
	searchOptionsUrl := "Search()?"
	for _, serverUrl := range serversUrlsArr {
		for _, repoName := range reposNamesArr {
			for _, pkgName := range packagesNamesArr {
				searchUrlsArr = append(searchUrlsArr, serverUrl + "/" + repoName + "/" + searchOptionsUrl + "id='" + pkgName + "'")
			}
		}
	}
}

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
