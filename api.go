package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

func getJson(url string, target interface{}) error {
	log.Printf("getting url: %s", url)
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	log.Printf("body: %s", body)

	return json.Unmarshal(body, target)
}
