package main

import (
    "bytes"
    "encoding/json"
    "io/ioutil"
    "github.com/sirupsen/logrus"
    "net/http"
    "time"

    "github.com/kurtosis-tech/ava-test-controller/rpc/pchain"
)

const (
    // TODO TODO TODO Get this from serialization of testnet
    TEST_TARGET_URL="http://172.23.0.2:9650/"
    RPC_BODY= `{"jsonrpc": "2.0", "method": "platform.getCurrentValidators", "params":{},"id": 1}`
    // TODO TODO TODO Put retry configuration into sensible client object
    RETRIES=5
    RETRY_WAIT_SECONDS=5*time.Second
)

func main() {
    logrus.Infof("Test controller has been started.")
    var jsonStr = []byte(RPC_BODY)
    var jsonBuffer = bytes.NewBuffer(jsonStr)
    logrus.Infof("Test request as string: %s", jsonBuffer.String())

    var validatorList pchain.ValidatorList
    for i := 0; i < RETRIES; i++ {
        resp, err := http.Post(TEST_TARGET_URL + pchain.GetPChainEndpoint(), "application/json", jsonBuffer)
        if err != nil {
            logrus.Infof("Attempted connection...: %s", err.Error())
            logrus.Infof("Could not connect on attempt %d, retrying...", i+1)
            time.Sleep(RETRY_WAIT_SECONDS)
            continue
        }
        defer resp.Body.Close()

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            logrus.Fatalln(err)
        }

        var validatorResponse pchain.ValidatorResponse
        json.Unmarshal(body, &validatorResponse)

        validatorList = validatorResponse.Result["validators"]
        if len(validatorList) > 0 {
            logrus.Infof("Found validators!")
            break
        }
    }
    for _, validator := range validatorList {
        logrus.Infof("Validator id: %s", validator.Id)
    }
    if len(validatorList) < 1 {
        logrus.Infof("Failed to find a single validator.")
    }
}
