package global_structs


type HttpRequestArgsStruct struct {
	UrlAddress  string
	DownloadFilePath  string
	UploadFilePath  string
	HeadersMap  map[string]string
    UserToUse  string
    PassToUse  string
    TimeoutSec  int
    Method  string
    SkipErrorsPrintOnReceivedHttpCode *int
}

type NugetPackageDetailsStruct struct {
    Name string
    Version string
    Checksum string
    ChecksumType string
    PkgDetailsUrl string
    PkgFileUrl string
}

type DownloadPackageDetailsStruct struct {
    PkgDetailsStruct NugetPackageDetailsStruct
    DownloadFilePath string
    DownloadFileChecksum  string
	DownloadFileChecksumType  string
}

type UploadPackageDetailsStruct struct {
    PkgDetailsStruct NugetPackageDetailsStruct
    UploadFilePath string
    UploadFileChecksum  string
	UploadFileChecksumType  string
}


