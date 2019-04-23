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
	fmt.Println("------------------")
	var count int
	for _, batch := range toRemove {
		fmt.Println(batch.Parent)
		files := make([]string, 0, len(batch.Dirs)+len(batch.RegFiles)+
			len(batch.Symlinks)+len(batch.Others))
		files = append(files, batch.Dirs...)
		files = append(files, batch.RegFiles...)
		files = append(files, batch.Symlinks...)
		files = append(files, batch.Others...)
		for _, file := range files {
			fmt.Println("   ", file)
			count++
		}
	}
	fmt.Println("------------------")
	fmt.Println("Total", count, "files.")
}
