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
	"strings"
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
			fmt.Printf("Usage: galendar [calendar]\n")
			os.Exit(0)
		case "-h":
			fmt.Printf("Usage: galendar [calendar]\n")
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
	t := time.Now()
	tMin := t.Format(time.RFC3339)
	tMax := t.Add(24 * time.Hour).Format(time.RFC3339)

	var calendarID string
	var err error
	if calendar != "" {
		calendarID, err = getIDFromList(srv, calendar)
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

	events, err := srv.Events.List(calendarID).ShowDeleted(false).
		SingleEvents(true).TimeMin(tMin).TimeMax(tMax).MaxResults(20).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v for calendar %s", err, calendarID)
	}

	fmt.Printf("Upcoming events for %s:\n\n", calendarID)
	if len(events.Items) > 0 {
		for _, i := range events.Items {
			var start string
			var end string
			// If the DateTime is an empty string the Event is an all-day Event.
			// So only Date is available.
			if i.Start.DateTime != "" {
				start = i.Start.DateTime
				end = i.End.DateTime
				startTime, err := time.Parse(time.RFC3339, start)
				if err != nil {
					log.Fatalf("Failed to parse event's time: %v", err)
				}

				start = onlyShowTime(start)
				end = onlyShowTime(end)
				if t.After(startTime) {
					fmt.Printf("Happening now: %s (%s-%s)\n", i.Summary, start, end)
				} else {
					fmt.Printf("%s (%s-%s)\n", i.Summary, start, end)
				}
			} else {
				start = i.Start.Date
				fmt.Printf("Full-day: %s (%s)\n", i.Summary, start)
			}
		}
	} else {
		fmt.Printf("No upcoming events found.\n")
	}
}

func onlyShowTime(dateTime string) string {
	time := strings.Split(strings.Split(dateTime, "T")[1], "Z")[0]
	return time
}

func getIDFromList(srv *calendar.Service, calendarID string) (string, error) {
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
