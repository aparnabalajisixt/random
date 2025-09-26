package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"sort"
	"strconv"

	branchProto "github.com/sixt/com.sixt.service.branch-go-stubs/proto"
)

type BranchesWrapper struct {
	Branches []Branch `json:"branches"`
}

type Branch struct {
	BranchID    int       `json:"branchId"`
	Name        string    `json:"name"`
	BranchType  int       `json:"type"`
	IsCorporate *bool     `json:"isCorporate,omitempty"`
	Config      *Config   `json:"config,omitempty"`
	Addresses   []Address `json:"addresses"`
}

type Address struct {
	Country Country `json:"country"`
}

type Country struct {
	Iso2Code string `json:"iso2Code"`
}

type Config struct {
	IsAgencyBranch *bool `json:"isAgencyBranch,omitempty"`
}

type Result struct {
	BranchID    int
	Name        string
	Country     string
	BranchType  string
	IsCorporate bool
	IsAgency    bool
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
		var isCorporate, isAgency bool

		isCorporate = b.IsCorporate != nil && *b.IsCorporate
		isAgency = b.Config != nil && b.Config.IsAgencyBranch != nil && *b.Config.IsAgencyBranch

		branchType := branchProto.BranchType_name[int32(b.BranchType)]

		// Apply condition: (!isCorporate) OR (isAgency)
		if !isCorporate || isAgency {
			results = append(results, Result{BranchID: b.BranchID, Name: b.Name, Country: countryCode, BranchType: branchType, IsCorporate: isCorporate, IsAgency: isAgency})
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
	if err := writer.Write([]string{"Branch", "Name", "Country", "BranchType", "IsCorporate", "IsAgency"}); err != nil {
		panic(err)
	}
	for _, r := range results {
		if err := writer.Write([]string{strconv.Itoa(r.BranchID), r.Name, r.Country, r.BranchType, isCorporateToString(r.IsCorporate), isAgencyToString(r.IsAgency)}); err != nil {
			panic(err)
		}
	}
}

func isCorporateToString(isCorporate bool) string {
	if isCorporate {
		return "Corporate"
	}
	return "Franchise"
}

func isAgencyToString(isAgency bool) string {
	if isAgency {
		return "Agency"
	}
	return "NotAgency"
}
