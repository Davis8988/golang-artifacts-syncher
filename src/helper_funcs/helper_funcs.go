package helper_funcs

import (
	"golang-artifacts-syncher/src/nuget_packages_xml"
	"golang-artifacts-syncher/src/global_structs"
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
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

    SrcServersUserToUse          string
    SrcServersPassToUse          string
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
    HttpRequestTimeoutSecondsInt int

    srcServersUrlsArr     []string
    srcReposNamesArr      []string
    DestServersUrlsArr    []string
    destReposNamesArr     []string
    packagesNamesArr      []string
    packagesVersionsArr   []string
    HttpRequestHeadersMap map[string]string
    packagesToDownloadMap sync.Map
)

func Init() {
    LogInfo.Print("Initializing helpers pkg vars")
    convertSyncedMapToString_Lock = sync.RWMutex{}
    appendPkgDetailsArr_Lock = sync.RWMutex{}
    appendDownloadedPkgDetailsArr_Lock = sync.RWMutex{}

    LogInfo.Print("Initializing from envs vars")
    SrcServersUserToUse = Getenv("SRC_SERVERS_USER_TO_USE", "")
	SrcServersPassToUse = Getenv("SRC_SERVERS_PASS_TO_USE", "")
	srcServersUrlsStr = Getenv("SRC_SERVERS_URLS_STR", "")
	srcReposNamesStr = Getenv("SRC_REPOS_NAMES_STR", "")
	destServersUrlsStr = Getenv("DEST_SERVERS_URLS_STR", "")
	destReposNamesStr = Getenv("DEST_REPOS_NAMES_STR", "")
	destServersUserToUse = Getenv("DEST_SERVERS_USER_TO_USE", "")
	destServersPassToUse = Getenv("DEST_SERVERS_PASS_TO_USE", "")
	packagesNamesStr = Getenv("PACKAGES_NAMES_STR", "")
	packagesVersionsStr = Getenv("PACKAGES_VERSIONS_STR", "")
	httpRequestHeadersStr = Getenv("HTTP_REQUEST_HEADERS_STR", "") // Example: "key=value;key1=value1;key2=value2"
	DownloadPkgsDirPath = Getenv("DOWNLOAD_PKGS_DIR_PATH", GetCurrentProgramDir())
	HttpRequestTimeoutSecondsInt = StrToInt(Getenv("HTTP_REQUEST_TIMEOUT_SECONDS_INT", "45"))
}

func PrintVars() {
    LogInfo.Printf("SRC_SERVERS_URLS_STR: '%s'", srcServersUrlsStr)
	LogInfo.Printf("SRC_REPOS_NAMES_STR: '%s'", srcReposNamesStr)
	LogInfo.Printf("SRC_SERVERS_USER_TO_USE: '%s'", SrcServersUserToUse)
	LogInfo.Printf("SRC_SERVERS_PASS_TO_USE: '%s'", strings.Repeat("*", len(SrcServersPassToUse)))
	LogInfo.Printf("DEST_SERVERS_URLS_STR: '%s'", destServersUrlsStr)
	LogInfo.Printf("DEST_REPOS_NAMES_STR: '%s'", destReposNamesStr)
	LogInfo.Printf("DEST_SERVERS_USER_TO_USE: '%s'", destServersUserToUse)
	LogInfo.Printf("DEST_SERVERS_PASS_TO_USE: '%s'", strings.Repeat("*", len(destServersPassToUse)))
	LogInfo.Printf("PACKAGES_NAMES_STR: '%s'", packagesNamesStr)
	LogInfo.Printf("PACKAGES_VERSIONS_STR: '%s'", packagesVersionsStr)
	LogInfo.Printf("HTTP_REQUEST_HEADERS_STR: '%s'", httpRequestHeadersStr)
	LogInfo.Printf("DOWNLOAD_PKGS_DIR_PATH: '%s'", DownloadPkgsDirPath)
	LogInfo.Printf("HTTP_REQUEST_TIMEOUT_SECONDS_INT: '%d'", HttpRequestTimeoutSecondsInt)

	LogInfo.Printf("srcServersUrlsArr: %v", srcServersUrlsArr)
	LogInfo.Printf("DestServersUrlsArr: %v", DestServersUrlsArr)
	LogInfo.Printf("srcReposNamesArr: %v", srcReposNamesArr)
	LogInfo.Printf("packagesNamesArr: %v", packagesNamesArr)
	LogInfo.Printf("packagesVersionsArr: %v", packagesVersionsArr)
	packagesToDownloadMapStr := Synched_ConvertSyncedMapToString(packagesToDownloadMap)
	LogInfo.Printf("packagesToDownloadMap: \n%v", packagesToDownloadMapStr)
}

