package global_structs

import (
    "github.com/hashicorp/go-version"
    "fmt"
    "reflect"
    "strings"
)

// AxisSorter sorts planets by Version.
type NugetPackageVersionSorter []NugetPackageDetailsStruct

func (a NugetPackageVersionSorter) Len() int           { return len(a) }
func (a NugetPackageVersionSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a NugetPackageVersionSorter) Less(i, j int) bool { 
    v1, err := version.NewVersion(a[i].Version)
    if err != nil {panic(fmt.Sprintf("%s\nFailed to parse version: '%s' during comparison func", err, a[i].Version))}
    v2, err := version.NewVersion(a[j].Version)
    if err != nil {panic(fmt.Sprintf("%s\nFailed to parse version: '%s' during comparison func", err, a[j].Version))}
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
}

type HttpResponseStruct struct {
    UrlAddress string
	BodyStr  string
	StatusStr  string
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

func (pkgDetailsStruct NugetPackageDetailsStruct) HashCode() string {
    return fmt.Sprintf("%s-%s", pkgDetailsStruct.Name, pkgDetailsStruct.Version)
}


type DownloadPackageDetailsStruct struct {
    PkgDetailsStruct NugetPackageDetailsStruct
    DownloadFilePath string
    DownloadFileChecksum  string
	DownloadFileChecksumType  string
	IsSuccessful  bool
}

type UploadPackageDetailsStruct struct {
    PkgDetailsStruct NugetPackageDetailsStruct
    UploadFilePath string
    UploadFileChecksum  string
	UploadFileChecksumType  string
    IsSuccessful  bool
}

type UploadedPackagesDataStruct struct {
    DestServerUrl string
    UploadedPackagesMap map[string] UploadPackageDetailsStruct  // 'name-version' -> UploadPackageDetailsStruct
}

type DownloadedPackagesDataStruct struct {
    SrcServerUrl string
    DownloadedPackagesMap map[string] DownloadPackageDetailsStruct  // 'name-version' -> DownloadPackageDetailsStruct
}

type AppConfiguration struct {
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
    HttpRequestGlobalDefaultTimeoutSecondsInt int
    HttpRequestDownloadTimeoutSecondsInt int
    HttpRequestUploadTimeoutSecondsInt int
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
}

func (appConfig AppConfiguration) ToString() string {
    v := reflect.ValueOf(appConfig)
    // fieldsArr := make([]interface{}, v.NumField())
    resultStr := ""
    typeOfObj := v.Type()
    passwordFieldsNamesMap := make(map[string] string)
    passwordFieldsNamesMap["SrcServersPassToUse"] = "SrcServersPassToUse"
    passwordFieldsNamesMap["DestServersPassToUse"] = "DestServersPassToUse"
    for i := 0; i< v.NumField(); i++ {
        fieldName := typeOfObj.Field(i).Name
        fieldValue := fmt.Sprintf("%v", v.Field(i).Interface())
        if _, isMapContainsKey := passwordFieldsNamesMap[fieldName]; isMapContainsKey {
            fieldValue = strings.Repeat("*", len(fieldValue))
        }

        resultStr += fmt.Sprintf("  %s : %v\n", fieldName, fieldValue)
    }
    return resultStr
}

