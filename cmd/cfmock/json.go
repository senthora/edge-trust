package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const jsonPath = "/usr/share/nginx/html/ips.json"

func loadResponse() IPDataResponse {
	fmt.Println("loading mock Cloudflare response")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("ips.json does not exist, initializing new response")
			return IPDataResponse{}
		}
		panic(fmt.Errorf("read ips.json: %w", err))
	}
	var response IPDataResponse

	fmt.Println("parsing ips.json")

	if err := json.Unmarshal(data, &response); err != nil {
		panic(fmt.Errorf("parse ips.json: %w", err))
	}
	return response
}

func saveResponse(response IPDataResponse) {
	fmt.Println("writing updated mock response")

	updated, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		panic(fmt.Errorf("marshal updated json: %w", err))
	}
	if err := os.WriteFile(jsonPath, updated, 0644); err != nil {
		panic(fmt.Errorf("write ips.json: %w", err))
	}
}

func deleteResponseFile() {
	fmt.Println("deleting ips.json")

	if err := os.Remove(jsonPath); err != nil {
		panic(fmt.Errorf("delete ips.json: %w", err))
	}
}
