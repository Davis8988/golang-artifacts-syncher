package main

import (
	"flag"
	"os"
	"log"
)

var (
	serversUrlsStr = ""
	reposNamesStr = ""
	packagesNamesStr = ""
	packagesVersionsStr = ""
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
	flag.Parse()
}

func main() {
	initVars()
	parseArgs()
	log.Print("Started")
	

}
