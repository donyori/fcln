package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

func main() {
	pDoesAutoRemove := flag.Bool("-y", false,
		"True if remove files without confirmation.")
	flag.Parse()
	args := flag.Args()
	startTime := time.Now()
	toRemove, err := GetToBeRemovedFiles(args...)
	fmt.Println("Elapsed", time.Since(startTime))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if len(toRemove) == 0 {
		fmt.Println("No file to remove.")
		return
	}
	PrintBatchList(toRemove)
	if !*pDoesAutoRemove {
		fmt.Print("Remove above file(s)? (Yes/No):")
		choice := ""
		for choice != "y" && choice != "n" {
			if choice != "" {
				fmt.Print("Please input Yes or No:")
			}
			fmt.Scanln(&choice)
			choice = strings.ToLower(choice)
			if choice == "yes" {
				choice = "y"
			} else if choice == "no" {
				choice = "n"
			}
		}
		if choice == "n" {
			fmt.Println("Give up removing file(s).")
			return
		}
	}
	startTime = time.Now()
	RemoveFiles(toRemove)
	fmt.Println("Finish. Elapsed", time.Since(startTime))
}
