package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// UpdateSendStatus updates the send status for an email in the Google Sheet
func UpdateSendStatus(email string, sendStatus bool) error {
	// Base API URL
	baseURL := "https://script.google.com/macros/s/AKfycbywKRyWKtP2ryCXkPog-ycXN2_z6J8jDKrSTQjX9KCADUsCBTzRW_SMNFF6bNM1Dco9/exec"

	// Build URL with query parameters
	params := url.Values{}
	params.Add("action", "update")
	params.Add("email", email)
	params.Add("sendStatus", fmt.Sprintf("%t", sendStatus))

	updateURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Make GET request to update the status
	resp, err := http.Get(updateURL)
	if err != nil {
		return fmt.Errorf("error making request to update send status: %w", err)
	}
	defer resp.Body.Close()

	// Check if response status code is OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response status when updating send status: %s", resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body for update: %w", err)
	}

	// Parse JSON response to check status
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("error parsing JSON response for update: %w", err)
	}

	// Check if update was successful
	status, ok := response["status"].(string)
	if !ok || status != "success" {
		return fmt.Errorf("update was not successful, response: %s", string(body))
	}

	return nil
}
