package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Structs for testing via JSON config
type configJSONEntity struct {
	Protocol      string                `json:"protocol"`
	Domain        string                `json:"domain"`
	CheckPageURLs bool                  `json:"checkUrls"`
	Urls          []configJSONURLEntity `json:"urls"`
}

type configJSONURLEntity struct {
	Url           string                    `json:"url"`
	StatusCode    int                       `json:"statusCode"`
	Elements      []configJSONElementEntity `json:"findElements"`
	SkipURLsCheck bool                      `json:"skipUrlsCheck"`
}

type configJSONElementEntity struct {
	Definition string `json:"def"`
	CountType  string `json:"countType"`
	Count      int    `json:"count"`
}

// Structs for testing via XML config
type configXMLEntity struct {
	XMLName xml.Name        `xml:"urlset"`
	Urls    []SitemapXmlUrl `xml:"url"`
}

type SitemapXmlUrl struct {
	Loc string `xml:"loc"`
}

// Config file data struct
type configData struct {
	// relative path, where config file is stored
	filePath string
	// file type, currently allowed types: "json" and "sitemapxml"
	fileType string
	// flag which allows outputting more data while tests are run
	verbose bool
}

// Struct `pageLinks` stores
// - urls found on a checked page in `URLs`
// - urls which has been already checked in `SourceURLs`
// - domain to add to urls from `URLs` property
type pageLinks struct {
	Domain     string
	URLs       map[string]map[string]int
	SourceURLs map[string]int
}

// Create new `pageLinks` instance
func newPageLinks() *pageLinks {
	pl := &pageLinks{
		URLs:       make(map[string]map[string]int),
		SourceURLs: make(map[string]int),
	}

	return pl
}

// Add new URL which should be checked
func (pl *pageLinks) addURL(key, value string) {
	if _, ok := pl.URLs[key]; !ok {
		pl.URLs[key] = map[string]int{}
	}

	if _, ok := pl.URLs[key][value]; !ok {
		pl.URLs[key][value] = 1
	}
}

// Add new URL to indicate that this URL has been checked already
func (pl *pageLinks) addSourceURL(key string) {
	if _, ok := pl.SourceURLs[key]; !ok {
		pl.SourceURLs[key] = 1
	}
}

// `savePageLinks` search for `a` tags inside a document and
// saves its' non-empty href attributes in a struct's field
func (pl *pageLinks) savePageLinks(doc *goquery.Document, pageURL string) {
	selection := doc.Find("a")
	selection.Each(func(i int, s *goquery.Selection) {
		if url, ok := s.Attr("href"); ok && "" != url && isLocalURL(url) {
			pl.addURL(url, pageURL)
		}
	})

	pl.addSourceURL(pageURL)
}

func (pl *pageLinks) checkPageLinks() {
	var fullURL string
	for k := range pl.URLs {
		fullURL = pl.Domain + k
		if _, ok := pl.SourceURLs[fullURL]; !ok {
			wg.Add(1)
			go func(URL string) {
				client := &http.Client{}
				req, _ := http.NewRequest(http.MethodHead, URL, nil)
				resp, err := client.Do(req)

				if err == nil {
					if resp.StatusCode == http.StatusOK {
						if cfData.verbose {
							notify(cn, fmt.Sprintf(statusCodeSuccessMsg, URL, http.StatusOK))
						}

					} else {
						// _ from range stores URLs where this url is met
						notify(cn, fmt.Sprintf(statusCodeFailureMsg, URL, http.StatusOK, resp.StatusCode))
					}
				} else {
					notify(cn, fmt.Sprintf(statusCodeSysFailMsg, URL))
				}

				wg.Done()
			}(fullURL)
		}
	}
}

// Fill configData fields // TODO needs proper testing
func (cd *configData) init() {
	var config, filename, filetype, verbose string
	flag.StringVar(&config, "config", "default", "Config name, default value is 'default'")
	flag.StringVar(&filename, "filename", "conf", "Config file name w/o extension, 'conf' by default")
	flag.StringVar(&filetype, "type", "json", "File type, current values 'json' and 'sitemapxml'")
	flag.StringVar(&verbose, "verbose", "y", "Do not show messages for success tests")
	flag.Parse()

	cd.fileType = "json"
	if filetype == "sitemapxml" {
		cd.fileType = "xml"
		filename = "sitemap"
	}

	cd.filePath = fmt.Sprintf("/configs/%s/%s.%s", config, filename, cd.fileType)
	cd.verbose = verbose == "y"
}