func ValidateEnvironment() {
    LogInfo.Print("Validating envs")

	// Validate len(packagesVersionsArr) == len(packagesNamesArr)  (Only when packagesVersionsArr is defined)
	if ! IsStrArrayEmpty(packagesVersionsArr) {
		LogInfo.Print("Comparing packages names & versions arrays lengths")
		if len(packagesVersionsArr) != len(packagesNamesArr) {
			errMsg := "Packages Versions to search count is different from Packages Names to search count\n"
			errMsg += "Can't search for packages versions & names which are not of the same count.\n"
			errMsg += "When passing packages versions to search - the versions count must be of the same count of packages names to search.\n"
			errMsg += "A version for each package name to search"
			LogError.Fatal(errMsg)
		}
	}

	LogInfo.Print("All Good")
}

func UpdateVars() {
    LogInfo.Print("Updating vars")
	srcServersUrlsArr = make([]string, 0, 4)
	DestServersUrlsArr = make([]string, 0, 4)
	srcReposNamesArr = make([]string, 0, 4)
	packagesNamesArr = make([]string, 0, 10)
	packagesVersionsArr = make([]string, 0, 10)
	if len(srcServersUrlsStr) > 1 {srcServersUrlsArr = strings.Split(srcServersUrlsStr, ";")}
	if len(srcReposNamesStr) > 1 {srcReposNamesArr = strings.Split(srcReposNamesStr, ";")}
	if len(destServersUrlsStr) > 1 {DestServersUrlsArr = strings.Split(destServersUrlsStr, ";")}
	if len(destReposNamesStr) > 1 {destReposNamesArr = strings.Split(destReposNamesStr, ";")}
	if len(packagesNamesStr) > 1 {packagesNamesArr = strings.Split(packagesNamesStr, ";")}
	if len(packagesVersionsStr) > 1 {packagesVersionsArr = strings.Split(packagesVersionsStr, ";")}
	HttpRequestHeadersMap = ParseHttpHeadersStrToMap(httpRequestHeadersStr)

	for i, pkgName := range packagesNamesArr {
		// If map doesn't contain value at: 'pkgName' - add one to point to empty string array: []
		packagesToDownloadMap.LoadOrStore(pkgName, make([]string, 0, 10))
		// If received a version array for it - add it to the list
		if len(packagesVersionsArr) > i {
			pkgVersion := packagesVersionsArr[i]
			currentVersionsArr := LoadStringArrValueFromSynchedMap(packagesToDownloadMap, pkgName)
			packagesToDownloadMap.Store(pkgName, append(currentVersionsArr, pkgVersion))
		}
	}
}

func PrepareSrcSearchAllPkgsVersionsUrlsArray() []string {
	var searchUrlsArr = make([]string, 0, 10) // Create a slice with length=0 and capacity=10

	LogInfo.Print("Preparing src search packages urls array")
	for _, srcServerUrl := range srcServersUrlsArr {
		for _, repoName := range srcReposNamesArr {
			for _, pkgName := range packagesNamesArr {
				versionsToSearchArr := LoadStringArrValueFromSynchedMap(packagesToDownloadMap, pkgName)
				if len(versionsToSearchArr) == 0 { // Either use search
					searchUrlsArr = append(searchUrlsArr, srcServerUrl+"/"+repoName+"/"+"Packages()?$filter=tolower(Id)%20eq%20'"+pkgName+"'")
					continue
				} // Or specific package details request for each specified requested version
				for _, pkgVersion := range versionsToSearchArr {
					searchUrlsArr = append(searchUrlsArr, srcServerUrl+"/"+repoName+"/"+"Packages(Id='"+pkgName+"',Version='"+pkgVersion+"')")
				}

			}
		}
	}
	return searchUrlsArr
}

