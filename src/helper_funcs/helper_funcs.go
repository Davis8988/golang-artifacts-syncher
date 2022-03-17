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

func InitVars() {
    mylog.Logger.Info("Initializing helpers pkg vars")
    global_vars.ConvertSyncedMapToString_Lock      = sync.RWMutex{}
    global_vars.JoinTwoPkgDetailsSlices_Lock       = sync.RWMutex{}
    global_vars.JoinTwoPkgDetailsMaps_Lock         = sync.RWMutex{}
    global_vars.AppendDownloadedPkgDetailsArr_Lock = sync.RWMutex{}
    global_vars.AppendPkgDetailsArr_Lock           = sync.RWMutex{}
    global_vars.AppendPkgDetailsMap_Lock           = sync.RWMutex{}

    mylog.Logger.Info("Initializing from envs vars")
    global_vars.SrcServersUserToUse                = Getenv("SRC_SERVERS_USER_TO_USE", "")
    global_vars.SrcServersPassToUse                = Getenv("SRC_SERVERS_PASS_TO_USE", "")
    global_vars.SrcServersUrlsStr                  = Getenv("SRC_SERVERS_URLS_STR", "")
    global_vars.DestServersUrlsStr                 = Getenv("DEST_SERVERS_URLS_STR", "")
    global_vars.DestServersUserToUse               = Getenv("DEST_SERVERS_USER_TO_USE", "")
    global_vars.DestServersPassToUse               = Getenv("DEST_SERVERS_PASS_TO_USE", "")
    global_vars.PackagesNamesStr                   = Getenv("PACKAGES_NAMES_STR", "")
    global_vars.PackagesVersionsStr                = Getenv("PACKAGES_VERSIONS_STR", "")
    global_vars.HttpRequestHeadersStr              = Getenv("HTTP_REQUEST_HEADERS_STR", "") // Example: "key=value;key1=value1;key2=value2"
    global_vars.DownloadPkgsDirPath                = Getenv("DOWNLOAD_PKGS_DIR_PATH", filepath.Join(GetCurrentProgramDir(), "Downloads"))
    global_vars.HttpRequestTimeoutSecondsInt       = StrToInt(Getenv("HTTP_REQUEST_TIMEOUT_SECONDS_INT", "45"))
    global_vars.SearchPackagesUrlSkipGroupCount    = StrToInt(Getenv("SEARCH_PACKAGES_URL_SKIP_GROUP_COUNT", "30"))
    global_vars.PackagesMaxConcurrentDownloadCount = StrToInt(Getenv("PACKAGES_MAX_CONCURRENT_DOWNLOAD_COUNT", "5"))
    global_vars.PackagesMaxConcurrentUploadCount   = StrToInt(Getenv("PACKAGES_MAX_CONCURRENT_UPLOAD_COUNT", "5"))
    global_vars.PackagesMaxConcurrentDeleteCount   = StrToInt(Getenv("PACKAGES_MAX_CONCURRENT_DELETE_COUNT", "5"))
    global_vars.PackagesDownloadLimitCount         = StrToInt(Getenv("PACKAGES_DOWNLOAD_LIMIT_COUNT", "-1"))
}

