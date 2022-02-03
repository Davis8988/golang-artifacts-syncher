package helpers

import (
    "golang-artifacts-syncher/src/nuget_packages_xml"
	"os"
	"sync"
	"fmt"
	"io/ioutil"
    "log"
    "net/http"
	"strings"
	"time"
	"strconv"
	"regexp"
    "path/filepath"
)


type HttpRequestArgsStruct struct {
	UrlAddress  string
	HeadersMap  map[string]string
    UserToUse  string
    PassToUse  string
    TimeoutSec  int
    Method  string
}

type NugetPackageDetailsStruct struct {
    Name string
    Version string
    Checksum string
    ChecksumType string
    PkgFileUrl string
}

type DownloadPackageDetailsStruct struct {
    PkgDetailsStruct NugetPackageDetailsStruct
    DownloadPath string
}



var (
    // Info writes logs in the color white
    LogInfo = log.New(os.Stdout, "\u001b[37m", log.LstdFlags)

    // Warning writes logs in the color yellow with "WARNING: " as prefix
    LogWarning = log.New(os.Stdout, "\u001b[33mWARNING: ", log.LstdFlags)

    // Error writes logs in the color red with " Error: " as prefix
    LogError = log.New(os.Stdout, "\u001b[35m Error: \u001B[31m", log.LstdFlags)

    // Debug writes logs in the color cyan with "Debug: " as prefix
    LogDebug = log.New(os.Stdout, "\u001b[36mDebug: ", log.LstdFlags)

    // Locks
    convertSyncedMapToString_Lock sync.RWMutex
    appendPkgDetailsArr_Lock sync.RWMutex
    appendDownloadedPkgDetailsArr_Lock sync.RWMutex
)

func Init() {
    LogInfo.Print("Initializing helpers pkg vars")
    convertSyncedMapToString_Lock = sync.RWMutex{}
    appendPkgDetailsArr_Lock = sync.RWMutex{}
    appendDownloadedPkgDetailsArr_Lock = sync.RWMutex{}
}

func trimQuotes(s string) string {
    if len(s) >= 2 {
        if s[0] == '"' && s[len(s)-1] == '"' {
            return s[1 : len(s)-1]
        }
    }
    return s
}

func GetCurrentProgramDir() string {
    ex, err := os.Executable()
    if err != nil {
        LogError.Fatalf("%s\nFailed getting current program's dir", err)
    }
    return filepath.Dir(ex)
}

// Attempts to resolve an environment variable, 
//  with a default value if it's empty
func Getenv(key, fallback string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        return fallback
    }
    return value
}

func StrToInt(strVar string) int {
	intVar, err := strconv.Atoi(strVar)
    if err != nil {
        LogError.Printf("%s\nFailed converting string: \"%s\" to integer", err, strVar)
        panic(err)
    }
    return intVar
}

func IsStrArrayEmpty(arrToCheck []string) bool {
	return len(arrToCheck) == 0
}

func LoadStringArrValueFromSynchedMap(synchedMap sync.Map, key string) [] string {
    currentInterfaceValue, _ := synchedMap.Load(key)
    var currentStrArr []string = currentInterfaceValue.([]string)
    return currentStrArr
}

func PrintSyncedMap(synchedMap sync.Map) {
	synchedMap.Range(func(key interface{}, value interface{}) bool {
		someVal, _ := synchedMap.Load(key)
		fmt.Println(someVal)
		return true
	})
}

func ConvertSyncedMapToString(synchedMap sync.Map) string {
	var result string
    synchedMap.Range(func(key interface{}, value interface{}) bool {
		currentInterfaceValue, _ := synchedMap.Load(key)
        var currentStrArr []string = currentInterfaceValue.([]string)
		keyStr := key.(string) // Convert to string
        result += keyStr + " : [" + strings.Join(currentStrArr, ", ") + "]\n"
        return true
	})
    return result
}

func ParseHttpHeadersStrToMap(httpRequestHeadersStr string) map[string]string {
    if len(httpRequestHeadersStr) <= 1 {return nil}
    
    httpRequestHeadersMap := make(map[string] string)
    tempHeadersPairsArr := make([]string, 0, 6)
    tempPairArr := make([]string, 0, 2)
    tempHeadersPairsArr = strings.Split(httpRequestHeadersStr, ";")
    LogInfo.Printf("Looping on headers values to init headers map")
    for _, headersPairStr := range tempHeadersPairsArr {
        tempPairArr = strings.Split(headersPairStr, "=")
        if len(tempPairArr) != 2 {
            LogError.Printf("Found header pair: \"%v\"  that is not in the right format of: \"key=value\"", tempPairArr)
            return nil
        }
        headerKey := tempPairArr[0]
        headerValue := tempPairArr[1]
        httpRequestHeadersMap[headerKey] = headerValue
    }
    return httpRequestHeadersMap
}

