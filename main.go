package main

import (
    "fmt"
    "flag"
    "os"
    "path/filepath"
    "log"
    "net/http"
    "encoding/json"
    "github.com/PuerkitoBio/goquery"
    "strings"
)

const (
    confFile = "run.conf"
    confFileNAErr = "Config file %s not found. Execution stopped.\n"
    hashSign = "#"
)

const (
    EQ = "eq"
    GT = "gt"
    GTE = "gte"
    LT = "lt"
    LTE = "lte"
    NE = "ne"
)

type configData struct {
    name string
    verbose bool
}

var cfData = new(configData)

type configEntity struct {
    Protocol string         `json:"protocol"`
    Domain string           `json:"domain"`
    Urls []configUrlEntity  `json:"urls"`
}

type configUrlEntity struct {
    Url string                      `json:"url"`
    StatusCode int                  `json:"statusCode"`
    Elements []configElementEntity  `json:"findElements"`

}

type configElementEntity struct {
    Definition string         `json:"def"`
    CountType string         `json:"countType"`
    Count int                `json:"count"`
}

var cEntity = new(configEntity)

func main() {
    // -config=confName
    var verbose string
    flag.StringVar(&cfData.name, "config", "default", "Config name, default value is 'default'")
    flag.StringVar(&verbose, "verbose", "y", "Do not show messages for success tests")
    flag.Parse()

    cfData.verbose = verbose == "y"

    curPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
        panic("/!\\ Unable to get current exec path")
    }

    confFile := curPath + "/configs/" + cfData.name + "/" + confFile
    if fileExists(confFile) {
        file, err := os.Open(confFile)
        if err != nil {
            panic("/!\\ Error opening file " + confFile)
        }
        defer file.Close()

        var decoder = json.NewDecoder(file)
        err = decoder.Decode(&cEntity)
        if err != nil {
            panic("/!\\ Err decoding json file: " + err.Error())
        }

        cEntity.runTests()

    } else {
        panic("/!\\ Config file " + confFile + " not found ")
    }
}


func (ce *configEntity) runTests() {
    fullDomain := fmt.Sprintf("%s://%s", ce.Protocol,  ce.Domain)
    for _, v := range ce.Urls {
        if "" != v.Url {
            v.runTest(fullDomain)
        }
    }

    log.Printf("All tests for config '%s' completed", cfData.name)
}


func (cue *configUrlEntity) runTest(domain string) {
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
                printMsg(fmt.Sprintf("Success. Requesting %s, expected status code %d confirmed\n", fullUrl, cue.StatusCode))
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
            printMsg(fmt.Sprintf("/!\\ Fail. Requesting %s, expected status code %d, got %d\n", fullUrl, cue.StatusCode, resp.StatusCode))

        }
    } else {
        printMsg(fmt.Sprintf("/!\\ SYSTEMFAIL. Error performing http-request to %s\n", fullUrl))
    }
}


func (cee *configElementEntity) testElement (doc *goquery.Document) {
    elDefinition := strings.TrimSpace(cee.Definition)
    elCouType := strings.TrimSpace(cee.CountType)
    selection := doc.Find( elDefinition )
    msg := fmt.Sprintf("/!\\ Fail. Not supported elCouType %s\n", elCouType)

    selLen := selection.Length()
    switch elCouType {
        case EQ:
            if selLen == cee.Count {
                if cfData.verbose {
                    msg = getSelectorTestSuccMsg(elDefinition, EQ, cee.Count)
                } else {
                    msg = "";
                }
            } else {
                msg = getSelectorTestFailMsg(elDefinition, EQ, cee.Count, selLen)
            }

        case GT:
            if selLen > cee.Count {
                if cfData.verbose {
                    msg = getSelectorTestSuccMsg(elDefinition, GT, cee.Count)
                } else {
                    msg = "";
                }
            } else {
                msg = getSelectorTestFailMsg(elDefinition, GT, cee.Count, selLen)
            }

        case GTE:
            if selLen >= cee.Count {
                if cfData.verbose {
                    msg = getSelectorTestSuccMsg(elDefinition, GTE,  cee.Count)
                } else {
                    msg = "";
                }
            } else {
                msg = getSelectorTestFailMsg(elDefinition, GTE, cee.Count, selLen)
            }

        case LT:
            if selLen < cee.Count {
                if cfData.verbose {
                    msg = getSelectorTestSuccMsg(elDefinition, LT, cee.Count)
                } else {
                    msg = "";
                }
            } else {
                msg = getSelectorTestFailMsg(elDefinition, LT, cee.Count, selLen)
            }

        case LTE:
            if selLen <= cee.Count {
                if cfData.verbose {
                    msg = getSelectorTestSuccMsg(elDefinition, LTE, cee.Count)
                } else {
                    msg = "";
                }
            } else {
                msg = getSelectorTestFailMsg(elDefinition, LTE, cee.Count, selLen)
            }

        case NE:
            if selLen != cee.Count {
                if cfData.verbose {
                    msg = getSelectorTestSuccMsg(elDefinition, NE, cee.Count)
                } else {
                    msg = "";
                }
            } else {
                msg = getSelectorTestFailMsg(elDefinition, NE, cee.Count, selLen)
            }
    }

    if "" != msg {
        printMsg(msg)
    }
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
    return  fmt.Sprintf(
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
