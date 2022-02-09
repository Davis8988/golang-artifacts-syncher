package helper_funcs

import (
	"golang-artifacts-syncher/src/nuget_packages_xml"
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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func InitVars() {
    mylog.LogInfo.Print("Initializing helpers pkg vars")
    global_vars.ConvertSyncedMapToString_Lock = sync.RWMutex{}
    global_vars.AppendPkgDetailsArr_Lock = sync.RWMutex{}
    global_vars.AppendDownloadedPkgDetailsArr_Lock = sync.RWMutex{}

    mylog.LogInfo.Print("Initializing from envs vars")
    global_vars.SrcServersUserToUse = Getenv("SRC_SERVERS_USER_TO_USE", "")
	global_vars.SrcServersPassToUse = Getenv("SRC_SERVERS_PASS_TO_USE", "")
	global_vars.SrcServersUrlsStr = Getenv("SRC_SERVERS_URLS_STR", "")
	global_vars.SrcReposNamesStr = Getenv("SRC_REPOS_NAMES_STR", "")
	global_vars.DestServersUrlsStr = Getenv("DEST_SERVERS_URLS_STR", "")
	global_vars.DestReposNamesStr = Getenv("DEST_REPOS_NAMES_STR", "")
	global_vars.DestServersUserToUse = Getenv("DEST_SERVERS_USER_TO_USE", "")
	global_vars.DestServersPassToUse = Getenv("DEST_SERVERS_PASS_TO_USE", "")
	global_vars.PackagesNamesStr = Getenv("PACKAGES_NAMES_STR", "")
	global_vars.PackagesVersionsStr = Getenv("PACKAGES_VERSIONS_STR", "")
	global_vars.HttpRequestHeadersStr = Getenv("HTTP_REQUEST_HEADERS_STR", "") // Example: "key=value;key1=value1;key2=value2"
	global_vars.DownloadPkgsDirPath = Getenv("DOWNLOAD_PKGS_DIR_PATH", GetCurrentProgramDir())
	global_vars.HttpRequestTimeoutSecondsInt = StrToInt(Getenv("HTTP_REQUEST_TIMEOUT_SECONDS_INT", "45"))
}

func PrintVars() {
    mylog.LogInfo.Printf("SRC_SERVERS_URLS_STR: '%s'", global_vars.SrcServersUrlsStr)
	mylog.LogInfo.Printf("SRC_REPOS_NAMES_STR: '%s'", global_vars.SrcReposNamesStr)
	mylog.LogInfo.Printf("SRC_SERVERS_USER_TO_USE: '%s'", global_vars.SrcServersUserToUse)
	mylog.LogInfo.Printf("SRC_SERVERS_PASS_TO_USE: '%s'", strings.Repeat("*", len(global_vars.SrcServersPassToUse)))
	mylog.LogInfo.Printf("DEST_SERVERS_URLS_STR: '%s'", global_vars.DestServersUrlsStr)
	mylog.LogInfo.Printf("DEST_REPOS_NAMES_STR: '%s'", global_vars.DestReposNamesStr)
	mylog.LogInfo.Printf("DEST_SERVERS_USER_TO_USE: '%s'", global_vars.DestServersUserToUse)
	mylog.LogInfo.Printf("DEST_SERVERS_PASS_TO_USE: '%s'", strings.Repeat("*", len(global_vars.DestServersPassToUse)))
	mylog.LogInfo.Printf("PACKAGES_NAMES_STR: '%s'", global_vars.PackagesNamesStr)
	mylog.LogInfo.Printf("PACKAGES_VERSIONS_STR: '%s'", global_vars.PackagesVersionsStr)
	mylog.LogInfo.Printf("HTTP_REQUEST_HEADERS_STR: '%s'", global_vars.HttpRequestHeadersStr)
	mylog.LogInfo.Printf("DOWNLOAD_PKGS_DIR_PATH: '%s'", global_vars.DownloadPkgsDirPath)
	mylog.LogInfo.Printf("HTTP_REQUEST_TIMEOUT_SECONDS_INT: '%d'", global_vars.HttpRequestTimeoutSecondsInt)

	mylog.LogInfo.Printf("srcServersUrlsArr: %v", global_vars.SrcServersUrlsArr)
	mylog.LogInfo.Printf("DestServersUrlsArr: %v", global_vars.DestServersUrlsArr)
	mylog.LogInfo.Printf("srcReposNamesArr: %v", global_vars.SrcReposNamesArr)
	mylog.LogInfo.Printf("packagesNamesArr: %v", global_vars.PackagesNamesArr)
	mylog.LogInfo.Printf("packagesVersionsArr: %v", global_vars.PackagesVersionsArr)
	packagesToDownloadMapStr := Synched_ConvertSyncedMapToString(global_vars.PackagesToDownloadMap)
	mylog.LogInfo.Printf("packagesToDownloadMap: \n%v", packagesToDownloadMapStr)
}

func ValidateEnvironment() {
    mylog.LogInfo.Print("Validating envs")

	// Validate len(packagesVersionsArr) == len(packagesNamesArr)  (Only when packagesVersionsArr is defined)
	if ! IsStrArrayEmpty(global_vars.PackagesVersionsArr) {
		mylog.LogInfo.Print("Comparing packages names & versions arrays lengths")
		if len(global_vars.PackagesVersionsArr) != len(global_vars.PackagesNamesArr) {
			errMsg := "Packages Versions to search count is different from Packages Names to search count\n"
			errMsg += "Can't search for packages versions & names which are not of the same count.\n"
			errMsg += "When passing packages versions to search - the versions count must be of the same count of packages names to search.\n"
			errMsg += "A version for each package name to search"
			mylog.LogError.Fatal(errMsg)
		}
	}

	mylog.LogInfo.Print("All Good")
}

func UpdateVars() {
    mylog.LogInfo.Print("Updating vars")
	global_vars.SrcServersUrlsArr = make([]string, 0, 4)
	global_vars.DestServersUrlsArr = make([]string, 0, 4)
	global_vars.SrcReposNamesArr = make([]string, 0, 4)
	global_vars.PackagesNamesArr = make([]string, 0, 10)
	global_vars.PackagesVersionsArr = make([]string, 0, 10)
	if len(global_vars.SrcServersUrlsStr) > 1 {global_vars.SrcServersUrlsArr = strings.Split(global_vars.SrcServersUrlsStr, ";")}
	if len(global_vars.SrcReposNamesStr) > 1 {global_vars.SrcReposNamesArr = strings.Split(global_vars.SrcReposNamesStr, ";")}
	if len(global_vars.DestServersUrlsStr) > 1 {global_vars.DestServersUrlsArr = strings.Split(global_vars.DestServersUrlsStr, ";")}
	if len(global_vars.DestReposNamesStr) > 1 {global_vars.DestReposNamesArr = strings.Split(global_vars.DestReposNamesStr, ";")}
	if len(global_vars.PackagesNamesStr) > 1 {global_vars.PackagesNamesArr = strings.Split(global_vars.PackagesNamesStr, ";")}
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

func PrepareSrcSearchAllPkgsVersionsUrlsArray() []string {
	var searchUrlsArr = make([]string, 0, 10) // Create a slice with length=0 and capacity=10

	mylog.LogInfo.Print("Preparing src search packages urls array")
	for _, srcServerUrl := range global_vars.SrcServersUrlsArr {
		for _, repoName := range global_vars.SrcReposNamesArr {
			for _, pkgName := range global_vars.PackagesNamesArr {
				versionsToSearchArr := LoadStringArrValueFromSynchedMap(global_vars.PackagesToDownloadMap, pkgName)
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
	mylog.LogInfo.Printf("Filtering found pkgs by requested versions")
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


func UploadDownloadedPackage(uploadPkgStruct global_structs.UploadPackageDetailsStruct) global_structs.UploadPackageDetailsStruct {
	pkgPrintStr := fmt.Sprintf("%s==%s", uploadPkgStruct.PkgDetailsStruct.Name, uploadPkgStruct.PkgDetailsStruct.Version)
	pkgName := uploadPkgStruct.PkgDetailsStruct.Name
	pkgVersion := uploadPkgStruct.PkgDetailsStruct.Version

	// Check if package already exists. If so, then compare it's checksum and skip on matching
	for _, destServerUrl := range global_vars.DestServersUrlsArr {
		for _, repoName := range global_vars.DestReposNamesArr {
			destServerRepo := destServerUrl + "/" + repoName
			mylog.LogInfo.Printf("Checking if pkg: '%s' already exists at dest server: %s", pkgPrintStr, destServerRepo)
			checkDestServerPkgExistUrl := destServerRepo + "/" + "Packages(Id='" + pkgName + "',Version='" + pkgVersion + "')"
			httpRequestArgs := global_structs.HttpRequestArgsStruct{
				UrlAddress: checkDestServerPkgExistUrl,
				HeadersMap: global_vars.HttpRequestHeadersMap,
				UserToUse:  global_vars.DestServersUserToUse,
				PassToUse:  global_vars.DestServersPassToUse,
				TimeoutSec: global_vars.HttpRequestTimeoutSecondsInt,
				Method:     "GET",
			}

			foundPackagesDetailsArr := SearchPackagesAvailableVersionsByURLRequest(httpRequestArgs)
			mylog.LogInfo.Printf("Found: %s", foundPackagesDetailsArr)

			emptyNugetPackageDetailsStruct := global_structs.NugetPackageDetailsStruct{}
			shouldCompareChecksum := true
			if len(foundPackagesDetailsArr) != 1 {
				mylog.LogInfo.Printf("Found multiple or no packages: \"%d\" - Should be only 1. Skipping checksum comparison. Continuing with the upload..", len(foundPackagesDetailsArr))
				shouldCompareChecksum = false
			} else if len(foundPackagesDetailsArr) == 1 && foundPackagesDetailsArr[0] == emptyNugetPackageDetailsStruct {
				mylog.LogInfo.Print("No package found. Continuing with the upload..")
				shouldCompareChecksum = false
			}
			
			if shouldCompareChecksum {
				// Check the checksum:
				mylog.LogInfo.Printf("Comparing found package's checksum to know if should upload to: %s or not", destServerRepo)
				foundPackageChecksum := foundPackagesDetailsArr[0].Checksum
				fileToUploadChecksum := uploadPkgStruct.UploadFileChecksum
				if foundPackageChecksum == fileToUploadChecksum {
				fileName := filepath.Base(uploadPkgStruct.UploadFilePath)
				mylog.LogWarning.Printf("Checksum match: upload target file already exists in dest server: '%s' \n"+
					"Skipping upload of pkg: \"%s\"", destServerRepo, fileName)
				return uploadPkgStruct
				}
			}
			
			if len(destServerRepo) > 1 {
				lastChar := destServerRepo[len(destServerRepo)-1:]
				mylog.LogInfo.Printf("Adding '/' char to dest server repo url: \"%s\"", destServerRepo)
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
        mylog.LogError.Fatalf("%s\nFailed getting current program's dir", err)
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
        mylog.LogError.Printf("%s\nFailed converting string: \"%s\" to integer", err, strVar)
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
    mylog.LogInfo.Printf("Looping on headers values to init headers map")
    for _, headersPairStr := range tempHeadersPairsArr {
        tempPairArr := strings.Split(headersPairStr, "=")
        if len(tempPairArr) != 2 {
            mylog.LogError.Printf("Found header pair: \"%v\"  that is not in the right format of: \"key=value\"", tempPairArr)
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
    mylog.LogDebug.Printf("Creating dir: %s", dirPath)
    err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		mylog.LogError.Printf("%s\nFailed creating dir: \"%s\"", err, dirPath)
        panic(err)
	}
}

func CreateFile(filePath string) *os.File {
    dirPath := filepath.Dir(filePath)
    CreateDir(dirPath)
    mylog.LogDebug.Printf("Creating file: %s", filePath)
    // Create the file
    file, err := os.Create(filePath)
    if err != nil  {
        mylog.LogError.Printf("%s\nFailed creating file: \"%s\"", err, filePath)
        panic(err)
    }
    return file
}

func CalculateFileChecksum(filePath string) string {
    if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {return ""}  // If missing file: return empty
    mylog.LogDebug.Printf("Calculating sha512 checksum of file: %s", filePath)
    f, err := os.Open(filePath)
    if err != nil {
        mylog.LogError.Printf("%s\nFailed calculating sha512 checksum of file: \"%s\"", err, filePath)
        panic(err)
    }
    defer f.Close()
    h := sha512.New()
    if _, err := io.Copy(h, f); err != nil {
        mylog.LogError.Printf("%s\nFailed calculating sha512 checksum of file: \"%s\"", err, filePath)
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

    mylog.LogInfo.Printf("Making an HTTP '%s' request to URL: \"%s\"", method, urlAddress)

    client := http.Client{Timeout: time.Duration(timeoutSec) * time.Second,}
    
    var body io.Reader
    var writer *multipart.Writer

    // Upload file (PUT requests):
    if method == "PUT" && len(uploadFilePath) > 0 {
        body, writer = ReadFileContentsIntoPartsForUpload(uploadFilePath, "package")
    }

    req, err := http.NewRequest(method, urlAddress, body)
    if err != nil {
        mylog.LogError.Printf("%s\nFailed creating HTTP request object for URL: \"%s\"", err, urlAddress)
        return ""
    }

    // Incase pushing a file, then add the Content Type header from the reader (includes boundary)
    if method == "PUT" && len(uploadFilePath) > 0 {
        mylog.LogInfo.Printf("Adding header:  'Content-Type'")
        req.Header.Add("Content-Type", writer.FormDataContentType())
    }

    // Adding headers:
    for k := range headersMap {
        mylog.LogInfo.Printf("Adding header:  '%s'=\"%s\"", k, headersMap[k])
        req.Header.Add(k, headersMap[k])
    }

    // Adding creds
    if len(username) > 0 && len(password) > 0 {
        mylog.LogInfo.Printf("Adding creds of user:  '%s'", username)
        req.SetBasicAuth(username, password)
    }

    // Make the http request
    response, err := client.Do(req)
    if err != nil {
        mylog.LogError.Printf("%s\nFailed while making the request: %s", err, urlAddress)
        return ""
    }
  
    defer response.Body.Close() // Finally step: close the body obj
    
    // If got: downloadFilePath var, then Writer the body to file
    if len(downloadFilePath) > 0 {
        mylog.LogInfo.Printf("Downloading '%s' to:  %s", urlAddress, downloadFilePath)
        fileHandle := CreateFile(downloadFilePath)  // Create the file
        defer fileHandle.Close()

        _, err = io.Copy(fileHandle, response.Body)
        if err != nil  {
            mylog.LogError.Printf("%s\nFailed writing response Body to file: %s", err, downloadFilePath)
            panic(err)
        }
        return "" // Finish here
    }

    responseBody, err := ioutil.ReadAll(response.Body)
    if err != nil {
        mylog.LogError.Printf("%s\nFailed reading request's response body: %s", err, urlAddress)
        return ""
    }

    bodyStr := string(responseBody)
    msgStr := bodyStr
    if len(response.Status) > 0 {msgStr = fmt.Sprintf("%s  %s", response.Status, bodyStr)}
    // mylog.LogDebug.Printf(msgStr)

    if response.StatusCode >= 400 {
        mylog.LogError.Printf("%s", msgStr)
        mylog.LogError.Printf("Returned code: %d. HTTP request failure: %s", response.StatusCode, urlAddress)
    }

    return bodyStr
}

func ReadFileContentsIntoPartsForUpload(uploadFilePath string, headerFieldName string) (io.Reader, *multipart.Writer) {
    mylog.LogInfo.Printf("Reading file content for upload: \"%s\"", uploadFilePath)

    // If missing file: return empty body
    if _, err := os.Stat(uploadFilePath); errors.Is(err, os.ErrNotExist) {
        mylog.LogError.Printf("%s\nFailed uploading file: \"%s\" since it is missing. Failed preparing HTTP request object", err, uploadFilePath)
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
    mylog.LogDebug.Printf("Parsing URL for Name & Version: \"%s\"", pkgDetailsUrl)
    re := regexp.MustCompile("'(.*?)'")  // Find values in between quotes
    resultArr := re.FindAllString(pkgDetailsUrl, -1)  // -1 = find ALL available matches
    if len(resultArr) != 2 {
        mylog.LogError.Printf("Failed to parse URL for pkg Name & Version:  \"%s\"", pkgDetailsUrl)
        mylog.LogError.Printf("Found regex result count is: %d different from 2", len(resultArr))
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
    mylog.LogInfo.Printf("Parsing http request response for packages details")
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
    mylog.LogInfo.Printf("Downloading package: %s==%s", downloadPkgDetailsStruct.PkgDetailsStruct.Name, downloadPkgDetailsStruct.PkgDetailsStruct.Version)
    fileUrl := downloadPkgDetailsStruct.PkgDetailsStruct.PkgFileUrl
    downloadFilePath := downloadPkgDetailsStruct.DownloadFilePath
    downloadFileChecksum := downloadPkgDetailsStruct.DownloadFileChecksum
    fileChecksum := downloadPkgDetailsStruct.PkgDetailsStruct.Checksum
    if fileChecksum == downloadFileChecksum {
        fileName := filepath.Base(downloadFilePath)
        mylog.LogWarning.Printf("Checksum match: download target file already exists. Skipping download of: \"%s\"", fileName)
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
	mylog.LogInfo.Printf("Uploading package: \"%s\" from: %s", pkgPrintStr, uploadPkgStruct.UploadFilePath)
    httpRequestArgsStruct.Method = "PUT"
    httpRequestArgsStruct.UploadFilePath = uploadPkgStruct.UploadFilePath
    MakeHttpRequest(httpRequestArgsStruct)

}

func Synched_ConvertSyncedMapToString(synchedMap sync.Map) string {
	global_vars.ConvertSyncedMapToString_Lock.Lock()
	result := ConvertSyncedMapToString(synchedMap)
	defer global_vars.ConvertSyncedMapToString_Lock.Unlock()
	return result
}

func Synched_AppendPkgDetailsObj(arr_1 *[] global_structs.NugetPackageDetailsStruct, arr_2 [] global_structs.NugetPackageDetailsStruct) {
    global_vars.AppendPkgDetailsArr_Lock.Lock()
    *arr_1 = append(*arr_1, arr_2...)
    global_vars.AppendPkgDetailsArr_Lock.Unlock()
}

func Synched_AppendDownloadedPkgDetailsObj(arr_1 *[] global_structs.DownloadPackageDetailsStruct, downloadPkgDetailsStruct global_structs.DownloadPackageDetailsStruct) {
    global_vars.AppendDownloadedPkgDetailsArr_Lock.Lock()
    *arr_1 = append(*arr_1, downloadPkgDetailsStruct)
    global_vars.AppendDownloadedPkgDetailsArr_Lock.Unlock()
}
