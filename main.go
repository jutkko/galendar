package main

import (
	"fmt"
	"os"

	"github.com/jutkko/galendar/auth"
	"github.com/jutkko/galendar/query"
)

func main() {
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

	srv := auth.GetService()
	query.Query(srv, calendar)
}
