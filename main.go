package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/schollz/closestmatch"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

func main() {
	srv := getService()

	var calendar string
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--help":
			fallthrough
		case "-h":
			fmt.Printf("Usage: galendar [someone's name]\n")
			fmt.Printf("  To find someone's calendar events\n")
			os.Exit(0)
		default:
			calendar = os.Args[1]
		}
	}

	query(srv, calendar)
}

// getService does the oauth dance and creates a service from the provided credentials
func getService() *calendar.Service {
	ctx := context.Background()

	// Get client secret
	b, err := getClientSecret()
	if err != nil {
		log.Fatalf("Unable to get client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/calendar-go-quickstart.json
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
	}

	return srv
}

// getClientSecret read the json file in ~/.credentials
func getClientSecret() ([]byte, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	clientSecretFile := filepath.Join(usr.HomeDir, ".credentials/galendar_client_secret.json")
	return ioutil.ReadFile(clientSecretFile)
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}

	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	return filepath.Join(tokenCacheDir, "galendar_client_token.json"), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}

	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func query(srv *calendar.Service, calendar string) {
	currTime := time.Now()
	tMin := currTime.Format(time.RFC3339)
	tMax := currTime.Add(48 * time.Hour).Format(time.RFC3339)

	calendarID := getMatchingCalendar(calendar, srv)
	events, err := srv.Events.List(calendarID).ShowDeleted(false).
		SingleEvents(true).TimeMin(tMin).TimeMax(tMax).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v for calendar: %s", err, calendarID)
	}

	printEvents(calendarID, events, currTime)
}

func printEvents(calendarID string, events *calendar.Events, currTime time.Time) {
	if len(events.Items) > 0 {
		fmt.Printf("Upcoming events for %s:\n\n", calendarID)

		for _, i := range events.Items {
			// If the DateTime is an empty string the Event is an all-day Event.
			// So only Date is available.
			if i.Start.DateTime != "" {
				startTime, err := time.Parse(time.RFC3339, i.Start.DateTime)
				if err != nil {
					log.Fatalf("Failed to parse event's time: %v", err)
				}

				endTime, err := time.Parse(time.RFC3339, i.End.DateTime)
				if err != nil {
					log.Fatalf("Failed to parse event's time: %v", err)
				}

				if currTime.After(startTime) {
					fmt.Printf("Happening now: %s\n", fmtEvent(i.Summary, parseTimeHumanReadable(startTime), parseTimeHumanReadable(endTime), i.Location))
				} else {
					if startTime.Day() == currTime.Day() {
						fmt.Printf("%s\n", fmtEvent(i.Summary, parseTimeHumanReadable(startTime), parseTimeHumanReadable(endTime), i.Location))
					} else {
						fmt.Printf("Not today: ")
						fmt.Printf("%s\n", fmtEvent(i.Summary, parseTimeHumanReadable(startTime), parseTimeHumanReadable(endTime), i.Location))
					}
				}
			} else {
				fmt.Printf("Full-day: %s (%s)\n", i.Summary, i.Start.Date)
			}
		}
	} else {
		fmt.Printf("No upcoming events found.\n")
	}
}

func getMatchingCalendar(calendar string, srv *calendar.Service) string {
	var calendarID string
	var err error
	if calendar != "" {
		calendarID, err = getIDFromList(calendar, srv)
		if err != nil {
			log.Fatalf("Unable to find a calendar from the provided calendar %s: %v", calendarID, err)
		}

		if calendarID == "" {
			log.Fatalf("No matching calendar from the provided calendar: %s", calendar)
		}

		if calendar != calendarID {
			fmt.Printf("No exact match for %s, but found %s\n\n", calendar, calendarID)
		}
	} else {
		calendarID = "primary"
	}

	return calendarID
}

func parseTimeHumanReadable(t time.Time) string {
	// Use this specific format for readability
	return t.Format("15:04")
}

func parseTimeDateHumanReadable(t time.Time) string {
	// Use this specific format for readability
	return t.Format("Mon Jan _2 15:04")
}

func fmtEvent(summary, startTime, endTime, location string) string {
	if location == "" {
		location = "-"
	}

	return fmt.Sprintf("%s %s-%s @ %s", summary, startTime, endTime, location)
}

func getIDFromList(calendarID string, srv *calendar.Service) (string, error) {
	list, err := srv.CalendarList.List().Do()
	if err != nil {
		return "", err
	}

	infos := []string{}
	bagSizes := []int{3}
	for _, calendar := range list.Items {
		infos = append(infos, calendar.Id)
	}

	cm := closestmatch.New(infos, bagSizes)

	return cm.Closest(calendarID), nil
}