// Load finded file and instatiate required enitites
func (cd *configData) load() {
	curPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic("/!\\ Unable to get current exec path")
	}

	confFile := curPath + cd.filePath
	if fileExists(confFile) {
		file, err := os.Open(confFile)
		if err != nil {
			panic("/!\\ Error opening file " + confFile)
		}
		defer file.Close()

		switch cd.fileType {
		case typeJSON:
			var decoder = json.NewDecoder(file)
			err = decoder.Decode(&cJSONEntity)
			if err != nil {
				panic("/!\\ Err decoding json file: " + err.Error())
			}
			cJSONEntity.runTests()

		case typeXML:
			var decoder = xml.NewDecoder(file)
			err = decoder.Decode(&cXMLEntity)
			if err != nil {
				panic("/!\\ Err decoding json file: " + err.Error())
			}
			cXMLEntity.runTests()
		}
	} else {
		panic("/!\\ Config file " + confFile + " not found ")
	}
}

// Notification interface
type notifier interface {
	SendNotification(string)
}

type cliNotifier struct{}

// Realisation of interface method for `cliNotifier`
func (cn cliNotifier) SendNotification(msg string) {
	fmt.Print("---------------------------\n")
	log.Printf(msg)
	fmt.Print("---------------------------\n\n")
}

func notify(n notifier, msg string) {
	n.SendNotification(msg)
}

var cfData = configData{}
var cJSONEntity = configJSONEntity{}
var cXMLEntity = configXMLEntity{}
var wg sync.WaitGroup
var pl = newPageLinks()
var cn = cliNotifier{}

const (
	typeJSON = "json"
	typeXML  = "xml"

	sizeEq  = "eq"
	sizeGt  = "gt"
	sizeGte = "gte"
	sizeLt  = "lt"
	sizeLte = "lte"
	sizeNe  = "ne"

	prefixHTTP       = "http"
	prefixHTTPS      = "https"
	prefixNoProtocol = "//"
	prefixMailTo     = "mailto:"
	prefixSkypeTo    = "skype:"
	prefixTelTo      = "tel:"

	statusCodeSuccessMsg  = "Success. Requesting %s, expected status code %d confirmed\n"
	statusCodeFailureMsg  = "/!\\ Fail. Requesting %s, expected status code %d, got %d\n"
	statusCodeSysFailMsg  = "/!\\ SYSTEMFAIL. Error performing http-request to %s\n"
	selectorTestUnsuppMsg = "/!\\ Fail. Not supported elCouType %s\n"
)

// Available command line options
// -config = name of config, which is path in app's /config path, required
// -type = type of cofig file, currently allowed types: "json" and "sitemapxml", "json" is default, optional
// -filename = custom name of config file, by default it's "conf", optional
// -verbose = output more data when tests are run, "n" by default, optional
//
// Example run: ./main -config=some_site -type=sitemapxml -verbose=n -filename=load
func main() {
	cfData.init()
	cfData.load()

	wg.Wait()

	pl.checkPageLinks()
	wg.Wait()

	log.Printf("All tests for config '%s' completed", cfData.filePath)
}

// Run tests for each url in a config
func (ce *configJSONEntity) runTests() {
	fullDomain := fmt.Sprintf("%s://%s", ce.Protocol, ce.Domain)
	pl.Domain = fullDomain
	for _, v := range ce.Urls {
		if "" != v.Url {
			wg.Add(1)
			go func(i configJSONURLEntity) {
				i.runTest(fullDomain, ce.CheckPageURLs)
			}(v)
		}
	}
}

// `runTest` performs test for a single url from config
func (cue *configJSONURLEntity) runTest(domain string, checkURLs bool) {
	fullURL := domain + cue.Url
	reqType := "HEAD"
	checkPageLinks := checkURLs && !cue.SkipURLsCheck
	if 0 < len(cue.Elements) || checkPageLinks {
		reqType = "GET"
	}

	client := &http.Client{}
	req, _ := http.NewRequest(reqType, fullURL, nil)
	resp, err := client.Do(req)

	if err == nil {
		if resp.StatusCode == cue.StatusCode {
			if cfData.verbose {
				notify(cn, fmt.Sprintf(statusCodeSuccessMsg, fullURL, cue.StatusCode))
			}

			if reqType == "GET" {
				doc, err := goquery.NewDocumentFromResponse(resp)
				if err == nil {
					if 0 < len(cue.Elements) {
						for _, v := range cue.Elements {
							v.testElement(doc)
						}
					}

					if checkPageLinks {
						pl.savePageLinks(doc, fullURL)
					}
				} else {
					notify(cn, fmt.Sprintf("/!\\ SYSTEMFAIL. Error reading http-request body from %s\n", fullURL))
				}
			}
		} else {
			notify(cn, fmt.Sprintf(statusCodeFailureMsg, fullURL, cue.StatusCode, resp.StatusCode))
		}
	} else {
		notify(cn, fmt.Sprintf(statusCodeSysFailMsg, fullURL))
	}

	wg.Done()
}

