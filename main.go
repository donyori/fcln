package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	startTime := time.Now()
	toRemove, err := getToRemove(os.Args[1:]...)
	fmt.Println("Elapsed", time.Since(startTime))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if len(toRemove) == 0 {
		fmt.Println("No file to remove.")
		return
	}
	printBatchList(toRemove)
}
