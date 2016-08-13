package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

// A structure to capture the only things we care about from a response from
// https://clever.com/developers/docs/explorer#endpoint_sections_sections
type CleverResponse struct {
	Data []struct {
		Data struct {
			Students []string
		}
	}
	Paging struct {
		Current int
		Total   int
		Count   int
	}
	Links []struct {
		Rel string
		Uri string
	}
}

// getFromClever takes a path to append to the Clever base URI and makes a request.
// It then decodes the response into a CleverResponse.
func getFromClever(path string) (CleverResponse, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.clever.com%v", path), nil)
	req.Header.Set("Authorization", "Bearer DEMO_TOKEN")
	client := http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	response := CleverResponse{}

	err := json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Printf("Error decoding json %v\n", err)
	}
	return response, nil
}

// getStudentAverage repeatedly calls the Clever sections API, fetching batchSize
// sections at a time and using the returned "links" structure to walk forward.
func getStudentAverage(batchSize int) (float64, error) {
	if batchSize == 0 {
		batchSize = 100
	}

	cleverResponse, err := getFromClever(fmt.Sprintf("/v1.1/sections?limit=%v", batchSize))
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Error retrieving initial clever response: %v\n", err))
	}

	// Grab number of total sections from paging blob.
	// Why does this "paging" structure disappear when "starting_after" is specified???
	// :(
	// You weren't there when I needed you, Paging struct.
	numSections := cleverResponse.Paging.Count
	numStudents := 0

	for {
		for _, sec := range cleverResponse.Data {
			numStudents += len(sec.Data.Students)
		}

		nextUri := ""
		for _, link := range cleverResponse.Links {
			if link.Rel == "next" {
				nextUri = link.Uri
				break
			}
		}

		if nextUri == "" {
			break
		}

		cleverResponse, err = getFromClever(nextUri)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("Getting from clever failed with msg %v", err))
		}
	}

	return float64(numStudents) / float64(numSections), nil
}

func main() {
	average, err := getStudentAverage(100)
	if err != nil {
		log.Fatalf("Error getting student per section average: %v\n", err)
	}

	fmt.Printf("Average # students per section: %v \n", average)
}
