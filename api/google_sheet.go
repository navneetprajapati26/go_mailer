package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GoogleSheetResponse represents the response structure from the Google Sheet API
type GoogleSheetResponse struct {
	Status string      `json:"status"`
	Data   []SheetData `json:"data"`
}

// SheetData represents each entry in the Google Sheet
type SheetData struct {
	CompanyName  string    `json:"CompanyName"`
	Roll         string    `json:"Roll"`
	EmployeeName string    `json:"EmployeeName"`
	Email        string    `json:"Email"`
	SendAt       time.Time `json:"SendAt"`
	SendStatus   bool      `json:"SendStatus"`
}

// FetchGoogleSheetData makes a request to the Google Sheet API and returns the parsed data
func FetchGoogleSheetData() (*GoogleSheetResponse, error) {
	// Google Sheet API URL
	apiURL := "https://script.google.com/macros/s/AKfycbywKRyWKtP2ryCXkPog-ycXN2_z6J8jDKrSTQjX9KCADUsCBTzRW_SMNFF6bNM1Dco9/exec"

	// Make GET request
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error making request to Google Sheet API: %w", err)
	}
	defer resp.Body.Close()

	// Check if response status code is OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Parse JSON response into struct
	var sheetResponse GoogleSheetResponse
	if err := json.Unmarshal(body, &sheetResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON response: %w", err)
	}

	return &sheetResponse, nil
}