// `testElement` tests if size of a selecteor mathes condition
func (cee *configJSONElementEntity) testElement(doc *goquery.Document) {
	elDefinition := strings.TrimSpace(cee.Definition)
	elCouType := strings.TrimSpace(cee.CountType)
	selection := doc.Find(elDefinition)
	msg := fmt.Sprintf(selectorTestUnsuppMsg, elCouType)

	selLen := selection.Length()
	switch elCouType {
	case sizeEq:
		if selLen == cee.Count {
			msg = ""
			if cfData.verbose {
				msg = getSelectorTestSuccMsg(elDefinition, sizeEq, cee.Count)
			}
		} else {
			msg = getSelectorTestFailMsg(elDefinition, sizeEq, cee.Count, selLen)
		}

	case sizeGt:
		if selLen > cee.Count {
			msg = ""
			if cfData.verbose {
				msg = getSelectorTestSuccMsg(elDefinition, sizeGt, cee.Count)
			}
		} else {
			msg = getSelectorTestFailMsg(elDefinition, sizeGt, cee.Count, selLen)
		}

	case sizeGte:
		if selLen >= cee.Count {
			msg = ""
			if cfData.verbose {
				msg = getSelectorTestSuccMsg(elDefinition, sizeGte, cee.Count)
			}
		} else {
			msg = getSelectorTestFailMsg(elDefinition, sizeGte, cee.Count, selLen)
		}

	case sizeLt:
		if selLen < cee.Count {
			msg = ""
			if cfData.verbose {
				msg = getSelectorTestSuccMsg(elDefinition, sizeLt, cee.Count)
			}
		} else {
			msg = getSelectorTestFailMsg(elDefinition, sizeLt, cee.Count, selLen)
		}

	case sizeLte:
		if selLen <= cee.Count {
			msg = ""
			if cfData.verbose {
				msg = getSelectorTestSuccMsg(elDefinition, sizeLte, cee.Count)
			}
		} else {
			msg = getSelectorTestFailMsg(elDefinition, sizeLte, cee.Count, selLen)
		}

	case sizeNe:
		if selLen != cee.Count {
			msg = ""
			if cfData.verbose {
				msg = getSelectorTestSuccMsg(elDefinition, sizeNe, cee.Count)
			}
		} else {
			msg = getSelectorTestFailMsg(elDefinition, sizeNe, cee.Count, selLen)
		}
	}

	if "" != msg {
		notify(cn, msg)
	}
}

func (ce *configXMLEntity) runTests() {
	for _, v := range ce.Urls {
		if "" != v.Loc {
			wg.Add(1)
			go func(i SitemapXmlUrl) {
				i.runTest()
			}(v)
		}
	}
}

func (sxu *SitemapXmlUrl) runTest() {
	reqType := "HEAD"
	client := &http.Client{}
	req, _ := http.NewRequest(reqType, sxu.Loc, nil)
	resp, err := client.Do(req)
	/* consider that all urls in sitemap shoud give 200 status code */
	requiredStatusCode := 200

	if err == nil {
		if resp.StatusCode == requiredStatusCode {
			if cfData.verbose {
				notify(cn, fmt.Sprintf(statusCodeSuccessMsg, sxu.Loc, requiredStatusCode))
			}
		} else {
			notify(cn, fmt.Sprintf(statusCodeFailureMsg, sxu.Loc, requiredStatusCode, resp.StatusCode))
		}
	} else {
		notify(cn, fmt.Sprintf(statusCodeSysFailMsg, sxu.Loc))
	}

	wg.Done()
}

func getSelectorTestSuccMsg(elDef string, couType string, cou int) string {
	return fmt.Sprintf(
		"Success. Selector: '%s'. Expected size '%s %d' confirmed\n",
		elDef,
		couType,
		cou,
	)
}

func getSelectorTestFailMsg(elDef string, couType string, exCou int, realCou int) string {
	return fmt.Sprintf(
		"/!\\ Fail. Selector: '%s'. Expected size '%s %d', received size %d\n",
		elDef,
		couType,
		exCou,
		realCou,
	)
}

func fileExists(fileName string) bool {
	result := true

	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			result = false
		}
	}

	return result
}

// Every url which not starts with `http` or `https` or `//` or `mailto:` or `tel:` is considered local
func isLocalURL(url string) bool {
	return !strings.HasPrefix(url, prefixHTTP) &&
		!strings.HasPrefix(url, prefixHTTPS) &&
		!strings.HasPrefix(url, prefixNoProtocol) &&
		!strings.HasPrefix(url, prefixMailTo) &&
		!strings.HasPrefix(url, prefixSkypeTo) &&
		!strings.HasPrefix(url, prefixTelTo)
}
