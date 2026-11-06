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

type date struct{ time.Time }

func (d *date) UnmarshalJSON(b []byte) error {
	// the quotes are part of the raw JSON token
	t, err := time.Parse(`"2006-01-02"`, string(b))
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

const endpoint = "https://api.github.com/graphql"

type day struct {
	ContributionCount int          `json:"contributionCount"`
	ContributionLevel string       `json:"contributionLevel"`
	WeekDay           time.Weekday `json:"weekday"`
	Color             string       `json:"color"`
	Date              date         `json:"date"`
}

type week struct {
	ContribuionDays []day `json:"contributionDays"`
	//FirstDay        time.Weekday `json:"firstDay"`
}

type month struct {
	Name       string `json:"name"`
	FirstDay   date   `json:"firstDay"`
	TotalWeeks int    `json:"totalWeeks"`
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

	username := os.Getenv("GITHUB_USER")
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
						contributionDays {
							date
							weekday
							contributionCount
							contributionLevel
							color
						}
					}
					months {
						firstDay
						name
						totalWeeks
					}
				}
			}
		}
	}`

	to := time.Now().UTC()
	from := to.AddDate(-1, 0, 0)

	fmt.Println(to.Format(time.RFC3339))
	fmt.Println(from.Format(time.RFC3339))

	//fromT, err := time.Parse("2006-01-02", "2025-11-06")
	//if err != nil {
	//	log.Fatal(err)
	//}

	body, err := json.Marshal(request{
		Query: query,
		Variables: map[string]any{
			"login": username,
			"from":  from.Format(time.RFC3339),
			"to":    to.Format(time.RFC3339),
		},
	})

	fmt.Println("JSON formated")

	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		log.Fatal("Error: http Request")
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	fmt.Println("HTTP header ready")

	client := &http.Client{Timeout: 15 * time.Second}
	resq, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resq.Body.Close()

	fmt.Println("Request sent")

	raw, err := io.ReadAll(resq.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Reading response")

	var out response
	if err := json.Unmarshal(raw, &out); err != nil {
		log.Fatalf("response error: %v", err)
	}
	if len(out.Errors) > 0 {
		log.Fatalf("graphql error: %s", out.Errors[0].Message)
	}

	col := out.Data.User.ContributionCollection
	cal := col.ContributionCalendar

	fmt.Println("Range:", from.Format("2006-01-02"), "to", to.Format("2006-01-02"))
	fmt.Println("Total contributions:", cal.TotalContributions)
	fmt.Println("Commit contributions:", col.TotalCommitContributions)
	fmt.Println("Restricted (private):", col.RestrictedContribution)

	fmt.Println("\nMonths:")
	for _, m := range cal.Months {
		fmt.Printf("  %-4s starts %s, spans %d weeks\n",
			m.Name, m.FirstDay.Format("2006-01-02"), m.TotalWeeks)
	}

	fmt.Println("\nDays:")
	for _, w := range cal.Weeks {
		for _, d := range w.ContribuionDays {
			fmt.Printf("  %s %-9s %3d  %-15s %s\n",
				d.Date.Format("2006-01-02"), d.WeekDay, d.ContributionCount,
				d.ContributionLevel, d.Color)
		}
	}
}
