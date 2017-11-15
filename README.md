# galendar
This is a golang binary that will retrieve the current and future events from a
given calendar. 

# How to use
Want to see what is coming up in someone's calendar? Provided that you have
access to the calendar, just do `galendar someone@example.com`. An example output
can look like
```
Upcoming events for someone@example.com:
Full-day: Weekend Yey (2017-10-08)
Happening now: Coding (11:00-11:30)
阅读 (19:00-20:00)
```

## Fuzzy matching
galendar can also infer the calendar from an inpartial input, e.g., when the
email is long and you only remember a part of it. `galendar someone` or
`galendar one` will also return the correct result. Note that in order to use
this, you must have the calendar
[added](https://support.google.com/calendar/answer/37100?co=GENIE.Platform%3DDesktop&hl=en)
to your calendar list.

# Installation
Run

```
go get github.com/jutkko/galendar
```

## Prerequisites
You will need to have [golang](https://golang.org/dl/) installed on your machine.

galendar calls the [Google Calendar
API](https://developers.google.com/google-apps/calendar/). To be able to read
information from the API, you will need to enable the API. For guidance, please
see the
[docs](https://developers.google.com/google-apps/calendar/quickstart/dotnet#step_1_turn_on_the_api_name).

After having downloaded the credentials from the docs, please move the file to
`~/.credentials/galendar_client_secret.json`. Depending on which Google Account
you have created the secret with, the corresponsing calendars will be shown
when running `galendar`.

# Uninstall
Simple. 

```
rm -rf $GOPATH/src/github.com/jutkko/galendar
rm $GOPATH/bin/galendar
```
