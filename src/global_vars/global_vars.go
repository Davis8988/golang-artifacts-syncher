package global_vars


var (
    // Log
    LogInfo = log.New(os.Stdout, "\u001b[37m", log.LstdFlags)
    LogWarning = log.New(os.Stdout, "\u001b[33mWARNING: ", log.LstdFlags)
    LogError = log.New(os.Stdout, "\u001b[35m Error: \u001B[31m", log.LstdFlags)
    LogDebug = log.New(os.Stdout, "\u001b[36mDebug: ", log.LstdFlags)


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
