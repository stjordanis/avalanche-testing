package main

import (
    "bytes"
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
)

// TODO TODO TODO Get this from serialization of testnet
const (
    TEST_TARGET_URL="http://127.0.0.1:9651/"
    PCHAIN_ENDPOINT="ext/P"
    RPC_BODY= `{"jsonrpc": "2.0", "method": "platform.getCurrentValidators", "params":{},"id": 1}`
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

    resp, err := http.Post(TEST_TARGET_URL + PCHAIN_ENDPOINT, "application/json", jsonBuffer)
    if err != nil {
        log.Fatalln(err)
    }

    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatalln(err)
    }

    var validatorResponse ValidatorResponse
    json.Unmarshal(body, &validatorResponse)

    validatorList := validatorResponse.Result["validators"]

    for _, validator := range validatorList {
        log.Printf("Validator id: %s", validator.Id)
    }

    if len(validatorList) < 1 {
        panic("Failed to find a single validator.")
    }
}
