package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Row struct {
	Timestamp float64
	Value     float64
}

func (r *Row) UnmarshalJSON(bs []byte) error {
	arr := []interface{}{}
	json.Unmarshal(bs, &arr)

	r.Timestamp = arr[0].(float64)
	val, err := strconv.ParseFloat(arr[1].(string), 64)
	if err != nil {
		panic("Couldnt unmarshall row" + string(bs))
	}
	r.Value = val
	return nil
}

type PromSeries struct {
	Metrics map[string]string `json:"metric"`
	Values  []Row
}

type PromData struct {
	ResultType string       `json:"resultType"`
	Result     []PromSeries `json:"result"`
}

type PromResponse struct {
	Status string   `json:"status"`
	Data   PromData `json:"data"`
}

// Example Response
// "status" : "success",
//  "data" : {
//     "resultType" : "matrix",
//     "result" : [
//        {
//           "metric" : {
//              "__name__" : "up",
//              "job" : "prometheus",
//              "instance" : "localhost:9090"
//           },
//           "values" : [
//              [ 1435781430.781, "1" ],
//              [ 1435781445.781, "1" ],
//              [ 1435781460.781, "1" ]
//           ]
//        },
//        {
//           "metric" : {
//              "__name__" : "up",
//              "job" : "node",
//              "instance" : "localhost:9091"
//           },
//           "values" : [
//              [ 1435781430.781, "0" ],
//              [ 1435781445.781, "0" ],
//              [ 1435781460.781, "1" ]
//           ]
//        }
//     ]
//  }

func main() {
	prom_endpoint := os.Getenv("PROM_ENDPOINT")
	if len(prom_endpoint) == 0 {
		panic("PROM_ENDPOINT env var not defined")
	}

	if len(os.Args) < 2 {
		panic("No query provided")
	}

	prom_query := os.Args[1]

	fmt.Printf("Query is %s\n", prom_query)

	duration, err := time.ParseDuration("-1h")
	if err != nil {
		panic("Error parsing duration")
	}

	startTimestamp := time.Now().Add(duration).Unix()
	endTimestamp := time.Now().Unix()
	step := "1m"
	request_uri := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=%s", prom_endpoint, prom_query, startTimestamp, endTimestamp, step)

	resp, err := http.Get(request_uri)
	if err != nil {
		fmt.Println(err.Error())
		panic("Error while fetching query result")
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("Cant read response body")
	}
	var promResponse PromResponse
	err = json.Unmarshal(bodyBytes, &promResponse)
	if err != nil {
		fmt.Println(err.Error())
		panic("Couldnt decode response body")
	}

	fmt.Printf("Response Status %s\n", promResponse.Status)
	fmt.Printf("Response Res type %s\n", promResponse.Data.ResultType)
	fmt.Printf("Value %f\n", promResponse.Data.Result[0].Values[0].Value)
	fmt.Printf("Timestamp %f\n", promResponse.Data.Result[0].Values[0].Timestamp)
}
