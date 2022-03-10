package global_structs

import (
    "github.com/hashicorp/go-version"
    "fmt"
)

// AxisSorter sorts planets by Version.
type NugetPackageVersionSorter []NugetPackageDetailsStruct

func (a NugetPackageVersionSorter) Len() int           { return len(a) }
func (a NugetPackageVersionSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a NugetPackageVersionSorter) Less(i, j int) bool { 
    v1, err := version.NewVersion(a[i].Version)
    if err != nil {panic(fmt.Sprintf("Failed to parse version: '%s' during comparison func", a[i].Version))}
    v2, err := version.NewVersion(a[j].Version)
    if err != nil {panic(fmt.Sprintf("Failed to parse version: '%s' during comparison func", a[j].Version))}
    return v1.LessThan(v2)  // Sorts so that first is the lowest
    // return v1.GreaterThan(v2)  // Sorts so that first is the greatest
}

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

type HttpResponseStruct struct {
	BodyStr  string
	StatusCode  int
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

