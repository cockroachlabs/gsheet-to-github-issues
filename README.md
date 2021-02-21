# gsheets-to-github-issues

This project syncs a Google Sheet containing issues that require filing
and/or syncing with Github Issues. Once an issue is created, it will
edit the issue if needbe.

## Instructions

* Create a Google Sheet with columns (`github_issue`, `github_status`, `github_title`,
  `github_labels`, `github_ignore`, `github_assigned`, `github_reacts`).
* Ensure `github_title` is filled in with some titles. [Example spreadsheet](https://docs.google.com/spreadsheets/d/1USkxXFMZr_4lvnFR8YjUtJwdfCpZ0D6bI2uHDcV32-I/edit#gid=0).
* Get a [Google API Sheets Key](https://developers.google.com/sheets/api/quickstart/go)
  `credentials.json` file, and put it in `google_credentials.json` in the running directory.
  If doing this yourself, set the redirect URI to `http://localhost`.
* Get a [Github Personal Access Token](https://github.com/settings/tokens),
  and set the env `export GITHUB_API_KEY=<access token>`.
* Run the tool `go run . --spreadsheet_id='<spreadsheet_id>' --spreadsheet_range='<spreadsheet sheet name>' --template='<template file in templates/'`. Spreadsheet range can be a subset, e.g. `Geography!A1:C10`, but the header row must be the first row available.
  * On first invocation, you will need to follow the OAuth flow.
    Follow it, and paste the "code" in the URL into the command line.
    You will only need to do this once.
  * Optionally, you can use `--github_project_column_id=<col_id>` to
    automagically put this onto a Github column.

## Example invocations

* Geography: `go run . --spreadsheet_id=1USkxXFMZr_4lvnFR8YjUtJwdfCpZ0D6bI2uHDcV32-I --spreadsheet_range="Geography" --template=geography_builtin.tmpl.md` --github_project_column_id=8546137
* Geometry: `go run . --spreadsheet_id=1USkxXFMZr_4lvnFR8YjUtJwdfCpZ0D6bI2uHDcV32-I --spreadsheet_range="Geometry" --template=geometry_builtin.tmpl.md`--github_project_column_id=8546137
* GeometryCreations: `go run . --spreadsheet_id=1USkxXFMZr_4lvnFR8YjUtJwdfCpZ0D6bI2uHDcV32-I --spreadsheet_range="GeometryCreations" --template=geometry_creations_builtin.tmpl.md --github_project_column_id=8546137`
* Altogether: `go run . --spreadsheet_id=1USkxXFMZr_4lvnFR8YjUtJwdfCpZ0D6bI2uHDcV32-I --spreadsheet_range="Geography" --template=geography_builtin.tmpl.md --github_project_column_id=8546137 && go run . --spreadsheet_id=1USkxXFMZr_4lvnFR8YjUtJwdfCpZ0D6bI2uHDcV32-I --spreadsheet_range="GeometryCreations" --template=geometry_creations_builtin.tmpl.md --github_project_column_id=8546137 && go run . --spreadsheet_id=1USkxXFMZr_4lvnFR8YjUtJwdfCpZ0D6bI2uHDcV32-I --spreadsheet_range="Geometry" --template=geometry_builtin.tmpl.md --github_project_column_id=8546137`
