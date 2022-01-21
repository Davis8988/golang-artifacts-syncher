package main

import (
	"flag"
	"os"
	"log"
	"strings"
)

var (
	serversUrlsStr = ""
	reposNamesStr = ""
	packagesNamesStr = ""
	packagesVersionsStr = ""

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
	log.Print("Initializing vars")
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
}

func parseArgs() {
	log.Print("Parsing args")
	flag.Parse()
}

func updateVars() {
	log.Print("Updating vars")
	serversUrlsArr = strings.Split(";", serversUrlsStr)
	reposNamesArr = strings.Split(";", reposNamesStr)
	packagesNamesArr = strings.Split(";", packagesNamesStr)
}

func main() {
	log.Print("Started")
	initVars()
	parseArgs()
	updateVars()
	printVars()
	
	log.Print("Finished")
}
