package nuget_packages_xml

import (
    "encoding/xml"
)

// Feed was generated 2022-01-31 20:01:23 using 'zek'
type MultiplePackagesDetailsXmlStruct struct {
    XMLName xml.Name `xml:"feed"`
    Text    string   `xml:",chardata"`
    Base    string   `xml:"base,attr"`
    D       string   `xml:"d,attr"`
    M       string   `xml:"m,attr"`
    Xmlns   string   `xml:"xmlns,attr"`
    Title   struct {
            Text string `xml:",chardata"` // Search
            Type string `xml:"type,attr"`
    } `xml:"title"`
    ID      string `xml:"id"`      // http://localhost:8081/rep...
    Updated string `xml:"updated"` // 2022-01-31T18:01:20.608Z
    Link    struct {
            Text  string `xml:",chardata"`
            Rel   string `xml:"rel,attr"`
            Title string `xml:"title,attr"`
            Href  string `xml:"href,attr"`
    } `xml:"link"`
    Entry []struct {
            Text  string `xml:",chardata"`
            ID    string `xml:"id"` // http://localhost:8081/rep...
            Title struct {
                    Text string `xml:",chardata"` // 7zip, 7zip, 7zip, 7zip, 7...
                    Type string `xml:"type,attr"`
            } `xml:"title"`
            Summary struct {
                    Text string `xml:",chardata"` // 7-Zip is a file archiver ...
                    Type string `xml:"type,attr"`
            } `xml:"summary"`
            Updated string `xml:"updated"` // 2022-01-22T09:08:07.154Z,...
            Author  struct {
                    Text string `xml:",chardata"`
                    Name string `xml:"name"` // Igor Pavlov, Igor Pavlov,...
            } `xml:"author"`
            Link []struct {
                    Text  string `xml:",chardata"`
                    Rel   string `xml:"rel,attr"`
                    Title string `xml:"title,attr"`
                    Href  string `xml:"href,attr"`
            } `xml:"link"`
            Category struct {
                    Text   string `xml:",chardata"`
                    Term   string `xml:"term,attr"`
                    Scheme string `xml:"scheme,attr"`
            } `xml:"category"`
            Content struct {
                    Text string `xml:",chardata"`
                    Type string `xml:"type,attr"`
                    Src  string `xml:"src,attr"`
            } `xml:"content"`
            Properties struct {
                    Text      string `xml:",chardata"`
                    Version   string `xml:"Version"` // 15.05, 15.06, 15.07, 15.0...
                    Copyright struct {
                            Text string `xml:",chardata"`
                            Null string `xml:"null,attr"`
                    } `xml:"Copyright"`
                    Created struct {
                            Text string `xml:",chardata"` // 2022-01-22T09:08:07.158Z,...
                            Type string `xml:"type,attr"`
                    } `xml:"Created"`
                    Dependencies  string `xml:"Dependencies"` // 7zip.install:[15.05], 7zi...
                    Description   string `xml:"Description"`  // 7-Zip is a file archiver ...
                    DownloadCount struct {
                            Text string `xml:",chardata"` // 0, 0, 0, 0, 0
                            Type string `xml:"type,attr"`
                    } `xml:"DownloadCount"`
                    GalleryDetailsUrl struct {
                            Text string `xml:",chardata"`
                            Null string `xml:"null,attr"`
                    } `xml:"GalleryDetailsUrl"`
                    IconUrl         string `xml:"IconUrl"` // https://cdn.rawgit.com/fe...
                    IsLatestVersion struct {
                            Text string `xml:",chardata"` // false, false, false, fals...
                            Type string `xml:"type,attr"`
                    } `xml:"IsLatestVersion"`
                    IsAbsoluteLatestVersion struct {
                            Text string `xml:",chardata"` // false, false, false, fals...
                            Type string `xml:"type,attr"`
                    } `xml:"IsAbsoluteLatestVersion"`
                    IsPrerelease struct {
                            Text string `xml:",chardata"` // false, false, false, fals...
                            Type string `xml:"type,attr"`
                    } `xml:"IsPrerelease"`
                    Published struct {
                            Text string `xml:",chardata"` // 2022-01-22T09:08:07.158Z,...
                            Type string `xml:"type,attr"`
                    } `xml:"Published"`
                    Language struct {
                            Text string `xml:",chardata"`
                            Null string `xml:"null,attr"`
                    } `xml:"Language"`
                    LicenseUrl           string `xml:"LicenseUrl"`           // http://www.7-zip.org/lice...
                    PackageHash          string `xml:"PackageHash"`          // 0IWbo8P8A2gJUufDUTJYAGVcb...
                    PackageHashAlgorithm string `xml:"PackageHashAlgorithm"` // SHA512, SHA512, SHA512, S...
                    PackageSize          struct {
                            Text string `xml:",chardata"` // 3055, 3056, 3055, 3056, 4...
                            Type string `xml:"type,attr"`
                    } `xml:"PackageSize"`
                    ProjectUrl     string `xml:"ProjectUrl"` // http://www.7-zip.org/, ht...
                    ReportAbuseUrl struct {
                            Text string `xml:",chardata"`
                            Null string `xml:"null,attr"`
                    } `xml:"ReportAbuseUrl"`
                    ReleaseNotes struct {
                            Text string `xml:",chardata"` // http://www.7-zip.org/hist...
                            Null string `xml:"null,attr"`
                    } `xml:"ReleaseNotes"`
                    RequireLicenseAcceptance struct {
                            Text string `xml:",chardata"` // false, false, false, fals...
                            Type string `xml:"type,attr"`
                    } `xml:"RequireLicenseAcceptance"`
                    Tags                 string `xml:"Tags"`  // 7zip zip archiver admin, ...
                    Title                string `xml:"Title"` // 7-Zip, 7-Zip, 7-Zip, 7-Zi...
                    VersionDownloadCount struct {
                            Text string `xml:",chardata"` // 0, 0, 0, 0, 0
                            Type string `xml:"type,attr"`
                    } `xml:"VersionDownloadCount"`
            } `xml:"properties"`
    } `xml:"entry"`
}

