package helper_funcs

import (
	"golang-artifacts-syncher/src/global_structs"
	"golang-artifacts-syncher/src/global_vars"
	"golang-artifacts-syncher/src/mylog"
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"sort"
)

var (
    startTime time.Time
)

func InitSuccessIndicatorFilePath() {
    global_vars.SuccessIndicatorFile = Filepath_Join(GetCurrentProgramDir(), "syncher_execution_finished_successfully.txt")
}

func InitVars() {
    global_vars.ErrorsDetected = false

    mylog.Logger.Info("Initializing helpers pkg vars")
    global_vars.ConvertSyncedMapToString_Lock      = sync.RWMutex{}
    global_vars.JoinTwoPkgDetailsSlices_Lock       = sync.RWMutex{}
    global_vars.JoinTwoPkgDetailsMaps_Lock         = sync.RWMutex{}
    global_vars.AppendDownloadedPkgDetailsArr_Lock = sync.RWMutex{}
    global_vars.AppendUploadedPkgDetailsArr_Lock   = sync.RWMutex{}
    global_vars.AppendPkgDetailsArr_Lock           = sync.RWMutex{}
    global_vars.AppendPkgDetailsMap_Lock           = sync.RWMutex{}
    global_vars.ErrorsDetected_Lock                = sync.RWMutex{}

    httpRequestGlobalDefaultTimeoutSecondsInt := 210
    mylog.Logger.Info("Initializing from envs vars")
    global_vars.AppConfig = global_structs.AppConfiguration{
        SrcServersUserToUse                : Getenv("SRC_SERVERS_USER_TO_USE", ""),
        SrcServersPassToUse                : Getenv("SRC_SERVERS_PASS_TO_USE", ""),
        SrcServersUrlsStr                  : Getenv("SRC_SERVERS_URLS_STR", ""),
        DestServersUrlsStr                 : Getenv("DEST_SERVERS_URLS_STR", ""),
        DestServersUserToUse               : Getenv("DEST_SERVERS_USER_TO_USE", ""),
        DestServersPassToUse               : Getenv("DEST_SERVERS_PASS_TO_USE", ""),
        PackagesNamesStr                   : Getenv("PACKAGES_NAMES_STR", ""),
        PackagesVersionsStr                : Getenv("PACKAGES_VERSIONS_STR", ""),
        HttpRequestHeadersStr              : Getenv("HTTP_REQUEST_HEADERS_STR", ""), // Example: "key=value;key1=value1;key2=value2"
        DownloadPkgsDirPath                : Getenv("DOWNLOAD_PKGS_DIR_PATH", Filepath_Join(GetCurrentProgramDir(), "Downloads")),
        ChecksumFilesDirPath               : Getenv("CHECKSUM_FILES_DIR_PATH", Filepath_Join(GetCurrentProgramDir(), "Checksums")),
        HttpRequestGlobalDefaultTimeoutSecondsInt : StrToInt(Getenv("HTTP_REQUEST_GLOBAL_DEFAULT_TIMEOUT_SECONDS_INT", strconv.Itoa(httpRequestGlobalDefaultTimeoutSecondsInt))),
        HttpRequestDownloadTimeoutSecondsInt      : StrToInt(Getenv("HTTP_REQUEST_DOWNLOAD_TIMEOUT_SECONDS_INT", strconv.Itoa(httpRequestGlobalDefaultTimeoutSecondsInt))),
        HttpRequestUploadTimeoutSecondsInt        : StrToInt(Getenv("HTTP_REQUEST_UPLOAD_TIMEOUT_SECONDS_INT", strconv.Itoa(httpRequestGlobalDefaultTimeoutSecondsInt))),
        SearchPackagesUrlSkipGroupCount           : StrToInt(Getenv("SEARCH_PACKAGES_URL_SKIP_GROUP_COUNT", "30")),
        PackagesMaxConcurrentDownloadCount        : StrToInt(Getenv("PACKAGES_MAX_CONCURRENT_DOWNLOAD_COUNT", "5")),
        PackagesMaxConcurrentUploadCount          : StrToInt(Getenv("PACKAGES_MAX_CONCURRENT_UPLOAD_COUNT", "5")),
        PackagesMaxConcurrentDeleteCount          : StrToInt(Getenv("PACKAGES_MAX_CONCURRENT_DELETE_COUNT", "5")),
        PackagesDownloadLimitCount                : StrToInt(Getenv("PACKAGES_DOWNLOAD_LIMIT_COUNT", "-1")),
    }
    
    // Log level - this is set in mylog module
    global_vars.AppConfig.LogLevel = mylog.Logger.Level.String()
    
}

