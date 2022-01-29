package helpers

import (
	"os"
	"sync"
	"fmt"
	"io/ioutil"
    "log"
    "net/http"
	"strings"
)

type HttpRequestArgs struct {
	urlAddress  string
	headersMap  map[string]string
    userToUse  string
    passToUse  string
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
    tempHeadersPairsArr := make([]string, 0, 6)
    tempPairArr := make([]string, 0, 2)
    tempHeadersPairsArr = strings.Split(httpRequestHeadersStr, ";")
    LogInfo.Printf("Looping on headers values to init headers map")
    for _, headersPairStr := range tempHeadersPairsArr {
        tempPairArr := strings.Split(headersPairStr, "=")
        if len(tempPairArr) != 2 {
            LogError.Printf("Found header pair: \"%v\"  that is not in the right format of: \"key=value\"", tempPairArr)
            return nil
        }
        headerKey := tempPairArr[0]
        headerValue := tempPairArr[1]
    }
}

func SearchPackagesAvailableVersionsByURLRequest(argsMap map[string]string) [] string {
    urlToCheck := argsMap["HTTP_REQUEST_URL"]
    headers := ""
    if v, found := m["pi"]; found {
        headers := argsMap["HTTP_REQUEST_URL"]
    }
    packagesAavilableVersions := make([] string, 0, 10)
    LogInfo.Printf("Querying URL: \"%s\"", urlToCheck)

    resp, err := http.Get(urlToCheck)
    if err != nil {
        LogError.Printf("%s\nFailed querying: %s", err, urlToCheck)
        return nil
    }
  
    defer resp.Body.Close()  // Finally step: close the body obj
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        LogError.Printf("%s\nFailed querying: %s", err, urlToCheck)
        return nil
    }
  
    LogInfo.Printf(string(body))

    return packagesAavilableVersions
}
