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
type configJsonEntity struct {
	Protocol string                `json:"protocol"`
	Domain   string                `json:"domain"`
	Urls     []configJsonUrlEntity `json:"urls"`
}

type configJsonUrlEntity struct {
	Url        string                    `json:"url"`
	StatusCode int                       `json:"statusCode"`
	Elements   []configJsonElementEntity `json:"findElements"`
}

type configJsonElementEntity struct {
	Definition string `json:"def"`
	CountType  string `json:"countType"`
	Count      int    `json:"count"`
}

// Structs for testing via XML config
type configXmlEntity struct {
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
			err = decoder.Decode(&cJsonEntity)
			if err != nil {
				panic("/!\\ Err decoding json file: " + err.Error())
			}
			cJsonEntity.runTests()

		case typeXML:
			var decoder = xml.NewDecoder(file)
			err = decoder.Decode(&cXmlEntity)
			if err != nil {
				panic("/!\\ Err decoding json file: " + err.Error())
			}
			cXmlEntity.runTests()
		}
	} else {
		panic("/!\\ Config file " + confFile + " not found ")
	}
}

var cfData = new(configData)
var cJsonEntity = new(configJsonEntity)
var cXmlEntity = new(configXmlEntity)
var wg sync.WaitGroup

const (
	typeJSON = "json"
	typeXML  = "xml"

	EQ  = "eq"
	GT  = "gt"
	GTE = "gte"
	LT  = "lt"
	LTE = "lte"
	NE  = "ne"

	statusCodeSuccessMsg = "Success. Requesting %s, expected status code %d confirmed\n"
	statusCodeFailureMsg = "/!\\ Fail. Requesting %s, expected status code %d, got %d\n"
	statusCodeSysFailMsg = "/!\\ SYSTEMFAIL. Error performing http-request to %s\n"
)

/*
Available command line options
-config = name of config, which is path in app's /config path, required
-type = type of cofig file, currently allowed types: "json" and "sitemapxml", "json" is default, optional
-filename = custom name of config file, by default it's "conf", optional
-verbose = output more data when tests are run, "n" by default, optional

Example run: ./main -config=some_site -type=sitemapxml -verbose=n -filename=load
*/
func main() {
	cfData.init()
	cfData.load()

	wg.Wait()
	log.Printf("All tests for config '%s' completed", cfData.filePath)
}

func (ce *configJsonEntity) runTests() {
	fullDomain := fmt.Sprintf("%s://%s", ce.Protocol, ce.Domain)
	for _, v := range ce.Urls {
		if "" != v.Url {
			wg.Add(1)
			go v.runTest(fullDomain)
		}
	}
}

func (cue *configJsonUrlEntity) runTest(domain string) {
	fullUrl := domain + cue.Url
	reqType := "HEAD"
	if 0 < len(cue.Elements) {
		reqType = "GET"
	}

	client := &http.Client{}
	req, _ := http.NewRequest(reqType, fullUrl, nil)
	resp, err := client.Do(req)

	if err == nil {
		if resp.StatusCode == cue.StatusCode {
			if cfData.verbose {
				printMsg(fmt.Sprintf(statusCodeSuccessMsg, fullUrl, cue.StatusCode))
			}

			if 0 < len(cue.Elements) {
				doc, err := goquery.NewDocumentFromResponse(resp)
				if err == nil {
					for _, v := range cue.Elements {
						v.testElement(doc)
					}
				} else {
					printMsg(fmt.Sprintf("/!\\ SYSTEMFAIL. Error reading http-request body from %s\n", fullUrl))
				}
			}
		} else {
			printMsg(fmt.Sprintf(statusCodeFailureMsg, fullUrl, cue.StatusCode, resp.StatusCode))
		}
	} else {
		printMsg(fmt.Sprintf(statusCodeSysFailMsg, fullUrl))
	}

	wg.Done()
}

func (cee *configJsonElementEntity) testElement(doc *goquery.Document) {
	elDefinition := strings.TrimSpace(cee.Definition)
	elCouType := strings.TrimSpace(cee.CountType)
	selection := doc.Find(elDefinition)
	msg := fmt.Sprintf("/!\\ Fail. Not supported elCouType %s\n", elCouType)

	selLen := selection.Length()
	switch elCouType {
	case EQ:
		if selLen == cee.Count {
			if cfData.verbose {
				msg = getSelectorTestSuccMsg(elDefinition, EQ, cee.Count)
			} else {
				msg = ""
			}
		} else {
			msg = getSelectorTestFailMsg(elDefinition, EQ, cee.Count, selLen)
		}

	case GT:
		if selLen > cee.Count {
			if cfData.verbose {
				msg = getSelectorTestSuccMsg(elDefinition, GT, cee.Count)
			} else {
				msg = ""
			}
		} else {
			msg = getSelectorTestFailMsg(elDefinition, GT, cee.Count, selLen)
		}

	case GTE:
		if selLen >= cee.Count {
			if cfData.verbose {
				msg = getSelectorTestSuccMsg(elDefinition, GTE, cee.Count)
			} else {
				msg = ""
			}
		} else {
			msg = getSelectorTestFailMsg(elDefinition, GTE, cee.Count, selLen)
		}

	case LT:
		if selLen < cee.Count {
			if cfData.verbose {
				msg = getSelectorTestSuccMsg(elDefinition, LT, cee.Count)
			} else {
				msg = ""
			}
		} else {
			msg = getSelectorTestFailMsg(elDefinition, LT, cee.Count, selLen)
		}

	case LTE:
		if selLen <= cee.Count {
			if cfData.verbose {
				msg = getSelectorTestSuccMsg(elDefinition, LTE, cee.Count)
			} else {
				msg = ""
			}
		} else {
			msg = getSelectorTestFailMsg(elDefinition, LTE, cee.Count, selLen)
		}

	case NE:
		if selLen != cee.Count {
			if cfData.verbose {
				msg = getSelectorTestSuccMsg(elDefinition, NE, cee.Count)
			} else {
				msg = ""
			}
		} else {
			msg = getSelectorTestFailMsg(elDefinition, NE, cee.Count, selLen)
		}
	}

	if "" != msg {
		printMsg(msg)
	}
}

func (ce *configXmlEntity) runTests() {
	for _, v := range ce.Urls {
		if "" != v.Loc {
			wg.Add(1)
			go v.runTest()
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
				printMsg(fmt.Sprintf(statusCodeSuccessMsg, sxu.Loc, requiredStatusCode))
			}
		} else {
			printMsg(fmt.Sprintf(statusCodeFailureMsg, sxu.Loc, requiredStatusCode, resp.StatusCode))
		}
	} else {
		printMsg(fmt.Sprintf(statusCodeSysFailMsg, sxu.Loc))
	}

	wg.Done()
}

func printMsg(msg string) {
	prLine(false)
	log.Printf(msg)
	prLine(true)
}

func prLine(doubleNl bool) {
	var nl = "\n"
	if doubleNl {
		nl = "\n\n"
	}
	fmt.Print("---------------------------" + nl)
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
