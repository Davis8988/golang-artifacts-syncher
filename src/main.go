package main

import (
	"flag"
	"os"
	"fmt"
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

func main() {
	flag.Parse()
	initVars()
	fmt.Println("Hello World")
	fmt.Println("SERVERS_URLS_STR: '{}'", serversUrlsStr)
	fmt.Println("REPOS_NAMES_STR: '{}'", reposNamesStr)
	fmt.Println("PACKAGES_NAMES_STR: '{}'", packagesNamesStr)
	fmt.Println("PACKAGES_VERSIONS_STR: '{}'", packagesVersionsStr)

}
