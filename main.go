package main

import (
    "fmt"
    "flag"
    "os"
    "path/filepath"
    "bufio"
    //"log"
    //"loggingPackage"
    "strings"
    "strconv"
    "net/http"
)

const (
    confFile = "run.conf"
    confFileNAErr = "Config file %s not found. Execution stopped.\n"
    hashSign = "#"

)

func main() {
    var testConfig string
    flag.StringVar(&testConfig, "config", "default", "Config name, default value is 'default'")
    flag.Parse()

    curPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
        // TODO
    }

    confFile := curPath + "/configs/" + testConfig + "/" + confFile
    fmt.Println(confFile)

    if fileExists(confFile) {
        runTests(confFile)
    } else {
        fmt.Printf(confFileNAErr, confFile)
    }
}

func runTests(confFile string) {    // TODO
    file, err := os.Open(confFile)
    if err != nil {
        //log.Fatal(err)
        // TODO
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
                runTest(strParts)
            } else {
                // log that we skip this line
            }
        }
    }

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

    intStatus, err := strconv.Atoi(statusCode)
    if err != nil {
        // TODO
    }

    fmt.Printf("Requesting %s, expecting status code %s\n", url, statusCode)
    resp, err := http.Head(strParts[0])
    if err != nil {
        // TODO
    }

    if resp.StatusCode == intStatus {
        fmt.Print("Success\n")
    } else {
        fmt.Printf("/!\\ Fail. Expected status code %s, got %d\n", statusCode, resp.StatusCode)
    }
    fmt.Println("---------------")
}
