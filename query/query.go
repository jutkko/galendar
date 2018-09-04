package query

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/closestmatch"
	calendar "google.golang.org/api/calendar/v3"
)

func Query(srv *calendar.Service, calendar string) {
	currTime := time.Now()
	tMin := currTime.Format(time.RFC3339)
	tMax := currTime.Add(24 * time.Hour).Format(time.RFC3339)

	list, err := srv.CalendarList.List().Do()
	if err != nil {
		log.Fatalf("Unable to find the calendar list from the provided service: %v", err)
	}

	calendarID := getMatchingCalendar(calendar, list.Items)
	if calendarID == "" {
		log.Fatalf("No matching calendar from the provided calendar: %s", calendar)
	}

	events, err := srv.Events.List(calendarID).ShowDeleted(false).
		SingleEvents(true).TimeMin(tMin).TimeMax(tMax).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v for calendar: %s", err, calendarID)
	}

	printEvents(calendarID, events.Items, currTime)
}

func printEvents(calendarID string, events []*calendar.Event, currTime time.Time) {
	if len(events) > 0 {
		for _, i := range events {
			atts := i.Attendees
			responseStatus := "accepted"
			for _, att := range atts {
				if att.Self {
					responseStatus = att.ResponseStatus
				}
			}
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
					green := color.New(color.FgGreen).SprintFunc()
					fmt.Printf(green("Happening now: %s\n"), fmtEvent(i.Summary, parseTimeHumanReadable(startTime), parseTimeHumanReadable(endTime), i.Location, responseStatus))
				} else {
					if startTime.Day() == currTime.Day() {
						fmt.Printf("%s\n", fmtEvent(i.Summary, parseTimeHumanReadable(startTime), parseTimeHumanReadable(endTime), i.Location, responseStatus))
					} else {
						yellow := color.New(color.FgYellow).SprintFunc()
						fmt.Printf(yellow("Not today: "))
						fmt.Printf("%s\n", fmtEvent(i.Summary, parseTimeHumanReadable(startTime), parseTimeHumanReadable(endTime), i.Location, responseStatus))
					}
				}
			} else {
				fmt.Printf("Full-day: %s (%s)\n", i.Summary, i.Start.Date)
			}
		}
	} else {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf(red("No upcoming events found.\n"))
	}
}

func getMatchingCalendar(queryCalendar string, calendarListEntries []*calendar.CalendarListEntry) string {
	var matchingCalendar string
	if queryCalendar != "" {
		matchingCalendar = getIDFromList(queryCalendar, calendarListEntries)

		if queryCalendar != matchingCalendar && matchingCalendar != "" {
			yellow := color.New(color.FgYellow).PrintfFunc()
			yellow("No exact match for %s, but found %s\n\n", queryCalendar, matchingCalendar)
		}
	} else {
		matchingCalendar = "primary"
	}

	return matchingCalendar
}

func parseTimeHumanReadable(t time.Time) string {
	// Use this specific format for readability
	return t.Format("15:04")
}

func parseTimeDateHumanReadable(t time.Time) string {
	// Use this specific format for readability
	return t.Format("Mon Jan _2 15:04")
}

func fmtEvent(summary, startTime, endTime, location, responseStatus string) string {
	if location == "" {
		location = "-"
	}

	toColour := fmt.Sprintf("%s @ %s %s-%s", summary, location, startTime, endTime)
	if responseStatus == "accepted" {
		green := color.New(color.FgGreen).SprintFunc()
		return green(toColour)
	}

	if responseStatus == "declined" {
		red := color.New(color.FgRed).SprintFunc()
		return red(toColour)
	}

	yellow := color.New(color.FgYellow).SprintFunc()
	return yellow(toColour)
}

func getIDFromList(calendarID string, calendars []*calendar.CalendarListEntry) string {
	infos := []string{}
	bagSizes := []int{3}
	for _, calendar := range calendars {
		infos = append(infos, calendar.Id)
	}

	cm := closestmatch.New(infos, bagSizes)
	return cm.Closest(calendarID)
}
