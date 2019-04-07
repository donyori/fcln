package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	startTime := time.Now()
	toRemove, err := getToRemove(true, os.Args[1:]...)
	fmt.Println("Elapsed", time.Since(startTime))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if len(toRemove) == 0 {
		fmt.Println("No file to remove.")
		return
	}
	fmt.Println("------------------")
	for _, batch := range toRemove {
		fmt.Println(batch.Parent)
		for _, dir := range batch.Dirs {
			fmt.Println("   ", dir)
		}
		for _, regFile := range batch.RegFiles {
			fmt.Println("   ", regFile)
		}
	}
}
