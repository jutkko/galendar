package query

import (
	"testing"

	calendar "google.golang.org/api/calendar/v3"
)

func TestPrintEvents(t *testing.T) {
}

func TestGetMatchingCalendar(t *testing.T) {
	calendars := []*calendar.CalendarListEntry{
		&calendar.CalendarListEntry{
			Id: "myCalendar",
		},
		&calendar.CalendarListEntry{
			Id: "yourCalendar",
		},
	}

	matchingCalendar := getMatchingCalendar("myCalendar", calendars)
	if matchingCalendar != "myCalendar" {
		t.Error("Expected 'myCalendar' but got ", matchingCalendar)
	}
}

func TestGetMatchingCalendarPartial(t *testing.T) {
	calendars := []*calendar.CalendarListEntry{
		&calendar.CalendarListEntry{
			Id: "myCale",
		},
		&calendar.CalendarListEntry{
			Id: "notYours",
		},
	}

	matchingCalendar := getMatchingCalendar("myCalendar", calendars)
	if matchingCalendar != "myCale" {
		t.Error("Expected 'myCale' but got ", matchingCalendar)
	}
}

func TestGetMatchingCalendarNoCalendarID(t *testing.T) {
	calendars := []*calendar.CalendarListEntry{
		&calendar.CalendarListEntry{
			Id: "myCale",
		},
		&calendar.CalendarListEntry{
			Id: "notYours",
		},
	}

	matchingCalendar := getMatchingCalendar("", calendars)
	if matchingCalendar != "primary" {
		t.Error("Expected 'primary' but got ", matchingCalendar)
	}
}

func TestGetMatchingCalendarNoMatch(t *testing.T) {
	calendars := []*calendar.CalendarListEntry{
		&calendar.CalendarListEntry{
			Id: "myCale",
		},
		&calendar.CalendarListEntry{
			Id: "notYours",
		},
	}

	matchingCalendar := getMatchingCalendar("heyYou", calendars)
	if matchingCalendar != "" {
		t.Error("Expected '' but got ", matchingCalendar)
	}
}

func TestFmtEvent(t *testing.T) {
	fmtEvent := fmtEvent("We eat veggie hotdogs!", "10:00", "12:00", "London")
	if fmtEvent != "We eat veggie hotdogs! 10:00-12:00 @ London" {
		t.Error("Expected 'We eat veggie hotdogs! 10:00-12:00 @ London' but got ", fmtEvent)
	}
}

func TestFmtEventEmptyLocation(t *testing.T) {
	exptectedString := "We eat veggie hotdogs! 10:00-12:00 @ -"
	fmtEvent := fmtEvent("We eat veggie hotdogs!", "10:00", "12:00", "")
	if fmtEvent != exptectedString {
		t.Error("Expected 'We eat veggie hotdogs! 10:00-12:00 @ -' but got ", fmtEvent)
	}
}

func TestGetIDFromList(t *testing.T) {
	expectedCalendarID := "mycalendar"
	calendars := []*calendar.CalendarListEntry{
		&calendar.CalendarListEntry{
			Id: "mycalendar",
		},
		&calendar.CalendarListEntry{
			Id: "yourCalendar",
		},
	}

	actualCalendarID := getIDFromList("mycal", calendars)
	if expectedCalendarID != actualCalendarID {
		t.Error("Expected 'mycalendar' but got ", actualCalendarID)
	}
}

func TestGetIDFromListNoMatch(t *testing.T) {
	expectedCalendarID := ""
	calendars := []*calendar.CalendarListEntry{
		&calendar.CalendarListEntry{
			Id: "mycalendar",
		},
		&calendar.CalendarListEntry{
			Id: "yourCalendar",
		},
	}

	actualCalendarID := getIDFromList("oneTwoThree", calendars)
	if expectedCalendarID != actualCalendarID {
		t.Error("Expected '' but got ", actualCalendarID)
	}
}
