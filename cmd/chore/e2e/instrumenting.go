package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cortexproject/cortex/pkg/ingester/client"
	"gopkg.in/raintank/schema.v1"
)

func graphite(orbID, cloudURL, cloudKey string, test func(orbconfig string) error) func(orbconfig string) error {

	send := func(value float64) {
		if err := sendGraphiteStatus(orbID, cloudURL, cloudKey, value); err != nil {
			panic(err)
		}
	}

	return func(orbconfig string) error {
		send(0.5)
		err := test(orbconfig)
		var value float64
		if err == nil {
			value = 1
		}
		send(value)
		return err
	}
}

func cort(orbID, cloudURL, cloudKey string) {
	client.MakeIngesterClient()
}

func sendGraphiteStatus(orbID string, cloudURL, cloudKey string, value float64) error {
	name := "my.test.metric"
	metrics := schema.MetricDataArray{&schema.MetricData{
		Name:     name,
		Metric:   name,
		Interval: 24 * 60 * 60,
		Value:    value,
		Unit:     "",
		Time:     time.Now().Unix(),
		Mtype:    "gauge",
		Tags:     []string{fmt.Sprintf("orb=%s", orbID)},
	}}

	// encode as json
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", cloudURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+cloudKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	buf := make([]byte, 4096)
	n, err := resp.Body.Read(buf)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("sending metric to graphana cloud graphite api at %s failed with status %s and response %s", cloudURL, resp.Status, string(buf[:n]))
	}
	fmt.Println("Value", value, "sent to grafana cloud graphite api")
	return nil
}
