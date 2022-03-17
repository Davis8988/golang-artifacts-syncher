package global_vars

import (
	"sync"
)

var (
    // Locks
    ConvertSyncedMapToString_Lock sync.RWMutex
    JoinTwoPkgDetailsSlices_Lock sync.RWMutex
    JoinTwoPkgDetailsMaps_Lock sync.RWMutex
    AppendDownloadedPkgDetailsArr_Lock sync.RWMutex
    AppendPkgDetailsArr_Lock sync.RWMutex
    AppendPkgDetailsMap_Lock sync.RWMutex

    SrcServersUserToUse          string
    SrcServersPassToUse          string
    SrcServersUrlsStr            string
    DestServersUrlsStr           string
    DestServersUserToUse         string
    DestServersPassToUse         string
    PackagesNamesStr             string
    PackagesVersionsStr          string
    HttpRequestHeadersStr        string
    DownloadPkgsDirPath          string
    LogLevel                     string
    HttpRequestTimeoutSecondsInt       int
    SearchPackagesUrlSkipGroupCount    int  // Used for URL searching requests of Nuget pkgs - Can't query for all at once, need to query multiple times and skip previous results.
    PackagesDownloadLimitCount         int  
    PackagesMaxConcurrentDownloadCount int  
    PackagesMaxConcurrentUploadCount   int  
    PackagesMaxConcurrentDeleteCount   int  

    SrcServersUrlsArr     []string
    DestServersUrlsArr    []string
    PackagesNamesArr      []string
    PackagesVersionsArr   []string
    HttpRequestHeadersMap map[string]string
    PackagesToDownloadMap sync.Map
)
