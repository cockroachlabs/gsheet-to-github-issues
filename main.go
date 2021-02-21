package main

import (
	"flag"
	"html/template"
	"log"
	"strings"
)

const (
	flagNameSpreadsheetID    = "spreadsheet_id"
	flagNameSpreadsheetRange = "spreadsheet_range"
	flagNameTemplate         = "template"
)

var (
	flagSpreadsheetID = flag.String(
		flagNameSpreadsheetID,
		"",
		"The ID of the spreadsheet to coordinate with.",
	)

	flagSpreadsheetRange = flag.String(
		flagNameSpreadsheetRange,
		"",
		"The range of the spreadsheet to read.",
	)

	flagTemplate = flag.String(
		flagNameTemplate,
		"",
		"The name of the template to use",
	)

	flagGithubOwner = flag.String(
		"github_owner",
		"cockroachdb",
		"The owner on GitHub to sync with.",
	)

	flagGithubRepo = flag.String(
		"github_repo",
		"cockroach",
		"The repo on GitHub to sync with.",
	)

	flagProjectColumnID = flag.Int(
		"github_project_column_id",
		0,
		"The ColumnID to place any new issue in.",
	)
)

func main() {
	flag.Parse()

	if *flagSpreadsheetID == "" || *flagTemplate == "" || *flagSpreadsheetRange == "" {
		log.Fatalf("--%s, --%s and --%s must be set", flagNameSpreadsheetID, flagNameTemplate, flagNameSpreadsheetRange)
	}

	tpl, err := template.ParseFiles("templates/" + *flagTemplate)
	if err != nil {
		log.Fatalf("failed to parse template: %v", err)
	}

	// Grab necessary clients.
	sheetsClient, err := getSheetsClient()
	if err != nil {
		log.Fatalf("failed to get client for google sheets: %v", err)
	}

	ghClient, err := getGithubClient()
	if err != nil {
		log.Fatalf("failed to get client for github: %v", err)
	}

	// Read the rows of the sheets.
	log.Printf("grabbing spreadsheet values from google sheets...")
	values, err := sheetsClient.Spreadsheets.Values.Get(*flagSpreadsheetID, *flagSpreadsheetRange).Do()
	if err != nil {
		log.Fatalf("failed to get values: %v", err)
	}

	headers, err := headersFromSheet(values)
	if err != nil {
		log.Fatalf("failed to get headers: %v", err)
	}

	rows, err := sheetToRows(headers, values)
	if err != nil {
		log.Fatalf("failed to convert rows: %v", err)
	}

	spreadsheetName := strings.Split((*flagSpreadsheetRange), "!")[0]

	// Sync with Github Issues.
	log.Print("syncing with github issues...")

	updates, err := sync(tpl, spreadsheetName, headers, rows, ghClient, int64(*flagProjectColumnID))
	if err != nil {
		log.Fatalf("failed to sync: %v", err)
	}

	// Re-sync with Google Sheets.
	log.Printf("re-syncing with google sheets...")
	for _, update := range updates {
		if _, err := sheetsClient.Spreadsheets.Values.Update(*flagSpreadsheetID, update.Range, update).ValueInputOption("USER_ENTERED").Do(); err != nil {
			log.Fatalf("failed to update range in google sheets: %v", err)
		}
	}

	log.Printf("sync complete!")
}
