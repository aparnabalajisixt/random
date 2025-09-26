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
	Name        string    `json:"name"`
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
	Name     string
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

	// Read corporate countries from CSV
	corporate := make(map[string]struct{})
	corporateFile, err := os.Open("corporate.csv")
	if err != nil {
		panic(err)
	}
	defer corporateFile.Close()
	reader := csv.NewReader(corporateFile)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}
	for i, rec := range records {
		if i == 0 {
			continue // skip header
		}
		if len(rec) >= 2 {
			corporate[rec[1]] = struct{}{}
		}
	}

	var results []Result
	for _, b := range wrapper.Branches {
		countryCode := ""
		if len(b.Addresses) > 0 {
			countryCode = b.Addresses[0].Country.Iso2Code
		}

		// Skip if not a corporate country
		if _, corporateCountry := corporate[countryCode]; !corporateCountry {
			continue
		}

		isCorporate := b.IsCorporate != nil && *b.IsCorporate
		isAgency := b.IsAgency != nil && *b.IsAgency

		// Apply condition: (!isCorporate) OR (isAgency)
		if !isCorporate || isAgency {
			results = append(results, Result{BranchID: b.BranchID, Name: b.Name, Country: countryCode})
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
	if err := writer.Write([]string{"Branch", "Name", "Country"}); err != nil {
		panic(err)
	}
	for _, r := range results {
		if err := writer.Write([]string{strconv.Itoa(r.BranchID), r.Name, r.Country}); err != nil {
			panic(err)
		}
	}
}