func Fmt_Println(a ...interface{}) (n int, err error) {
    return fmt.Println(a ...)
}

func Fmt_Print(a ...interface{}) (n int, err error) {
    return fmt.Print(a ...)
}

func PrintVars() {
    appConfigStr := global_vars.AppConfig.ToString()
    mylog.Logger.Infof("Configuration: \n%s", appConfigStr)
	
	packagesToDownloadMapStr := Synched_ConvertSyncedMapToString(global_vars.PackagesToDownloadMap)
	mylog.Logger.Infof("packagesToDownloadMap: \n%v", packagesToDownloadMapStr)
}

func CreateRequiredFiles() {
    CreateDir(global_vars.AppConfig.ChecksumFilesDirPath)
    CreateDir(global_vars.AppConfig.DownloadPkgsDirPath)
}

func ValidateEnvironment() {
    mylog.Logger.Info("Validating envs")

	// Validate len(packagesVersionsArr) == len(packagesNamesArr)  (Only when packagesVersionsArr is defined)
	if ! IsStrArrayEmpty(global_vars.AppConfig.PackagesVersionsArr) {
		mylog.Logger.Debug("Comparing packages names & versions arrays lengths")
		if len(global_vars.AppConfig.PackagesVersionsArr) != len(global_vars.AppConfig.PackagesNamesArr) {
			errMsg := "Packages Versions to search count is different from Packages Names to search count\n"
			errMsg += "Can't search for packages versions & names which are not of the same count.\n"
			errMsg += "When passing packages versions to search - the versions count must be of the same count of packages names to search.\n"
			errMsg += "A version for each package name to search"
			mylog.Logger.Fatal(errMsg)
		}
	}

    mylog.Logger.Debug("Validating all src URLs addresses: %s", global_vars.AppConfig.SrcServersUrlsArr)
    for i, srcServerUrl := range global_vars.AppConfig.SrcServersUrlsArr {
        if len(srcServerUrl) == 1 {continue}
        lastChar := srcServerUrl[len(srcServerUrl)-1:]
        if lastChar == "/" {continue}
        mylog.Logger.Debugf("Fix: Adding '/' char to src server repo url: \"%s\"", srcServerUrl)
        srcServerUrl += "/"
        global_vars.AppConfig.SrcServersUrlsArr[i] = srcServerUrl
    }
    
    mylog.Logger.Debug("Validating all dest URLs addresses: %s", global_vars.AppConfig.DestServersUrlsArr)
    for i, destServerUrl := range global_vars.AppConfig.DestServersUrlsArr {
        if len(destServerUrl) == 1 {continue}
        lastChar := destServerUrl[len(destServerUrl)-1:]
        if lastChar == "/" {continue}
        mylog.Logger.Debugf("Fix: Adding '/' char to dest server repo url: \"%s\"", destServerUrl)
        destServerUrl += "/"
        global_vars.AppConfig.DestServersUrlsArr[i] = destServerUrl
    }

	mylog.Logger.Info("All Good")
}

