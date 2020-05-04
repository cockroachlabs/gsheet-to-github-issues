package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

type cell string

type headerRow struct {
	headers []string
	idx     map[string]int
}

func (h *headerRow) mustGetLetter(header string) string {
	idx := h.idx[header]
	letters := ""
	for idx > 0 {
		letters += string('A' + (idx % 26))
		idx /= 26
	}
	return letters
}

type sheetRow struct {
	values  []cell
	headers *headerRow
}

func (r *sheetRow) toMap() map[string]cell {
	ret := make(map[string]cell, len(r.values))
	for idx, h := range r.headers.headers {
		ret[h] = r.values[idx]
	}
	return ret
}

func (r *sheetRow) get(h string) (cell, bool) {
	if _, ok := r.headers.idx[h]; !ok {
		return "", false
	}
	return r.values[r.headers.idx[h]], true
}

func (r *sheetRow) mustGet(h string) string {
	c, ok := r.get(h)
	if !ok {
		panic(fmt.Errorf("cannot find header: %s", h))
	}
	return string(c)
}

// headersFromValues returns the headers from a google sheet.
func headersFromSheet(values *sheets.ValueRange) (*headerRow, error) {
	if len(values.Values) == 0 {
		return nil, fmt.Errorf("expected at least one row in range")
	}
	ret := &headerRow{
		idx: map[string]int{},
	}
	row := values.Values[0]
	for idx, header := range row {
		h, ok := header.(string)
		if !ok {
			return nil, fmt.Errorf("expected header row at idx %d to be string, found %T: %v", idx, header, header)
		}
		ret.headers = append(ret.headers, h)
		ret.idx[h] = idx
	}
	return ret, nil
}

// sheetToRows converts values from a sheet range to an interface map of them.
func sheetToRows(headers *headerRow, values *sheets.ValueRange) ([]sheetRow, error) {
	ret := make([]sheetRow, 0, len(values.Values)-1)
	for valueRowIdx, valueRow := range values.Values[1:] {
		row := sheetRow{
			values:  make([]cell, len(headers.headers)),
			headers: headers,
		}
		for i := 0; i < len(headers.headers); i++ {
			switch c := valueRow[i].(type) {
			case bool:
				row.values[i] = cell(fmt.Sprintf("%t", c))
			case float64:
				row.values[i] = cell(fmt.Sprintf("%f", c))
			case string:
				row.values[i] = cell(c)
			default:
				return nil, fmt.Errorf("unexpected value row %d at %d: type %T unhandled: %v", valueRowIdx, i, c, c)
			}
		}
		ret = append(ret, row)
	}
	return ret, nil
}

//
// Boilerplate from quickstart: https://developers.google.com/sheets/api/quickstart/go
//

func getSheetsClient() (*sheets.Service, error) {
	credentialsFilePath := "google_credentials.json"
	if name, ok := os.LookupEnv("GOOGLE_CREDENTIALS_PATH"); ok {
		credentialsFilePath = name
	}

	credentialsBytes, err := ioutil.ReadFile(credentialsFilePath)
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(credentialsBytes, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return nil, err
	}
	client := getClient(config)

	return sheets.New(client)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokenPath := "google_token.json"
	if name, ok := os.LookupEnv("GOOGLE_TOKEN_PATH"); ok {
		tokenPath = name
	}
	tok, err := tokenFromFile(tokenPath)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenPath, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
