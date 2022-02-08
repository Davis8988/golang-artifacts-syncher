package global_vars

import (
	"sync"
)

var (
    // Locks
    convertSyncedMapToString_Lock sync.RWMutex
    appendPkgDetailsArr_Lock sync.RWMutex
    appendDownloadedPkgDetailsArr_Lock sync.RWMutex

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
    DownloadPkgsDirPath          string
    httpRequestTimeoutSecondsInt int

    srcServersUrlsArr     []string
    srcReposNamesArr      []string
    DestServersUrlsArr    []string
    destReposNamesArr     []string
    packagesNamesArr      []string
    packagesVersionsArr   []string
    httpRequestHeadersMap map[string]string
    packagesToDownloadMap sync.Map
)
