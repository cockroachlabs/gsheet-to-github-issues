package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v30/github"
	"google.golang.org/api/sheets/v4"
)

type templateInput struct {
	GithubUser    string
	GithubUserURL string
	Time          string
	Args          map[string]cell
}

func makeValueRange(
	spreadsheetName string, numRows int, headers *headerRow, h string,
) *sheets.ValueRange {
	return &sheets.ValueRange{
		MajorDimension: "ROWS",
		Range: fmt.Sprintf(
			"%s!%s%d:%s%d",
			spreadsheetName,
			headers.mustGetLetter(h),
			2,
			headers.mustGetLetter(h),
			numRows+1,
		),
		Values: [][]interface{}{},
	}
}

// sync syncs between the sheet rows and the Google Sheet.
// It returns the updates necessary for Google Sheets.
func sync(
	tpl *template.Template,
	spreadsheetName string,
	headers *headerRow,
	rows []sheetRow,
	ghClient *github.Client,
) ([]*sheets.ValueRange, error) {
	updateIssues := makeValueRange(spreadsheetName, len(rows), headers, "github_issue")
	updateStatus := makeValueRange(spreadsheetName, len(rows), headers, "github_status")
	updateAssigned := makeValueRange(spreadsheetName, len(rows), headers, "github_assigned")
	updateReacts := makeValueRange(spreadsheetName, len(rows), headers, "github_reacts")

	ctx := context.Background()
	timeNow := time.Now().UTC().Format(time.RFC3339)
	ghUser, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		title := row.mustGet("github_title")
		issue := row.mustGet("github_issue")
		labels := strings.Split(row.mustGet("github_labels"), ",")
		sort.Strings(labels)
		ignore, ok := row.get("github_ignore")
		if ok && ignore != "" {
			log.Printf(" ignoring %q", title)
			updateIssues.Values = append(updateIssues.Values, []interface{}{issue})
			updateStatus.Values = append(updateStatus.Values, []interface{}{row.mustGet("github_status")})
			updateAssigned.Values = append(updateAssigned.Values, []interface{}{row.mustGet("github_assigned")})
			updateReacts.Values = append(updateReacts.Values, []interface{}{row.mustGet("github_reacts")})
			continue
		}

		var buf bytes.Buffer
		tpl.Execute(&buf, templateInput{
			GithubUser:    ghUser.GetLogin(),
			GithubUserURL: ghUser.GetHTMLURL(),
			Time:          timeNow,
			Args:          row.toMap(),
		})
		body := buf.String()
		body += fmt.Sprintf(
			"\n\n<sub>:robot: This issue was synced with a spreadsheet by [gsheets-to-github-issues](https://github.com/cockroachlabs/gsheet-to-github-issues) by [%s](%s) on %s. Changes to titles, body and labels may be overwritten.</sub>",
			ghUser.GetLogin(),
			ghUser.GetHTMLURL(),
			timeNow,
		)

		var iss *github.Issue
		if issue == "" {
			iss, _, err = ghClient.Issues.Create(ctx, *flagGithubOwner, *flagGithubRepo, &github.IssueRequest{
				Title:  &title,
				Body:   &body,
				Labels: &labels,
			})
			log.Printf("* created issue for %q: %s", title, iss.GetHTMLURL())
			if err != nil {
				return nil, err
			}
		} else {
			// Truncate leading `#`.
			if len(issue) == 0 || issue[0] != '#' {
				return nil, fmt.Errorf("expecting issue to lead with #, found %q", issue)
			}
			issueNum, err := strconv.ParseInt(issue[1:], 10, 64)
			if err != nil {
				return nil, err
			}
			iss, _, err = ghClient.Issues.Get(ctx, *flagGithubOwner, *flagGithubRepo, int(issueNum))
			if err != nil {
				return nil, err
			}
			log.Printf("* checking existing issue for %q: %s", title, iss.GetHTMLURL())
			if !bodyMatch(body, iss.GetBody()) || iss.GetTitle() != title || !labelsMatch(labels, iss.Labels) {
				log.Printf("** modifying issue as there is a mismatch found")
				iss, _, err = ghClient.Issues.Edit(ctx, *flagGithubOwner, *flagGithubRepo, int(issueNum), &github.IssueRequest{
					Title:  &title,
					Body:   &body,
					Labels: &labels,
				})
				if err != nil {
					return nil, err
				}
			}
		}

		updateIssues.Values = append(updateIssues.Values, []interface{}{fmt.Sprintf(`=HYPERLINK("%s", "#%d")`, iss.GetHTMLURL(), iss.GetNumber())})
		updateStatus.Values = append(updateStatus.Values, []interface{}{iss.GetState()})
		var assigneeVal string
		if iss.GetAssignee() != nil {
			assigneeVal = fmt.Sprintf(`=HYPERLINK("%s", "@%s")`, iss.GetAssignee().GetHTMLURL(), iss.GetAssignee().GetLogin())
		}
		updateAssigned.Values = append(updateAssigned.Values, []interface{}{assigneeVal})
		updateReacts.Values = append(updateReacts.Values, []interface{}{iss.GetReactions().GetTotalCount()})
	}
	return []*sheets.ValueRange{updateIssues, updateStatus, updateAssigned, updateReacts}, nil
}

func bodyMatch(a string, b string) bool {
	aSplit := strings.Split(a, "\n\n")
	bSplit := strings.Split(b, "\n\n")
	return strings.Join(aSplit[:len(aSplit)-1], "\n\n") == strings.Join(bSplit[:len(bSplit)-1], "\n\n")
}

func labelsMatch(a []string, gh []*github.Label) bool {
	b := make([]string, len(gh))
	for i := range gh {
		b[i] = gh[i].GetName()
	}
	sort.Strings(b)
	return reflect.DeepEqual(a, b)
}
