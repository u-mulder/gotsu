package main

import (
    "fmt"
    "flag"
    "os"
    "path/filepath"
    "bufio"
    "log"
    //"loggingPackage"
    "strings"
    "strconv"
    "net/http"
    //"io/ioutil"
)

const (
    confFile = "run.conf"
    confFileNAErr = "Config file %s not found. Execution stopped.\n"
    hashSign = "#"
)

type configData struct {
    name string
    verbose bool
}

var cfData = new(configData)

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
        runTests(confFile)
    } else {
        panic("/!\\ Config file " + confFile + " not found ")
    }
}

func runTests(confFile string) {
    file, err := os.Open(confFile)
    if err != nil {
        panic("/!\\ Error opening file " + confFile)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var curStr string
    strParts := []string{}

    for scanner.Scan() {
        curStr = scanner.Text()
        if notCommentLine(curStr) {
            strParts = strings.Split(curStr, "->")
            if len(strParts) == 2 {
                runTest(strParts)   // call with go - breaks everything o_O
            } else {
                log.Printf("Conf string '%s' is unparsable", curStr)
            }
        }
    }

    fmt.Println("Tests completed")

    /*if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }*/
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

func notCommentLine(str string) bool {
    return hashSign != str[:1]
}

func runTest(strParts []string) {
    url := strings.TrimSpace(strParts[0])
    statusCode := strings.TrimSpace(strParts[1])

    if intStatus, err := strconv.Atoi(statusCode); err == nil {
        client := &http.Client{}
        req, _ := http.NewRequest("HEAD", url, nil)
        resp, _ := client.Do(req)
        if resp.StatusCode == intStatus {
            if cfData.verbose {
                prLine(false)
                log.Printf("Requesting %s, expecting status code %s\n", url, statusCode)
                log.Print("Success\n")
                prLine(true)
            }
        } else {
            prLine(false)
            log.Printf("Requesting %s, expecting status code %s\n", url, statusCode)
            log.Printf("/!\\ Fail. Expected status code %s, got %d\n", statusCode, resp.StatusCode)
            prLine(true)
        }
    } else {
        prLine(false)
        log.Printf("Code '%s' could not be converted to int, request skipped", statusCode)
        prLine(true)
    }
}

func prLine(doubleNl bool) {
    var nl = "\n"
    if doubleNl {
        nl = "\n\n"
    }
    fmt.Print("---------------" + nl)
}

