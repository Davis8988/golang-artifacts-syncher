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
)


type HttpRequestArgsStruct struct {
	UrlAddress  string
	HeadersMap  map[string]string
    UserToUse  string
    PassToUse  string
    TimeoutSec  int
    Method  string
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
)

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

func PraseHttpRequestResponseForPackagesVersions(responseBody string) map[string] [] string {
    parsedPackagesVersionsMap := make(map[string] [] string, 10)
    LogInfo.Printf("Parsing http request response for packages details")
    parsedPackagesDetailsStruct := nuget_packages_xml.ParseNugetPackagesXmlData(responseBody)
    LogInfo.Printf("%s", parsedPackagesDetailsStruct)
    return parsedPackagesVersionsMap
}

func SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs HttpRequestArgsStruct) [] string {
    
    responseBody := MakeAnHttpRequest(httpRequestArgs)
    if len(responseBody) == 0 {return nil}

    packagesAvailableVersions := make([] string, 0, 10)

    return packagesAvailableVersions
}
