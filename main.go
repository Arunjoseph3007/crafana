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

func getTTYSize() (int64, int64, error) {
	return 142, 34 - 1, nil
}

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

	if len(os.Args) < 2 {
		panic("No query provided")
	}

	width, _, err := getTTYSize()
	if err != nil {
		panic(err.Error())
	}

	prom_query := os.Args[1]
	fmt.Printf("Query is %s\n", prom_query)

	duration, err := time.ParseDuration("-1h")
	if err != nil {
		panic("Error parsing duration")
	}

	startTimestamp := time.Now().Add(duration)
	endTimestamp := time.Now()
	// TODO: step logic is flawed have to fix later
	step := (endTimestamp.Unix() - startTimestamp.Unix()) / width
	step += 1

	request_uri := fmt.Sprintf(
		"%s/api/v1/query_range?query=%s&start=%d&end=%d&step=%d",
		prom_endpoint,
		prom_query,
		startTimestamp.Unix(),
		endTimestamp.Unix(),
		step,
	)

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

	if promResponse.Status != "success" {
		fmt.Println("Prom request failed")
		return
	}
	if promResponse.Data.ResultType != "matrix" {
		fmt.Println("Unrecognized result type" + promResponse.Data.ResultType)
		return
	}
}