func UpdateVars() {
    mylog.Logger.Info("Updating vars")
	global_vars.AppConfig.SrcServersUrlsArr = make([]string, 0, 4)
	global_vars.AppConfig.DestServersUrlsArr = make([]string, 0, 4)
	global_vars.AppConfig.PackagesNamesArr = make([]string, 0, 10)
	global_vars.AppConfig.PackagesVersionsArr = make([]string, 0, 10)
	if len(global_vars.AppConfig.SrcServersUrlsStr)   > 1 {global_vars.AppConfig.SrcServersUrlsArr   = strings.Split(global_vars.AppConfig.SrcServersUrlsStr,   ";")}
	if len(global_vars.AppConfig.DestServersUrlsStr)  > 1 {global_vars.AppConfig.DestServersUrlsArr  = strings.Split(global_vars.AppConfig.DestServersUrlsStr,  ";")}
	if len(global_vars.AppConfig.PackagesNamesStr)    > 1 {global_vars.AppConfig.PackagesNamesArr    = strings.Split(global_vars.AppConfig.PackagesNamesStr,    ";")}
	if len(global_vars.AppConfig.PackagesVersionsStr) > 1 {global_vars.AppConfig.PackagesVersionsArr = strings.Split(global_vars.AppConfig.PackagesVersionsStr, ";")}
	global_vars.AppConfig.HttpRequestHeadersMap = ParseHttpHeadersStrToMap(global_vars.AppConfig.HttpRequestHeadersStr)

	for i, pkgName := range global_vars.AppConfig.PackagesNamesArr {
		// If map doesn't contain value at: 'pkgName' - add one to point to empty string array: []
		global_vars.PackagesToDownloadMap.LoadOrStore(pkgName, make([]string, 0, 10))
		// If received a version array for it - add it to the list
		if len(global_vars.AppConfig.PackagesVersionsArr) > i {
			pkgVersion := global_vars.AppConfig.PackagesVersionsArr[i]
			currentVersionsArr := LoadStringArrValueFromSynchedMap(global_vars.PackagesToDownloadMap, pkgName)
			global_vars.PackagesToDownloadMap.Store(pkgName, append(currentVersionsArr, pkgVersion))
		}
	}
}

func PrepareSrcSearchUrlsForPackageArray(pkgName string, pkgVersion string) []string {
	var searchUrlsArr = make([]string, 0, 10) // Create a slice with length=0 and capacity=10

	mylog.Logger.Info("Preparing src search packages urls array")
	for _, srcServerUrl := range global_vars.AppConfig.SrcServersUrlsArr {
        if len(srcServerUrl) == 1 {continue}
        versionsToSearchArr := LoadStringArrValueFromSynchedMap(global_vars.PackagesToDownloadMap, pkgName)
        
        if len(pkgVersion) == 0 { 
            // Either use search
            searchUrlsArr = append(searchUrlsArr, srcServerUrl + "Packages()?$filter=tolower(Id)%20eq%20'"+pkgName+"'")
            continue
        } else {
            // Or specific package details request for each specified requested version
            searchUrlsArr = append(searchUrlsArr, srcServerUrl + "Packages(Id='"+pkgName+"',Version='"+pkgVersion+"')")
        }

	}
	return searchUrlsArr
}