func FilterFoundPackagesByRequestedVersion(foundPackagesDetailsArr []global_structs.NugetPackageDetailsStruct) []global_structs.NugetPackageDetailsStruct {
	LogInfo.Printf("Filtering found pkgs by requested versions")
	var filteredPackagesDetailsArr []global_structs.NugetPackageDetailsStruct
	for _, pkgDetailStruct := range foundPackagesDetailsArr {
		pkgVersion := pkgDetailStruct.Version
		pkgName := pkgDetailStruct.Name
		versionsToSearchArr := LoadStringArrValueFromSynchedMap(packagesToDownloadMap, pkgName) // Use global var: packagesToDownloadMap
		if len(versionsToSearchArr) == 0 {
			filteredPackagesDetailsArr = append(filteredPackagesDetailsArr, pkgDetailStruct)
			continue
		}
		for _, requestedVersion := range versionsToSearchArr {
			if pkgVersion == requestedVersion {filteredPackagesDetailsArr = append(filteredPackagesDetailsArr, pkgDetailStruct)} // This version is requested - Add pkg details obj to the result filtered array
		}
	}
	return filteredPackagesDetailsArr
}


func UploadDownloadedPackage(uploadPkgStruct global_structs.UploadPackageDetailsStruct) global_structs.UploadPackageDetailsStruct {
	pkgPrintStr := fmt.Sprintf("%s==%s", uploadPkgStruct.PkgDetailsStruct.Name, uploadPkgStruct.PkgDetailsStruct.Version)
	pkgName := uploadPkgStruct.PkgDetailsStruct.Name
	pkgVersion := uploadPkgStruct.PkgDetailsStruct.Version

	// Check if package already exists. If so, then compare it's checksum and skip on matching
	for _, destServerUrl := range DestServersUrlsArr {
		for _, repoName := range destReposNamesArr {
			destServerRepo := destServerUrl + "/" + repoName
			LogInfo.Printf("Checking if pkg: '%s' already exists at dest server: %s", pkgPrintStr, destServerRepo)
			checkDestServerPkgExistUrl := destServerRepo + "/" + "Packages(Id='" + pkgName + "',Version='" + pkgVersion + "')"
			httpRequestArgs := global_structs.HttpRequestArgsStruct{
				UrlAddress: checkDestServerPkgExistUrl,
				HeadersMap: HttpRequestHeadersMap,
				UserToUse:  destServersUserToUse,
				PassToUse:  destServersPassToUse,
				TimeoutSec: HttpRequestTimeoutSecondsInt,
				Method:     "GET",
			}

			foundPackagesDetailsArr := SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs)
			LogInfo.Printf("Found: %s", foundPackagesDetailsArr)

			emptyNugetPackageDetailsStruct := global_structs.NugetPackageDetailsStruct{}
			shouldCompareChecksum := true
			if len(foundPackagesDetailsArr) != 1 {
				LogInfo.Printf("Found multiple or no packages: \"%d\" - Should be only 1. Skipping checksum comparison. Continuing with the upload..", len(foundPackagesDetailsArr))
				shouldCompareChecksum = false
			} else if len(foundPackagesDetailsArr) == 1 && foundPackagesDetailsArr[0] == emptyNugetPackageDetailsStruct {
				LogInfo.Print("No package found. Continuing with the upload..")
				shouldCompareChecksum = false
			}
			
			if shouldCompareChecksum {
				// Check the checksum:
				LogInfo.Printf("Comparing found package's checksum to know if should upload to: %s or not", destServerRepo)
				foundPackageChecksum := foundPackagesDetailsArr[0].Checksum
				fileToUploadChecksum := uploadPkgStruct.UploadFileChecksum
				if foundPackageChecksum == fileToUploadChecksum {
				fileName := filepath.Base(uploadPkgStruct.UploadFilePath)
				LogWarning.Printf("Checksum match: upload target file already exists in dest server: '%s' \n"+
					"Skipping upload of pkg: \"%s\"", destServerRepo, fileName)
				return uploadPkgStruct
				}
			}
			
			if len(destServerRepo) > 1 {
				lastChar := destServerRepo[len(destServerRepo)-1:]
				LogInfo.Printf("Adding '/' char to dest server repo url: \"%s\"", destServerRepo)
				if lastChar != "/" {destServerRepo += "/"}
			}
			httpRequestArgs.UrlAddress = destServerRepo
			// Upload the package file
			UploadPkg(uploadPkgStruct, httpRequestArgs)
		}
	}

	return uploadPkgStruct
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
    
    HttpRequestHeadersMap := make(map[string] string)
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
        HttpRequestHeadersMap[headerKey] = headerValue
    }
    return HttpRequestHeadersMap
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

