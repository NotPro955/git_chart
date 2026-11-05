package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const endpoint = "https://api.github.com/graphql"

type day struct {
	ContributionCount int          `json:"contributionCount"`
	ContributionLevel string       `json:"contributionLevel"`
	WeekDay           time.Weekday `json:"weekday"`
	Color             string       `json:"color"`
	Date              time.Time    `json:"date"`
}

type week struct {
	ContribuionDays []day        `json:"contributionDays"`
	FirstDay        time.Weekday `json:"firstDay"`
}

type month struct {
	Name       time.Month   `json:"name"`
	FirstDay   time.Weekday `json:"firstDay"`
	TotalWeeks int          `json:"totalWeeks"`
}

type calendar struct {
	TotalContributions int     `json:"totalContributions"`
	Weeks              []week  `json:"weeks"`
	Months             []month `json:"months"`
}

type collection struct {
	ContributionCalendar     calendar `json:"contributionCalendar"`
	TotalCommitContributions int      `json:"totalCommitContributions"`
	RestrictedContribution   int      `json:"restrictedContributionsCount"`
}

type user struct {
	ContributionCollection collection `json:"contributionsCollection"`
}

type request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type response struct {
	Data struct {
		User user `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func main() {
	godotenv.Load()

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB Token not found")
	}

	query := `
	query($login: String!, $from: DateTime, $to: DateTime) {
		user(login: $login){
			contributionsCollection(from: $from, to: $to){
				totalCommitContributions
				restrictedContributionsCount
				contributionCalendar {
					totalContributions
					weeks {
						firstday
						contributionDays {
							date
							weekday
							contributionCount
							contributionLevel
							color
						}
					}
					months {
						name
						firstday
						totalWeeks
					}
				}
			}
		}
	}`

	to := time.Now().UTC()
	from := to.AddDate(-1, 0, 0)

	body, err := json.Marshal(request{
		Query: query,
		Variables: map[string]any{
			"login": "NotPro955",
			"from":  from.Format(time.RFC3339),
			"to":    to.Format(time.RFC3339),
		},
	})

	if err != nil {
		log.Fatal("error: json body")
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		log.Fatal("Error: http Request")
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resq, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resq.Body.Close()

	raw, err := io.ReadAll(resq.Body)
	if err != nil {
		log.Fatal(err)
	}

	var out response
	if err := json.Unmarshal(raw, &out); err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
}