// Entry was generated 2022-02-03 17:54:45 by ELBIT_NT\E030331 on E030331N.
type SinglePackagesDetailsXmlStruct struct {
        XMLName xml.Name `xml:"entry"`
        Text    string   `xml:",chardata"`
        D       string   `xml:"d,attr"`
        M       string   `xml:"m,attr"`
        Xmlns   string   `xml:"xmlns,attr"`
        ID      string   `xml:"id"` // http://localhost:8081/rep...
        Title   struct {
                Text string `xml:",chardata"` // 7zip
                Type string `xml:"type,attr"`
        } `xml:"title"`
        Summary struct {
                Text string `xml:",chardata"` // 7-Zip is a file archiver ...
                Type string `xml:"type,attr"`
        } `xml:"summary"`
        Updated string `xml:"updated"` // 2022-01-22T09:08:11.110Z
        Author  struct {
                Text string `xml:",chardata"`
                Name string `xml:"name"` // Igor Pavlov
        } `xml:"author"`
        Link []struct {
                Text  string `xml:",chardata"`
                Rel   string `xml:"rel,attr"`
                Title string `xml:"title,attr"`
                Href  string `xml:"href,attr"`
        } `xml:"link"`
        Category struct {
                Text   string `xml:",chardata"`
                Term   string `xml:"term,attr"`
                Scheme string `xml:"scheme,attr"`
        } `xml:"category"`
        Content struct {
                Text string `xml:",chardata"`
                Type string `xml:"type,attr"`
                Src  string `xml:"src,attr"`
        } `xml:"content"`
        Properties struct {
                Text      string `xml:",chardata"`
                Version   string `xml:"Version"` // 15.07
                Copyright struct {
                        Text string `xml:",chardata"`
                        Null string `xml:"null,attr"`
                } `xml:"Copyright"`
                Created struct {
                        Text string `xml:",chardata"` // 2022-01-22T09:08:11.111Z
                        Type string `xml:"type,attr"`
                } `xml:"Created"`
                Dependencies  string `xml:"Dependencies"` // 7zip.install:[15.07]
                Description   string `xml:"Description"`  // 7-Zip is a file archiver ...
                DownloadCount struct {
                        Text string `xml:",chardata"` // 0
                        Type string `xml:"type,attr"`
                } `xml:"DownloadCount"`
                GalleryDetailsUrl struct {
                        Text string `xml:",chardata"`
                        Null string `xml:"null,attr"`
                } `xml:"GalleryDetailsUrl"`
                IconUrl         string `xml:"IconUrl"` // https://cdn.rawgit.com/fe...
                IsLatestVersion struct {
                        Text string `xml:",chardata"` // false
                        Type string `xml:"type,attr"`
                } `xml:"IsLatestVersion"`
                IsAbsoluteLatestVersion struct {
                        Text string `xml:",chardata"` // false
                        Type string `xml:"type,attr"`
                } `xml:"IsAbsoluteLatestVersion"`
                IsPrerelease struct {
                        Text string `xml:",chardata"` // false
                        Type string `xml:"type,attr"`
                } `xml:"IsPrerelease"`
                Published struct {
                        Text string `xml:",chardata"` // 2022-01-22T09:08:11.111Z
                        Type string `xml:"type,attr"`
                } `xml:"Published"`
                Language struct {
                        Text string `xml:",chardata"`
                        Null string `xml:"null,attr"`
                } `xml:"Language"`
                LicenseUrl           string `xml:"LicenseUrl"`           // http://www.7-zip.org/lice...
                PackageHash          string `xml:"PackageHash"`          // sspS9EaIX2/gtJ87jHitc5/Hq...
                PackageHashAlgorithm string `xml:"PackageHashAlgorithm"` // SHA512
                PackageSize          struct {
                        Text string `xml:",chardata"` // 3055
                        Type string `xml:"type,attr"`
                } `xml:"PackageSize"`
                ProjectUrl     string `xml:"ProjectUrl"` // http://www.7-zip.org/
                ReportAbuseUrl struct {
                        Text string `xml:",chardata"`
                        Null string `xml:"null,attr"`
                } `xml:"ReportAbuseUrl"`
                ReleaseNotes struct {
                        Text string `xml:",chardata"`
                        Null string `xml:"null,attr"`
                } `xml:"ReleaseNotes"`
                RequireLicenseAcceptance struct {
                        Text string `xml:",chardata"` // false
                        Type string `xml:"type,attr"`
                } `xml:"RequireLicenseAcceptance"`
                Tags                 string `xml:"Tags"`  // 7zip zip archiver admin
                Title                string `xml:"Title"` // 7-Zip
                VersionDownloadCount struct {
                        Text string `xml:",chardata"` // 0
                        Type string `xml:"type,attr"`
                } `xml:"VersionDownloadCount"`
        } `xml:"properties"`
}


func ParseMultipleNugetPackagesXmlData(xmlDataStr string) MultiplePackagesDetailsXmlStruct {
    xmlDataByteArr := []byte(xmlDataStr)
    var result MultiplePackagesDetailsXmlStruct
    xml.Unmarshal(xmlDataByteArr, &result)
    return result
}

func ParseSingleNugetPackagesXmlData(xmlDataStr string) SinglePackagesDetailsXmlStruct {
    xmlDataByteArr := []byte(xmlDataStr)
    var result SinglePackagesDetailsXmlStruct
    xml.Unmarshal(xmlDataByteArr, &result)
    return result
}