func MakeAnHttpRequest(httpRequestArgs HttpRequestArgsStruct) string {
    urlToCheck := httpRequestArgs.UrlAddress
    headersMap := httpRequestArgs.HeadersMap
    username := httpRequestArgs.UserToUse
    password := httpRequestArgs.PassToUse
    timeoutSec := httpRequestArgs.TimeoutSec
    method := httpRequestArgs.Method

    LogInfo.Printf("Querying URL: \"%s\"", urlToCheck)

    client := http.Client{
        Timeout: time.Duration(timeoutSec) * time.Second,
    }

    req, err := http.NewRequest(method, urlToCheck, nil)
    if err != nil {
        LogError.Printf("%s\nFailed creating HTTP request object for URL: \"%s\"", err, urlToCheck)
        return ""
    }

    // Adding headers:
    for k := range headersMap {
        LogInfo.Printf("Adding header:  '%s'=\"%s\"", k, headersMap[k])
        req.Header.Add(k, headersMap[k])
    }

    // Adding creds
    if len(username) > 0 && len(password) > 0 {
        LogInfo.Printf("Adding creds of user:  '%s'", username)
        req.SetBasicAuth(username, password)
    }

    // Make the http request
    response, err := client.Do(req)
    if err != nil {
        LogError.Printf("%s\nFailed querying: %s", err, urlToCheck)
        return ""
    }
  
    defer response.Body.Close() // Finally step: close the body obj
    
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        LogError.Printf("%s\nFailed querying: %s", err, urlToCheck)
        return ""
    }

    bodyStr := string(body)
    msgStr := bodyStr
    if len(response.Status) > 0 {msgStr = fmt.Sprintf("%s  %s", response.Status, bodyStr)}
    LogDebug.Printf(msgStr)

    if response.StatusCode >= 400 {LogError.Printf("Failed querying: %s", urlToCheck)}

    return bodyStr
}

func ParsePkgNameAndVersionFromFileURL(pkgFileUrl string) [] string {
    LogDebug.Printf("Parsing URL for Name & Version: \"%s\"", pkgFileUrl)
    re := regexp.MustCompile("'(.*?)'")  // Find values in between quotes
    resultArr := re.FindAllString(pkgFileUrl, -1)  // -1 = find ALL available matches
    if len(resultArr) != 2 {
        LogError.Printf("Failed to parse URL for pkg Name & Version:  \"%s\"", pkgFileUrl)
        LogError.Printf("Found regex result count is: %d different from 2", len(resultArr))
        return nil
    }
    // Trim
    for i, value := range resultArr {resultArr[i] = trimQuotes(value)}
    return resultArr
}


func ParseHttpRequestResponseForPackagesVersions(responseBody string) [] NugetPackageDetailsStruct {
    parsedPackagesVersionsArr := make([] NugetPackageDetailsStruct, 0)
    LogInfo.Printf("Parsing http request response for packages details")
    parsedPackagesDetailsStruct := nuget_packages_xml.ParseNugetPackagesXmlData(responseBody)
    for _, entryStruct := range parsedPackagesDetailsStruct.Entry {
        var pkgDetailsStruct NugetPackageDetailsStruct
        pkgDetailsStruct.Checksum = entryStruct.Properties.PackageHash
        pkgDetailsStruct.ChecksumType = entryStruct.Properties.PackageHashAlgorithm
        pkgDetailsStruct.PkgFileUrl = entryStruct.ID
        pkgDetailsStruct.Name = ""
        pkgDetailsStruct.Version = ""
        parsedNameAndVersionArr := ParsePkgNameAndVersionFromFileURL(pkgDetailsStruct.PkgFileUrl)
        if parsedNameAndVersionArr == nil {continue}
        pkgDetailsStruct.Name = parsedNameAndVersionArr[0]
        pkgDetailsStruct.Version = parsedNameAndVersionArr[1]
        parsedPackagesVersionsArr = append(parsedPackagesVersionsArr, pkgDetailsStruct)
    }
    return parsedPackagesVersionsArr
}

func SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs HttpRequestArgsStruct) [] NugetPackageDetailsStruct {
    responseBody := MakeAnHttpRequest(httpRequestArgs)
    if len(responseBody) == 0 {return nil}
    parsedPackagesDetailsArr := ParseHttpRequestResponseForPackagesVersions(responseBody)

    return parsedPackagesDetailsArr
}

func DownloadPkg(downloadPkgDetailsStruct DownloadPackageDetailsStruct) [] DownloadPackageDetailsStruct {
    LogInfo.Printf("Downloading pkg: %s-%s", downloadPkgDetailsStruct.PkgDetailsStruct.Name, downloadPkgDetailsStruct.PkgDetailsStruct.Version)
    return [] DownloadPackageDetailsStruct{downloadPkgDetailsStruct}
}

func Synched_ConvertSyncedMapToString(synchedMap sync.Map) string {
	convertSyncedMapToString_Lock.Lock()
	result := ConvertSyncedMapToString(synchedMap)
	defer convertSyncedMapToString_Lock.Unlock()
	return result
}

func Synched_AppendPkgDetailsObj(arr_1 *[] NugetPackageDetailsStruct, arr_2 [] NugetPackageDetailsStruct) {
    appendPkgDetailsArr_Lock.Lock()
    *arr_1 = append(*arr_1, arr_2...)
    appendPkgDetailsArr_Lock.Unlock()
}

func Synched_AppendDownloadedPkgDetailsObj(arr_1 *[] DownloadPackageDetailsStruct, arr_2 [] DownloadPackageDetailsStruct) {
    appendDownloadedPkgDetailsArr_Lock.Lock()
    *arr_1 = append(*arr_1, arr_2...)
    appendDownloadedPkgDetailsArr_Lock.Unlock()
}
