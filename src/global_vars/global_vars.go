package global_vars

import (
	"sync"
)

var (
    // Locks
    ConvertSyncedMapToString_Lock sync.RWMutex
    AppendPkgDetailsArr_Lock sync.RWMutex
    AppendDownloadedPkgDetailsArr_Lock sync.RWMutex

    SrcServersUserToUse          string
    SrcServersPassToUse          string
    SrcServersUrlsStr            string
    SrcReposNamesStr             string
    DestServersUrlsStr           string
    DestReposNamesStr            string
    DestServersUserToUse         string
    DestServersPassToUse         string
    PackagesNamesStr             string
    PackagesVersionsStr          string
    HttpRequestHeadersStr        string
    DownloadPkgsDirPath          string
    HttpRequestTimeoutSecondsInt int

    SrcServersUrlsArr     []string
    SrcReposNamesArr      []string
    DestServersUrlsArr    []string
    DestReposNamesArr     []string
    PackagesNamesArr      []string
    PackagesVersionsArr   []string
    HttpRequestHeadersMap map[string]string
    PackagesToDownloadMap sync.Map
)