func PrintVars() {
    mylog.Logger.Infof("SRC_SERVERS_URLS_STR: '%s'"            , global_vars.SrcServersUrlsStr)
	mylog.Logger.Infof("SRC_SERVERS_USER_TO_USE: '%s'"         , global_vars.SrcServersUserToUse)
	mylog.Logger.Infof("SRC_SERVERS_PASS_TO_USE: '%s'"         , strings.Repeat("*", len(global_vars.SrcServersPassToUse)))
	mylog.Logger.Infof("DEST_SERVERS_URLS_STR: '%s'"           , global_vars.DestServersUrlsStr)
	mylog.Logger.Infof("DEST_SERVERS_USER_TO_USE: '%s'"        , global_vars.DestServersUserToUse)
	mylog.Logger.Infof("DEST_SERVERS_PASS_TO_USE: '%s'"        , strings.Repeat("*", len(global_vars.DestServersPassToUse)))
	mylog.Logger.Infof("PACKAGES_NAMES_STR: '%s'"              , global_vars.PackagesNamesStr)
	mylog.Logger.Infof("PACKAGES_VERSIONS_STR: '%s'"           , global_vars.PackagesVersionsStr)
	mylog.Logger.Infof("HTTP_REQUEST_HEADERS_STR: '%s'"        , global_vars.HttpRequestHeadersStr)
	mylog.Logger.Infof("DOWNLOAD_PKGS_DIR_PATH: '%s'"          , global_vars.DownloadPkgsDirPath)
	mylog.Logger.Infof("HTTP_REQUEST_TIMEOUT_SECONDS_INT: '%d'", global_vars.HttpRequestTimeoutSecondsInt)

	mylog.Logger.Infof("srcServersUrlsArr: %v"      , global_vars.SrcServersUrlsArr)
	mylog.Logger.Infof("DestServersUrlsArr: %v"     , global_vars.DestServersUrlsArr)
	mylog.Logger.Infof("packagesNamesArr: %v"       , global_vars.PackagesNamesArr)
	mylog.Logger.Infof("packagesVersionsArr: %v"    , global_vars.PackagesVersionsArr)
	
	packagesToDownloadMapStr := Synched_ConvertSyncedMapToString(global_vars.PackagesToDownloadMap)
	mylog.Logger.Infof("packagesToDownloadMap: \n%v", packagesToDownloadMapStr)
}

func ValidateEnvironment() {
    mylog.Logger.Info("Validating envs")

	// Validate len(packagesVersionsArr) == len(packagesNamesArr)  (Only when packagesVersionsArr is defined)
	if ! IsStrArrayEmpty(global_vars.PackagesVersionsArr) {
		mylog.Logger.Debug("Comparing packages names & versions arrays lengths")
		if len(global_vars.PackagesVersionsArr) != len(global_vars.PackagesNamesArr) {
			errMsg := "Packages Versions to search count is different from Packages Names to search count\n"
			errMsg += "Can't search for packages versions & names which are not of the same count.\n"
			errMsg += "When passing packages versions to search - the versions count must be of the same count of packages names to search.\n"
			errMsg += "A version for each package name to search"
			mylog.Logger.Fatal(errMsg)
		}
	}

    mylog.Logger.Debug("Validating all src URLs addresses: %s", global_vars.SrcServersUrlsArr)
    for i, srcServerUrl := range global_vars.SrcServersUrlsArr {
        if len(srcServerUrl) == 1 {continue}
        lastChar := srcServerUrl[len(srcServerUrl)-1:]
        if lastChar == "/" {continue}
        mylog.Logger.Debugf("Fix: Adding '/' char to src server repo url: \"%s\"", srcServerUrl)
        srcServerUrl += "/"
        global_vars.SrcServersUrlsArr[i] = srcServerUrl
    }
    
    mylog.Logger.Debug("Validating all dest URLs addresses: %s", global_vars.DestServersUrlsArr)
    for i, destServerUrl := range global_vars.DestServersUrlsArr {
        if len(destServerUrl) == 1 {continue}
        lastChar := destServerUrl[len(destServerUrl)-1:]
        if lastChar == "/" {continue}
        mylog.Logger.Debugf("Fix: Adding '/' char to src server repo url: \"%s\"", destServerUrl)
        destServerUrl += "/"
        global_vars.DestServersUrlsArr[i] = destServerUrl
    }

	mylog.Logger.Info("All Good")
}

