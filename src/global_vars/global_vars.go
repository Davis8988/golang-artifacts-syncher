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

    // Success Indicator File Path
    SuccessIndicatorFile string

    PackagesToDownloadMap sync.Map

    // Slices
    DownloadedPkgsMap   [] global_structs.DownloadedPackagesDataStruct
    UploadedPkgsDataArr [] global_structs.UploadedPackagesDataStruct

    ErrorsDetected bool
)
