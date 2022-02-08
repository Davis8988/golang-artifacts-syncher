package helpers

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"golang-artifacts-syncher/src/nuget_packages_xml"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)


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

func TrimQuotes(s string) string {
    if len(s) >= 2 {
        if c := s[len(s)-1]; s[0] == c && (c == '"' || c == '\'') {
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
    if currentInterfaceValue == nil {return []string{}}  // Return empty
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
    tempHeadersPairsArr := strings.Split(httpRequestHeadersStr, ";")
    LogInfo.Printf("Looping on headers values to init headers map")
    for _, headersPairStr := range tempHeadersPairsArr {
        tempPairArr := strings.Split(headersPairStr, "=")
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

func CreateDir(dirPath string) {
    if _, err := os.Stat(dirPath); err == nil {return}  // If dir already exists - finish here
    LogDebug.Printf("Creating dir: %s", dirPath)
    err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		LogError.Printf("%s\nFailed creating dir: \"%s\"", err, dirPath)
        panic(err)
	}
}

func CreateFile(filePath string) *os.File {
    dirPath := filepath.Dir(filePath)
    CreateDir(dirPath)
    LogDebug.Printf("Creating file: %s", filePath)
    // Create the file
    file, err := os.Create(filePath)
    if err != nil  {
        LogError.Printf("%s\nFailed creating file: \"%s\"", err, filePath)
        panic(err)
    }
    return file
}

func CalculateFileChecksum(filePath string) string {
    if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {return ""}  // If missing file: return empty
    LogDebug.Printf("Calculating sha512 checksum of file: %s", filePath)
    f, err := os.Open(filePath)
    if err != nil {
        LogError.Printf("%s\nFailed calculating sha512 checksum of file: \"%s\"", err, filePath)
        panic(err)
    }
    defer f.Close()
    h := sha512.New()
    if _, err := io.Copy(h, f); err != nil {
        LogError.Printf("%s\nFailed calculating sha512 checksum of file: \"%s\"", err, filePath)
        panic(err)
    }

    return base64.StdEncoding.EncodeToString(h.Sum(nil))  // nil instead of a byte array to append to
}

func MakeHttpRequest(httpRequestArgs HttpRequestArgsStruct) string {
    urlAddress := httpRequestArgs.UrlAddress
    downloadFilePath := httpRequestArgs.DownloadFilePath
    uploadFilePath := httpRequestArgs.UploadFilePath
    headersMap := httpRequestArgs.HeadersMap
    username := httpRequestArgs.UserToUse
    password := httpRequestArgs.PassToUse
    timeoutSec := httpRequestArgs.TimeoutSec
    method := strings.ToUpper(httpRequestArgs.Method)

    LogInfo.Printf("Making an HTTP '%s' request to URL: \"%s\"", method, urlAddress)

    client := http.Client{Timeout: time.Duration(timeoutSec) * time.Second,}
    
    var body io.Reader
    var writer multipart.Writer

    // Upload file (PUT requests):
    if method == "PUT" && len(uploadFilePath) > 0 {
        body = ReadFileContentsIntoPartsForUpload(uploadFilePath, "package")
    }

    req, err := http.NewRequest(method, urlAddress, body)
    if err != nil {
        LogError.Printf("%s\nFailed creating HTTP request object for URL: \"%s\"", err, urlAddress)
        return ""
    }

    // Incase pushing a file, then add the Content Type header from the reader (includes boundary)
    if method == "PUT" && len(uploadFilePath) > 0 {
        LogError.Printf("Adding header:  'Content-Type'")
        req.Header.Add("Content-Type", writer.FormDataContentType())
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
        LogError.Printf("%s\nFailed while making the request: %s", err, urlAddress)
        return ""
    }
  
    defer response.Body.Close() // Finally step: close the body obj
    
    // If got: downloadFilePath var, then Writer the body to file
    if len(downloadFilePath) > 0 {
        LogInfo.Printf("Downloading '%s' to:  %s", urlAddress, downloadFilePath)
        fileHandle := CreateFile(downloadFilePath)  // Create the file
        defer fileHandle.Close()

        _, err = io.Copy(fileHandle, response.Body)
        if err != nil  {
            LogError.Printf("%s\nFailed writing response Body to file: %s", err, downloadFilePath)
            panic(err)
        }
        return "" // Finish here
    }

    responseBody, err := ioutil.ReadAll(response.Body)
    if err != nil {
        LogError.Printf("%s\nFailed reading request's response body: %s", err, urlAddress)
        return ""
    }

    bodyStr := string(responseBody)
    msgStr := bodyStr
    if len(response.Status) > 0 {msgStr = fmt.Sprintf("%s  %s", response.Status, bodyStr)}
    // LogDebug.Printf(msgStr)

    if response.StatusCode >= 400 {
        LogError.Printf("%s", msgStr)
        LogError.Printf("Returned code: %d. HTTP request failure: %s", response.StatusCode, urlAddress)
    }

    return bodyStr
}

func ReadFileContentsIntoPartsForUpload(uploadFilePath string, headerFieldName string) (io.Reader, *multipart.Writer) {
    LogInfo.Printf("Reading file content for upload: \"%s\"", uploadFilePath)

    // If missing file: return empty body
    if _, err := os.Stat(uploadFilePath); errors.Is(err, os.ErrNotExist) {
        LogError.Printf("%s\nFailed uploading file: \"%s\" since it is missing. Failed preparing HTTP request object", err, uploadFilePath)
        return nil, nil
    }

    file, err := os.Open(uploadFilePath)
	if err != nil {
		return nil, nil
	}

    fileContents, err := ioutil.ReadAll(file)
    if err != nil {
		return nil, nil
	}

    fi, err := file.Stat()
	if err != nil {
		return nil, nil
	}
    file.Close()

    body := new(bytes.Buffer)
    writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(headerFieldName, fi.Name())  // Use package as headerFieldName for Nuget packages files upload
	if err != nil {
		return nil, nil
	}
	part.Write(fileContents)
    
    err = writer.Close()
	if err != nil {
		return nil, nil
	}

    return body, writer
}

func ParsePkgNameAndVersionFromFileURL(pkgDetailsUrl string) [] string {
    LogDebug.Printf("Parsing URL for Name & Version: \"%s\"", pkgDetailsUrl)
    re := regexp.MustCompile("'(.*?)'")  // Find values in between quotes
    resultArr := re.FindAllString(pkgDetailsUrl, -1)  // -1 = find ALL available matches
    if len(resultArr) != 2 {
        LogError.Printf("Failed to parse URL for pkg Name & Version:  \"%s\"", pkgDetailsUrl)
        LogError.Printf("Found regex result count is: %d different from 2", len(resultArr))
        return nil
    }
    // Trim
    for i, value := range resultArr {resultArr[i] = TrimQuotes(value)}
    return resultArr
}

func ParseXmlDataToSinglePkgDetailsStruct(entryStruct nuget_packages_xml.SinglePackagesDetailsXmlStruct) NugetPackageDetailsStruct {
    var pkgDetailsStruct NugetPackageDetailsStruct
    pkgDetailsStruct.Checksum = entryStruct.Properties.PackageHash
    pkgDetailsStruct.ChecksumType = entryStruct.Properties.PackageHashAlgorithm
    pkgDetailsStruct.PkgDetailsUrl = entryStruct.ID
    pkgDetailsStruct.PkgFileUrl = entryStruct.Content.Src
    pkgDetailsStruct.Name = ""
    pkgDetailsStruct.Version = ""
    parsedNameAndVersionArr := ParsePkgNameAndVersionFromFileURL(pkgDetailsStruct.PkgDetailsUrl)
    if parsedNameAndVersionArr != nil {
        pkgDetailsStruct.Name = parsedNameAndVersionArr[0]
        pkgDetailsStruct.Version = parsedNameAndVersionArr[1]
    }
    return pkgDetailsStruct
}


func ParseHttpRequestResponseForPackagesVersions(responseBody string) [] NugetPackageDetailsStruct {
    parsedPackagesVersionsArr := make([] NugetPackageDetailsStruct, 0)
    LogInfo.Printf("Parsing http request response for packages details")
    parsedPackagesDetailsStruct := nuget_packages_xml.ParseMultipleNugetPackagesXmlData(responseBody)
    if len(parsedPackagesDetailsStruct.Entry) == 0 {  // If failed to parse entries, it might be only a single entry and in that case attempt to parse it
        entryStruct := nuget_packages_xml.ParseSingleNugetPackagesXmlData(responseBody)
        pkgDetailsStruct := ParseXmlDataToSinglePkgDetailsStruct(entryStruct)
        parsedPackagesVersionsArr = append(parsedPackagesVersionsArr, pkgDetailsStruct)
        return parsedPackagesVersionsArr  
    }
    for _, entryStruct := range parsedPackagesDetailsStruct.Entry {
        pkgDetailsStruct := ParseXmlDataToSinglePkgDetailsStruct(entryStruct)
        if len(pkgDetailsStruct.Name) == 0 || len(pkgDetailsStruct.Version) == 0 {continue}
        parsedPackagesVersionsArr = append(parsedPackagesVersionsArr, pkgDetailsStruct)
    }
    return parsedPackagesVersionsArr
}

func SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs HttpRequestArgsStruct) [] NugetPackageDetailsStruct {
    responseBody := MakeHttpRequest(httpRequestArgs)
    if len(responseBody) == 0 {return [] NugetPackageDetailsStruct {}}
    parsedPackagesDetailsArr := ParseHttpRequestResponseForPackagesVersions(responseBody)

    return parsedPackagesDetailsArr
}

func DownloadPkg(downloadPkgDetailsStruct DownloadPackageDetailsStruct) {
    LogInfo.Printf("Downloading package: %s==%s", downloadPkgDetailsStruct.PkgDetailsStruct.Name, downloadPkgDetailsStruct.PkgDetailsStruct.Version)
    fileUrl := downloadPkgDetailsStruct.PkgDetailsStruct.PkgFileUrl
    downloadFilePath := downloadPkgDetailsStruct.DownloadFilePath
    downloadFileChecksum := downloadPkgDetailsStruct.DownloadFileChecksum
    fileChecksum := downloadPkgDetailsStruct.PkgDetailsStruct.Checksum
    if fileChecksum == downloadFileChecksum {
        fileName := filepath.Base(downloadFilePath)
        LogWarning.Printf("Checksum match: download target file already exists. Skipping download of: \"%s\"", fileName)
        return
    }
    MakeHttpRequest(
        HttpRequestArgsStruct{
            UrlAddress: fileUrl,
            Method: "GET",
            DownloadFilePath: downloadFilePath,
        },
    )
}

func UploadPkg(uploadPkgStruct UploadPackageDetailsStruct, httpRequestArgsStruct HttpRequestArgsStruct) {
    pkgPrintStr := fmt.Sprintf("%s==%s", uploadPkgStruct.PkgDetailsStruct.Name, uploadPkgStruct.PkgDetailsStruct.Version)
	LogInfo.Printf("Uploading package: \"%s\" from: %s", pkgPrintStr, uploadPkgStruct.UploadFilePath)
    httpRequestArgsStruct.Method = "PUT"
    httpRequestArgsStruct.UploadFilePath = uploadPkgStruct.UploadFilePath
    MakeHttpRequest(httpRequestArgsStruct)

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

func Synched_AppendDownloadedPkgDetailsObj(arr_1 *[] DownloadPackageDetailsStruct, downloadPkgDetailsStruct DownloadPackageDetailsStruct) {
    appendDownloadedPkgDetailsArr_Lock.Lock()
    *arr_1 = append(*arr_1, downloadPkgDetailsStruct)
    appendDownloadedPkgDetailsArr_Lock.Unlock()
}