func UpdateVars() {
    mylog.Logger.Info("Updating vars")
	global_vars.SrcServersUrlsArr = make([]string, 0, 4)
	global_vars.DestServersUrlsArr = make([]string, 0, 4)
	global_vars.PackagesNamesArr = make([]string, 0, 10)
	global_vars.PackagesVersionsArr = make([]string, 0, 10)
	if len(global_vars.SrcServersUrlsStr)   > 1 {global_vars.SrcServersUrlsArr   = strings.Split(global_vars.SrcServersUrlsStr,   ";")}
	if len(global_vars.DestServersUrlsStr)  > 1 {global_vars.DestServersUrlsArr  = strings.Split(global_vars.DestServersUrlsStr,  ";")}
	if len(global_vars.PackagesNamesStr)    > 1 {global_vars.PackagesNamesArr    = strings.Split(global_vars.PackagesNamesStr,    ";")}
	if len(global_vars.PackagesVersionsStr) > 1 {global_vars.PackagesVersionsArr = strings.Split(global_vars.PackagesVersionsStr, ";")}
	global_vars.HttpRequestHeadersMap = ParseHttpHeadersStrToMap(global_vars.HttpRequestHeadersStr)

	for i, pkgName := range global_vars.PackagesNamesArr {
		// If map doesn't contain value at: 'pkgName' - add one to point to empty string array: []
		global_vars.PackagesToDownloadMap.LoadOrStore(pkgName, make([]string, 0, 10))
		// If received a version array for it - add it to the list
		if len(global_vars.PackagesVersionsArr) > i {
			pkgVersion := global_vars.PackagesVersionsArr[i]
			currentVersionsArr := LoadStringArrValueFromSynchedMap(global_vars.PackagesToDownloadMap, pkgName)
			global_vars.PackagesToDownloadMap.Store(pkgName, append(currentVersionsArr, pkgVersion))
		}
	}
}