func FilterFoundPackagesByRequestedVersion(foundPackagesDetailsArr []global_structs.NugetPackageDetailsStruct) []global_structs.NugetPackageDetailsStruct {
	var filteredPackagesDetailsArr []global_structs.NugetPackageDetailsStruct
	for _, pkgDetailStruct := range foundPackagesDetailsArr {
		pkgVersion := pkgDetailStruct.Version
		pkgName := pkgDetailStruct.Name
		versionsToSearchArr := LoadStringArrValueFromSynchedMap(global_vars.PackagesToDownloadMap, pkgName) // Use global var: packagesToDownloadMap
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


func Fmt_Sprintf(format string, a ...interface{}) string {
    return fmt.Sprintf(format, a...);
}

func Filepath_GetFileNameFromPath(somePath string) string {
    return filepath.Base(somePath);
}

func Filepath_Join(elem ...string) string {
    return filepath.Join(elem...);
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
        mylog.Logger.Fatalf("%s\nFailed getting current program's dir", err)
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
        mylog.Logger.Panicf("\n%s\nFailed converting string: \"%s\" to integer\n", err, strVar)
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
		mylog.Logger.Infoln(someVal);
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
    mylog.Logger.Infof("Looping on headers values to init headers map")
    for _, headersPairStr := range tempHeadersPairsArr {
        tempPairArr := strings.Split(headersPairStr, "=")
        if len(tempPairArr) != 2 {
            mylog.Logger.Panicf("\nFound header pair: \"%v\"  that is not in the right format of: \"key=value\"\n", tempPairArr)
            return nil
        }
        headerKey := tempPairArr[0]
        headerValue := tempPairArr[1]
        HttpRequestHeadersMap[headerKey] = headerValue
    }
    return HttpRequestHeadersMap
}

func GetFileName(filePath string) string {
    return filepath.Base(filePath)
}

func PathExists(pathToCheck string) bool {
    if _, err := os.Stat(pathToCheck); err == nil {return true}  
    return false
}

func Exit(exitCode int) {
    os.Exit(exitCode)
}

func CreateDir(dirPath string) {
    if PathExists(dirPath) {return}  // If dir already exists - finish here
    mylog.Logger.Debugf("Creating dir: %s", dirPath)
    err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		mylog.Logger.Panicf("\n%s\nFailed creating dir: \"%s\"\n", err, dirPath)
	}
}

func CreateFile(filePath string) *os.File {
    dirPath := filepath.Dir(filePath)
    CreateDir(dirPath)
    mylog.Logger.Debugf("Creating file: %s", filePath)
    // Create the file
    file, err := os.Create(filePath)
    if err != nil  {
        mylog.Logger.Panicf("\n%s\nFailed creating file: \"%s\"\n", err, filePath)
    }
    return file
}

func CalculateFileChecksum(filePath string) string {
    if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {return ""}  // If file doesn't exist: return empty
    mylog.Logger.Debugf("Calculating sha512 checksum of file: %s", filePath)
    f, err := os.Open(filePath)
    if err != nil {
        mylog.Logger.Panicf("\n%s\nFailed calculating sha512 checksum of file: \"%s\"\n", err, filePath)
    }
    defer f.Close()
    h := sha512.New()
    if _, err := io.Copy(h, f); err != nil {
        mylog.Logger.Panicf("\n%s\nFailed calculating sha512 checksum of file: \"%s\"\n", err, filePath)
    }

    return base64.StdEncoding.EncodeToString(h.Sum(nil))  // nil instead of a byte array to append to
}

func MakeHttpRequest(httpRequestArgs global_structs.HttpRequestArgsStruct) *global_structs.HttpResponseStruct {
    urlAddress := httpRequestArgs.UrlAddress
    downloadFilePath := httpRequestArgs.DownloadFilePath
    uploadFilePath := httpRequestArgs.UploadFilePath
    headersMap := httpRequestArgs.HeadersMap
    username := httpRequestArgs.UserToUse
    password := httpRequestArgs.PassToUse
    timeoutSec := httpRequestArgs.TimeoutSec
    method := strings.ToUpper(httpRequestArgs.Method)

    mylog.Logger.Infof("Making an HTTP '%s' request to URL: \"%s\"", method, urlAddress)

    client := http.Client{Timeout: time.Duration(timeoutSec) * time.Second,}
    
    var body io.Reader
    var writer *multipart.Writer

    // Upload file (PUT requests):
    if method == "PUT" && len(uploadFilePath) > 0 {
        body, writer = ReadFileContentsIntoPartsForUpload(uploadFilePath, "package")
    }

    req, err := http.NewRequest(method, urlAddress, body)
    if err != nil {
        mylog.Logger.Panicf("\n%s\nFailed creating HTTP request object for URL: \"%s\"\n", err, urlAddress)
        return nil
    }

    // Incase pushing a file, then add the Content Type header from the reader (includes boundary)
    if method == "PUT" && len(uploadFilePath) > 0 {
        mylog.Logger.Debugf("Adding header:  'Content-Type'")
        req.Header.Add("Content-Type", writer.FormDataContentType())
    }

    // Adding headers:
    for k := range headersMap {
        mylog.Logger.Debugf("Adding header:  '%s'=\"%s\"", k, headersMap[k])
        req.Header.Add(k, headersMap[k])
    }

    // Adding creds
    if len(username) > 0 && len(password) > 0 {
        mylog.Logger.Debugf("Adding creds of user:  '%s'", username)
        req.SetBasicAuth(username, password)
    }

    // Make the http request
    response, err := client.Do(req)
    if err != nil {
        mylog.Logger.Panicf("\n%s\n", err)
        return nil
    }
  
    defer response.Body.Close() // Finally step: close the body obj
    
    httpResponseResultStruct := global_structs.HttpResponseStruct {
        UrlAddress: urlAddress,
        BodyStr: "",
        StatusStr: response.Status,
        StatusCode: response.StatusCode,
    }

    // If got: downloadFilePath var, then Writer the body to file
    if len(downloadFilePath) > 0 {
        mylog.Logger.Infof("Downloading '%s' to:  %s", urlAddress, downloadFilePath)
        fileHandle := CreateFile(downloadFilePath)  // Create the file
        defer fileHandle.Close()

        _, err = io.Copy(fileHandle, response.Body)
        if err != nil  {
            mylog.Logger.Panicf("\n%s\nFailed writing response Body to file: %s\n", err, downloadFilePath)
        }
        return &httpResponseResultStruct // Finish here
    }

    responseBody, err := ioutil.ReadAll(response.Body)
    if err != nil {
        mylog.Logger.Panicf("\n%s\nFailed reading request's response body: %s\n", err, urlAddress)
        return &httpResponseResultStruct
    }

    httpResponseResultStruct.BodyStr = string(responseBody)

    return &httpResponseResultStruct
}

func ReadFileContentsIntoPartsForUpload(uploadFilePath string, headerFieldName string) (io.Reader, *multipart.Writer) {
    mylog.Logger.Infof("Reading file content for upload: \"%s\"", uploadFilePath)

    // If missing file: return empty body
    if _, err := os.Stat(uploadFilePath); errors.Is(err, os.ErrNotExist) {
        mylog.Logger.Panicf("\n%s\nFailed uploading file: \"%s\" since it is missing. Failed preparing HTTP request object\n", err, uploadFilePath)
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

func ReadFileContent(filePath string) string {
    mylog.Logger.Infof("Reading file: %s", filePath)
    content, err := os.ReadFile(filePath)
    if err != nil {
        mylog.Logger.Panic(err)
    }
    return strings.TrimSpace(string(content))
}

func WriteFileContent(filePath string, content string) {
    mylog.Logger.Infof("Writing '%s' to file: %s", content, filePath)
    if err := os.WriteFile(filePath, []byte(content), 0666); err != nil {
        mylog.Logger.Panic(err)
    }
}


func Synched_ConvertSyncedMapToString(synchedMap sync.Map) string {
	global_vars.ConvertSyncedMapToString_Lock.Lock()
	result := ConvertSyncedMapToString(synchedMap)
	defer global_vars.ConvertSyncedMapToString_Lock.Unlock()
	return result
}

func ConvertPkgDetailsArrayToMap(pkgDetailsArr [] global_structs.NugetPackageDetailsStruct) map[string] global_structs.NugetPackageDetailsStruct {
    convertedPkgDetailsMap := make(map[string] global_structs.NugetPackageDetailsStruct)
    for _, pkgDetailsStruct := range pkgDetailsArr {
        convertedPkgDetailsMap[pkgDetailsStruct.HashCode()] = pkgDetailsStruct
    }
    return convertedPkgDetailsMap
}

func ConvertPkgDetailsMapToArray(pkgDetailsMap map[string] global_structs.NugetPackageDetailsStruct) [] global_structs.NugetPackageDetailsStruct {
    mapKeysCount := len(pkgDetailsMap)
    if (mapKeysCount == 0) {return [] global_structs.NugetPackageDetailsStruct {} }  // Empty array
    resultPkgDetailsArray := make([]global_structs.NugetPackageDetailsStruct, 0, mapKeysCount)

    for _, pkgDetailsStruct := range pkgDetailsMap {
        resultPkgDetailsArray = append(resultPkgDetailsArray, pkgDetailsStruct)
    }
    return resultPkgDetailsArray
}

func AppendPkgDetailsArrayToMap(pkgDetailsMap map[string] global_structs.NugetPackageDetailsStruct, pkgDetailsArr [] global_structs.NugetPackageDetailsStruct) {
    for _, pkgDetailsStruct := range pkgDetailsArr {
        indStr := pkgDetailsStruct.HashCode()
        pkgDetailsMap[indStr] = pkgDetailsStruct
    }
}

func Synched_AddPkgDetailsStructMapToMap(map_1 map[string] map[string] map[string] global_structs.NugetPackageDetailsStruct, key string, pkgDetailsStructMap map[string] map[string] global_structs.NugetPackageDetailsStruct) {
    global_vars.AppendPkgDetailsArr_Lock.Lock()
    map_1[key] = pkgDetailsStructMap
    global_vars.AppendPkgDetailsArr_Lock.Unlock()
}

func Synched_AddPkgDetailsStructSliceToMap(map_1 map[string] []global_structs.NugetPackageDetailsStruct, key string, pkgDetailsStructSlice [] global_structs.NugetPackageDetailsStruct) {
    global_vars.AppendPkgDetailsArr_Lock.Lock()
    map_1[key] = pkgDetailsStructSlice
    global_vars.AppendPkgDetailsArr_Lock.Unlock()
}

func Synched_AddPkgDetailsStructToMap(map_1 map[string] global_structs.NugetPackageDetailsStruct, pkgDetailsStruct global_structs.NugetPackageDetailsStruct) {
    global_vars.AppendPkgDetailsArr_Lock.Lock()
    map_1[pkgDetailsStruct.HashCode()] = pkgDetailsStruct
    global_vars.AppendPkgDetailsArr_Lock.Unlock()
}

func Synched_JoinTwoPkgDetailsSlices(arr_1 *[] global_structs.NugetPackageDetailsStruct, arr_2 [] global_structs.NugetPackageDetailsStruct) {
    global_vars.JoinTwoPkgDetailsSlices_Lock.Lock()
    *arr_1 = append(*arr_1, arr_2...)
    global_vars.JoinTwoPkgDetailsSlices_Lock.Unlock()
}

func Synched_JoinTwoPkgDetailsMaps(map_1 map[string] global_structs.NugetPackageDetailsStruct, map_2 map[string] global_structs.NugetPackageDetailsStruct) {
    global_vars.JoinTwoPkgDetailsMaps_Lock.Lock()
    for key, val := range map_2 {
        map_1[key] = val
    }
    global_vars.JoinTwoPkgDetailsMaps_Lock.Unlock()
}

func Synched_AppendDownloadedPkgDetailsObj(arr_1 *[] global_structs.DownloadPackageDetailsStruct, downloadPkgDetailsStruct global_structs.DownloadPackageDetailsStruct) {
    global_vars.AppendDownloadedPkgDetailsArr_Lock.Lock()
    *arr_1 = append(*arr_1, downloadPkgDetailsStruct)
    global_vars.AppendDownloadedPkgDetailsArr_Lock.Unlock()
}

func Synched_AppendUploadedPkgDetailsObj(arr_1 *[] global_structs.UploadPackageDetailsStruct, uploadedPkgDetailsStruct global_structs.UploadPackageDetailsStruct) {
    global_vars.AppendUploadedPkgDetailsArr_Lock.Lock()
    *arr_1 = append(*arr_1, uploadedPkgDetailsStruct)
    global_vars.AppendUploadedPkgDetailsArr_Lock.Unlock()
}


func CompareNugetPackageDetailsStruct(pkg1, pkg2 global_structs.NugetPackageDetailsStruct) bool {
    return (pkg1 == pkg2) || (strings.Compare(pkg1.Name, pkg2.Name) == 0 && strings.Compare(pkg1.Version, pkg2.Version) == 0)
}

func SortNugetPackageDetailsStructArr(nugetPackageDetailsStructArr [] global_structs.NugetPackageDetailsStruct) {
    if len(nugetPackageDetailsStructArr) == 0 {return}
	mylog.Logger.Infof("Sorting found nuget packages array")
    sort.Sort(global_structs.NugetPackageVersionSorter(nugetPackageDetailsStructArr))
    mylog.Logger.Infof("Done")
}

func FilterLastNPackages(nugetPackageDetailsStructArr [] global_structs.NugetPackageDetailsStruct, lastNCount int) [] global_structs.NugetPackageDetailsStruct {
    if (lastNCount <=0) {return nugetPackageDetailsStructArr}
    arrayLen := len(nugetPackageDetailsStructArr)
	sliceInd := arrayLen - lastNCount
    if sliceInd < 0 {sliceInd = 0}
    return nugetPackageDetailsStructArr[sliceInd:]
}

func DeleteLocalUploadedPackages(uploadedPkgsArr []global_structs.UploadPackageDetailsStruct) {
    downloadPkgsDir := global_vars.AppConfig.DownloadPkgsDirPath
    mylog.Logger.Infof("Removing all uploaded packages from: %s", downloadPkgsDir)
    if ! PathExists(downloadPkgsDir) {
        mylog.Logger.Warnf("Download dir doesn't exist at: %s", downloadPkgsDir)
        return
    }
    files, err := ioutil.ReadDir(downloadPkgsDir)
    if err != nil {
        mylog.Logger.Fatal(err)
    }

    // Assign un-uploaded packages names - Skip them on delete
    unUploadedFileNamesMap := map[string]bool{}
    fileUploadedIndicator := true
    for _, uploadedPkgStruct := range(uploadedPkgsArr) {
        if (uploadedPkgStruct.IsSuccessful) {continue}
        pkgDetails := uploadedPkgStruct.PkgDetailsStruct
        expectedFilename := Fmt_Sprintf("%s.%s.nupkg", strings.ToLower(pkgDetails.Name), pkgDetails.Version)
        unUploadedFileNamesMap[expectedFilename] = fileUploadedIndicator
    }

    // Loop on found files - Delete files that are NOT in the assigned map
    for _, file := range(files) {
        filename := file.Name()
        
        // If filename is in the map:
        if _, isMapContainsKey := unUploadedFileNamesMap[filename]; isMapContainsKey {
            mylog.Logger.Warnf("Skip delete of un-uploaded local file: %s", filename)
            continue
        } 

        // filename is NOT in the map:
        mylog.Logger.Debugf("Delete: %s", filename)
        fileToDeletePath := Filepath_Join(downloadPkgsDir, filename)
        err := os.Remove(fileToDeletePath)
        if (err != nil) {
            mylog.Logger.Panicf("\n%s\nFailed removing: %s", err, filename)
        }
        
    }

    mylog.Logger.Info("Done")
}

func DeleteLocalUploadedPackagesChecksumFiles(uploadedPkgsArr []global_structs.UploadPackageDetailsStruct) {
    checksumFilesDirPath := global_vars.AppConfig.ChecksumFilesDirPath
    mylog.Logger.Infof("Removing all uploaded packages's checksum files from: %s", checksumFilesDirPath)
    if ! PathExists(checksumFilesDirPath) {
        mylog.Logger.Warnf("Checksum files dir doesn't exist at: %s", checksumFilesDirPath)
        return
    }
    files, err := ioutil.ReadDir(checksumFilesDirPath)
    if err != nil {
        mylog.Logger.Fatal(err)
    }

    // Assign un-uploaded packages names - Skip them on delete
    unUploadedFileNamesMap := map[string]bool{}
    fileUploadedIndicator := true
    for _, uploadedPkgStruct := range(uploadedPkgsArr) {
        if (uploadedPkgStruct.IsSuccessful) {continue}
        pkgDetails := uploadedPkgStruct.PkgDetailsStruct
        expectedFilename := Fmt_Sprintf("%s.%s_checksum.txt", strings.ToLower(pkgDetails.Name), pkgDetails.Version)
        unUploadedFileNamesMap[expectedFilename] = fileUploadedIndicator
    }

    // Loop on found files - Delete files that are NOT in the assigned map
    for _, file := range(files) {
        filename := file.Name()
        
        // If filename is in the map:
        if _, isMapContainsKey := unUploadedFileNamesMap[filename]; isMapContainsKey {
            mylog.Logger.Warnf("Skip delete of un-uploaded local file: %s", filename)
            continue
        } 

        // filename is NOT in the map:
        mylog.Logger.Debugf("Delete: %s", filename)
        fileToDeletePath := Filepath_Join(checksumFilesDirPath, filename)
        err := os.Remove(fileToDeletePath)
        if (err != nil) {
            mylog.Logger.Panicf("\n%s\nFailed removing: %s", err, filename)
        }
        
    }

    mylog.Logger.Info("Done")
}

func StartTimer() {
    startTime = time.Now()
}

func EndTimer() time.Duration {
    return time.Since(startTime)
}

func PrintFinishSummary(filteredFoundPackagesDetailsList []global_structs.NugetPackageDetailsStruct, downloadedPkgsArr []global_structs.DownloadPackageDetailsStruct, uploadedPkgsArr []global_structs.UploadPackageDetailsStruct) {
    failedDownloadingPkgsCount := 0
	failedUploadingPkgsCount := 0
	for _, downloadPkgStruct := range downloadedPkgsArr {
		if (downloadPkgStruct.IsSuccessful) {continue}
		failedDownloadingPkgsCount += 1
	}
	for _, uploadedPkgsStruct := range uploadedPkgsArr {
		if (uploadedPkgsStruct.IsSuccessful) {continue}
		failedUploadingPkgsCount += 1
	}
	mylog.Logger.Info("")
	appConfigStr := global_vars.AppConfig.ToString()
    mylog.Logger.Infof("Configuration: \n%s", appConfigStr)

	mylog.Logger.Info("")
	mylog.Logger.Info("Summary:")
	mylog.Logger.Infof(" Targeted %d packages:", len(filteredFoundPackagesDetailsList))
	for i, pkgDetailsStruct := range filteredFoundPackagesDetailsList {
		mylog.Logger.Infof("  %d) %s", i+1, pkgDetailsStruct.HashCode())
	}
	
	if (failedDownloadingPkgsCount > 0) {
		mylog.Logger.Warnf(" Failed downloading %d packages:", failedDownloadingPkgsCount)
		for _, downloadPkgStruct := range downloadedPkgsArr {
			if (downloadPkgStruct.IsSuccessful) {continue}
			mylog.Logger.Infof("  - %s", downloadPkgStruct.PkgDetailsStruct.HashCode())
		}
	}

	if (failedUploadingPkgsCount > 0) {
		mylog.Logger.Warnf(" Failed uploading %d packages:", failedUploadingPkgsCount)
		for _, uploadedPkgsStruct := range uploadedPkgsArr {
			if (uploadedPkgsStruct.IsSuccessful) {continue}
			mylog.Logger.Infof("  * %s", uploadedPkgsStruct.PkgDetailsStruct.HashCode())
		}
        mylog.Logger.Panicf("Failed uploading %d packages", failedUploadingPkgsCount)
	}
	mylog.Logger.Info("")
	mylog.Logger.Info("Done")
	mylog.Logger.Info("Finished")
	duration := EndTimer()
	mylog.Logger.Infof("Time: %v", duration)
	mylog.Logger.Info("")
}

func HandlePanicErrors() {
	if r := recover(); r != nil {
        Fmt_Println(Fmt_Sprintf("\n%v\nProgram failed with a panic error. Printing error stack ", r))
        DeleteSuccessIndicatorFile()
        panic(r)
    }
}

func DeleteSuccessIndicatorFile() {
    if ! PathExists(global_vars.SuccessIndicatorFile) {return}
    mylog.Logger.Debugf("Removing file: %s", global_vars.SuccessIndicatorFile)
    err := os.Remove(global_vars.SuccessIndicatorFile)
    if (err != nil) {
        mylog.Logger.Panicf("\n%s\nFailed removing file: %s", err, global_vars.SuccessIndicatorFile)
    }
}

func WriteSuccessIndicatorFile() {
    dt := time.Now()
    msg := Fmt_Sprintf("%s - Syncher Finished Executing Successfully", dt.Format("01-02-2006 15:04:05 Monday"))
    err := os.WriteFile(global_vars.SuccessIndicatorFile, []byte(msg), 0755)  // <- 0755 =  default permissions: -rwxr-xr-x
    if err != nil {
        mylog.Logger.Panicf("Unable to write file: %v", err)
    }
}

