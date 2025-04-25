package api

import (
	"fmt"
	"log"
)

// ExampleFetchGoogleSheetData demonstrates how to use the FetchGoogleSheetData function
func ExampleFetchGoogleSheetData() {
	// Fetch data from Google Sheet API
	response, err := FetchGoogleSheetData()
	if err != nil {
		log.Fatalf("Error fetching data from Google Sheet: %v", err)
	}

	// Check if the request was successful
	if response.Status != "success" {
		log.Println("API returned non-success status:", response.Status)
		return
	}

	// Print the number of records
	fmt.Printf("Fetched %d records from Google Sheet\n", len(response.Data))

	// Process each record
	for i, record := range response.Data {
		fmt.Printf("Record #%d:\n", i+1)
		fmt.Printf("  Company Name: %s\n", record.CompanyName)
		fmt.Printf("  Role: %s\n", record.Roll)
		fmt.Printf("  Employee Name: %s\n", record.EmployeeName)
		fmt.Printf("  Email: %s\n", record.Email)
		fmt.Printf("  Send At: %s\n", record.SendAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Send Status: %v\n", record.SendStatus)
		fmt.Println()
	}
}
