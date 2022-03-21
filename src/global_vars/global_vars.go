package global_vars

import (
    "golang-artifacts-syncher/src/global_structs"
	"sync"
)

var (
    // Log
    DefaultLogLevel = "INFO"

    // Locks
    ConvertSyncedMapToString_Lock sync.RWMutex
    JoinTwoPkgDetailsSlices_Lock sync.RWMutex
    JoinTwoPkgDetailsMaps_Lock sync.RWMutex
    AppendDownloadedPkgDetailsArr_Lock sync.RWMutex
    AppendUploadedPkgDetailsArr_Lock sync.RWMutex
    AppendPkgDetailsArr_Lock sync.RWMutex
    AppendPkgDetailsMap_Lock sync.RWMutex
    ErrorsDetected_Lock sync.RWMutex

    // Configration of the app
    AppConfig global_structs.AppConfiguration

    PackagesToDownloadMap sync.Map

    ErrorsDetected bool
)
