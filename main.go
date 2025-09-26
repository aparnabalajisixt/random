package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"sort"
	"strconv"
)

type BranchesWrapper struct {
	Branches []Branch `json:"branches"`
}

type Branch struct {
	BranchID    int       `json:"branchId"`
	IsCorporate *bool     `json:"isCorporate,omitempty"`
	IsAgency    *bool     `json:"isAgency,omitempty"`
	Addresses   []Address `json:"addresses"`
}

type Address struct {
	Country Country `json:"country"`
}

type Country struct {
	Iso2Code string `json:"iso2Code"`
}

type Result struct {
	BranchID int
	Country  string
}

func main() {
	// Read JSON file
	data, err := os.ReadFile("branches.json")
	if err != nil {
		panic(err)
	}

	// Unmarshal
	var wrapper BranchesWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		panic(err)
	}

	var results []Result
	for _, b := range wrapper.Branches {
		isCorporate := b.IsCorporate != nil && *b.IsCorporate
		isAgency := b.IsAgency != nil && *b.IsAgency

		// Apply condition: (!isCorporate) OR (isAgency)
		if !isCorporate || isAgency {
			countryCode := ""
			if len(b.Addresses) > 0 {
				countryCode = b.Addresses[0].Country.Iso2Code
			}
			results = append(results, Result{BranchID: b.BranchID, Country: countryCode})
		}
	}

	// Sort by country
	sort.Slice(results, func(i, j int) bool {
		if results[i].Country == results[j].Country {
			return results[i].BranchID < results[j].BranchID
		}
		return results[i].Country < results[j].Country
	})

	// Write results to CSV file
	outputFile, err := os.Create("results.csv")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Header
	if err := writer.Write([]string{"Branch", "Country"}); err != nil {
		panic(err)
	}
	for _, r := range results {
		if err := writer.Write([]string{strconv.Itoa(r.BranchID), r.Country}); err != nil {
			panic(err)
		}
	}
}
