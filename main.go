package main

import (
    "bytes"
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "time"
)

// TODO TODO TODO Get this from serialization of testnet
const (
    TEST_TARGET_URL="http://172.23.0.2:9650/"
    PCHAIN_ENDPOINT="ext/P"
    RPC_BODY= `{"jsonrpc": "2.0", "method": "platform.getCurrentValidators", "params":{},"id": 1}`
    RETRIES=5
    RETRY_WAIT_SECONDS=5*time.Second
)

type Validator struct {
    StartTime string
    EndTime string
    StakeAmount string
    Id string
}

type ValidatorList []*Validator

type ValidatorResponse struct {
    Jsonrpc string
    Result map[string]ValidatorList
    Id int
}

func main() {
    println("Sup world")

    var jsonStr = []byte(RPC_BODY)
    var jsonBuffer = bytes.NewBuffer(jsonStr)
    log.Printf("Request as string: %s", jsonBuffer.String())

    var validatorList ValidatorList
    for i := 0; i < RETRIES; i++ {
        resp, err := http.Post(TEST_TARGET_URL + PCHAIN_ENDPOINT, "application/json", jsonBuffer)
        if err != nil {
            log.Printf("Attempted connection...: %s", err.Error())
            log.Printf("Could not connect on attempt %d, retrying...", i+1)
            time.Sleep(RETRY_WAIT_SECONDS)
            continue
        }
        defer resp.Body.Close()

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            log.Fatalln(err)
        }

        var validatorResponse ValidatorResponse
        json.Unmarshal(body, &validatorResponse)

        validatorList = validatorResponse.Result["validators"]
        if len(validatorList) > 0 {
            log.Printf("Found validators!")
            break
        }
    }
    for _, validator := range validatorList {
        log.Printf("Validator id: %s", validator.Id)
    }
    if len(validatorList) < 1 {
        log.Printf("Failed to find a single validator.")
    }
}