func PrepareSrcSearchUrlsForPackageArray(pkgName string) []string {
	var searchUrlsArr = make([]string, 0, 10) // Create a slice with length=0 and capacity=10

	mylog.Logger.Info("Preparing src search packages urls array")
	for _, srcServerUrl := range global_vars.SrcServersUrlsArr {
        if len(srcServerUrl) == 1 {continue}
        versionsToSearchArr := LoadStringArrValueFromSynchedMap(global_vars.PackagesToDownloadMap, pkgName)
        if len(versionsToSearchArr) == 0 { // Either use search
            searchUrlsArr = append(searchUrlsArr, srcServerUrl + "Packages()?$filter=tolower(Id)%20eq%20'"+pkgName+"'")
            continue
        } // Or specific package details request for each specified requested version
        for _, pkgVersion := range versionsToSearchArr {
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
        mylog.Logger.Errorf("\n%s\nFailed converting string: \"%s\" to integer\n", err, strVar)
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
            mylog.Logger.Errorf("\nFound header pair: \"%v\"  that is not in the right format of: \"key=value\"\n", tempPairArr)
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

func CreateDir(dirPath string) {
    if _, err := os.Stat(dirPath); err == nil {return}  // If dir already exists - finish here
    mylog.Logger.Debugf("Creating dir: %s", dirPath)
    err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		mylog.Logger.Errorf("\n%s\nFailed creating dir: \"%s\"\n", err, dirPath)
        panic(err)
	}
}

func CreateFile(filePath string) *os.File {
    dirPath := filepath.Dir(filePath)
    CreateDir(dirPath)
    mylog.Logger.Debugf("Creating file: %s", filePath)
    // Create the file
    file, err := os.Create(filePath)
    if err != nil  {
        mylog.Logger.Errorf("\n%s\nFailed creating file: \"%s\"\n", err, filePath)
        panic(err)
    }
    return file
}

func CalculateFileChecksum(filePath string) string {
    if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {return ""}  // If missing file: return empty
    mylog.Logger.Debugf("Calculating sha512 checksum of file: %s", filePath)
    f, err := os.Open(filePath)
    if err != nil {
        mylog.Logger.Errorf("\n%s\nFailed calculating sha512 checksum of file: \"%s\"\n", err, filePath)
        panic(err)
    }
    defer f.Close()
    h := sha512.New()
    if _, err := io.Copy(h, f); err != nil {
        mylog.Logger.Errorf("\n%s\nFailed calculating sha512 checksum of file: \"%s\"\n", err, filePath)
        panic(err)
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
        mylog.Logger.Errorf("\n%s\nFailed creating HTTP request object for URL: \"%s\"\n", err, urlAddress)
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
        mylog.Logger.Infof("Adding creds of user:  '%s'", username)
        req.SetBasicAuth(username, password)
    }

    // Make the http request
    response, err := client.Do(req)
    if err != nil {
        mylog.Logger.Errorf("\n%s\n", err)
        return nil
    }
  
    defer response.Body.Close() // Finally step: close the body obj
    
    // If got: downloadFilePath var, then Writer the body to file
    if len(downloadFilePath) > 0 {
        mylog.Logger.Infof("Downloading '%s' to:  %s", urlAddress, downloadFilePath)
        fileHandle := CreateFile(downloadFilePath)  // Create the file
        defer fileHandle.Close()

        _, err = io.Copy(fileHandle, response.Body)
        if err != nil  {
            mylog.Logger.Errorf("\n%s\nFailed writing response Body to file: %s\n", err, downloadFilePath)
            panic(err)
        }
        return nil // Finish here
    }

    responseBody, err := ioutil.ReadAll(response.Body)
    if err != nil {
        mylog.Logger.Errorf("\n%s\nFailed reading request's response body: %s\n", err, urlAddress)
        return nil
    }

    bodyStr := string(responseBody)
    httpResponseResultStruct := global_structs.HttpResponseStruct {
        UrlAddress: urlAddress,
        BodyStr: bodyStr,
        StatusStr: response.Status,
        StatusCode: response.StatusCode,
    }

    return &httpResponseResultStruct
}

func ReadFileContentsIntoPartsForUpload(uploadFilePath string, headerFieldName string) (io.Reader, *multipart.Writer) {
    mylog.Logger.Infof("Reading file content for upload: \"%s\"", uploadFilePath)

    // If missing file: return empty body
    if _, err := os.Stat(uploadFilePath); errors.Is(err, os.ErrNotExist) {
        mylog.Logger.Errorf("\n%s\nFailed uploading file: \"%s\" since it is missing. Failed preparing HTTP request object\n", err, uploadFilePath)
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

func DeleteLocalUnuploadedPackages(uploadedPkgsArr []global_structs.DownloadPackageDetailsStruct) {
    if len(uploadedPkgsArr) == 0 {return}
    downloadPkgsDir := global_vars.DownloadPkgsDirPath
    mylog.Logger.Infof("Removing all unuploaded packages from: %s", downloadPkgsDir)
    files, err := ioutil.ReadDir(downloadPkgsDir)
    if err != nil {
        mylog.Logger.Fatal(err)
    }
    
    // Assign uploaded packages names
    uploadedFileNamesMap := map[string]int{}
    fileUploadedIndicator := 1
    for _, pkg := range(uploadedPkgsArr) {
        pkgDetails := pkg.PkgDetailsStruct
        expectedFilename := Fmt_Sprintf("%s.%s.nupkg", strings.ToLower(pkgDetails.Name), pkgDetails.Version)
        uploadedFileNamesMap[expectedFilename] = fileUploadedIndicator
    }

    // Loop on found files - Delete files that are NOT in the assigned map
    for _, file := range(files) {
        filename := file.Name()
        
        // If filename is not in the map:
        if _, isMapContainsKey := uploadedFileNamesMap[filename]; ! isMapContainsKey {
            mylog.Logger.Debugf("Delete: %s", filename)
            fileToDeletePath := filepath.Join(downloadPkgsDir, filename)
            err := os.Remove(fileToDeletePath)
            if (err != nil) {
                mylog.Logger.Errorf("\n%s\nFailed removing: %s", err, filename)
            }
        } else {
            // filename is in the map:
            mylog.Logger.Warnf("Skip delete local file: %s", filename)
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