func MakeHttpRequest(httpRequestArgs global_structs.HttpRequestArgsStruct) string {
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
    var writer *multipart.Writer

    // Upload file (PUT requests):
    if method == "PUT" && len(uploadFilePath) > 0 {
        body, writer = ReadFileContentsIntoPartsForUpload(uploadFilePath, "package")
    }

    req, err := http.NewRequest(method, urlAddress, body)
    if err != nil {
        LogError.Printf("%s\nFailed creating HTTP request object for URL: \"%s\"", err, urlAddress)
        return ""
    }

    // Incase pushing a file, then add the Content Type header from the reader (includes boundary)
    if method == "PUT" && len(uploadFilePath) > 0 {
        LogInfo.Printf("Adding header:  'Content-Type'")
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

func ParseXmlDataToSinglePkgDetailsStruct(entryStruct nuget_packages_xml.SinglePackagesDetailsXmlStruct) global_structs.NugetPackageDetailsStruct {
    var pkgDetailsStruct global_structs.NugetPackageDetailsStruct
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


func ParseHttpRequestResponseForPackagesVersions(responseBody string) [] global_structs.NugetPackageDetailsStruct {
    parsedPackagesVersionsArr := make([] global_structs.NugetPackageDetailsStruct, 0)
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

func SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs global_structs.HttpRequestArgsStruct) [] global_structs.NugetPackageDetailsStruct {
    responseBody := MakeHttpRequest(httpRequestArgs)
    if len(responseBody) == 0 {return [] global_structs.NugetPackageDetailsStruct {}}
    parsedPackagesDetailsArr := ParseHttpRequestResponseForPackagesVersions(responseBody)

    return parsedPackagesDetailsArr
}

func DownloadPkg(downloadPkgDetailsStruct global_structs.DownloadPackageDetailsStruct) {
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
        global_structs.HttpRequestArgsStruct{
            UrlAddress: fileUrl,
            Method: "GET",
            DownloadFilePath: downloadFilePath,
        },
    )
}

func UploadPkg(uploadPkgStruct global_structs.UploadPackageDetailsStruct, httpRequestArgsStruct global_structs.HttpRequestArgsStruct) {
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

func Synched_AppendPkgDetailsObj(arr_1 *[] global_structs.NugetPackageDetailsStruct, arr_2 [] global_structs.NugetPackageDetailsStruct) {
    appendPkgDetailsArr_Lock.Lock()
    *arr_1 = append(*arr_1, arr_2...)
    appendPkgDetailsArr_Lock.Unlock()
}

func Synched_AppendDownloadedPkgDetailsObj(arr_1 *[] global_structs.DownloadPackageDetailsStruct, downloadPkgDetailsStruct global_structs.DownloadPackageDetailsStruct) {
    appendDownloadedPkgDetailsArr_Lock.Lock()
    *arr_1 = append(*arr_1, downloadPkgDetailsStruct)
    appendDownloadedPkgDetailsArr_Lock.Unlock()
}
