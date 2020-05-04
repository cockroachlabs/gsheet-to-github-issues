package main

import (
	"flag"
	"fmt"
	"log"
)

const (
	flagNameSpreadsheetID    = "spreadsheet_id"
	flagNameSpreadsheetRange = "spreadsheet_range"
)

var flagSpreadsheetID = flag.String(
	flagNameSpreadsheetID,
	"",
	"The ID of the spreadsheet to coordinate with.",
)

var flagSpreadsheetRange = flag.String(
	flagNameSpreadsheetRange,
	"",
	"The range of the spreadsheet to read.",
)

func main() {
	flag.Parse()

	if *flagSpreadsheetID == "" || *flagSpreadsheetRange == "" {
		log.Fatalf("--%s and --%s must be set", flagNameSpreadsheetID, flagNameSpreadsheetRange)
	}

	sheetsCli, err := getSheetsClient()
	if err != nil {
		log.Fatalf("failed to get client for google sheets: %v", err)
	}

	values, err := sheetsCli.Spreadsheets.Values.Get(*flagSpreadsheetID, *flagSpreadsheetRange).Do()
	if err != nil {
		log.Fatalf("Failed to get values: %v", err)
	}
	fmt.Printf("%#v\n", values)
}
