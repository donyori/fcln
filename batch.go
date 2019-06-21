package main

import (
	"fmt"
	"io"
)

type Batch struct {
	Parent   string
	Dirs     []string
	RegFiles []string
	Symlinks []string
	Others   []string
}

type BatchList []*Batch

func (bl BatchList) Len() int {
	return len(bl)
}

func (bl BatchList) Less(i, j int) bool {
	return bl[i] != nil && (bl[j] == nil || bl[i].Parent < bl[j].Parent)
}

func (bl BatchList) Swap(i, j int) {
	bl[i], bl[j] = bl[j], bl[i]
}

func (bl BatchList) Print(w io.Writer) {
	if len(bl) == 0 {
		fmt.Fprintln(w, "<empty batch list>")
		return
	}
	var count int
	printFiles := func(name string, files []string) {
		n := len(files)
		if n == 0 {
			return
		}
		count += n
		fmt.Fprintf(w, "  %s:\n", name)
		for _, file := range files {
			if file == "" {
				count--
				continue
			}
			fmt.Fprintln(w, "   ", file)
		}
	}
	bar18 := "------------------"
	fmt.Fprintln(w, bar18)
	for _, batch := range bl {
		fmt.Fprintln(w, batch.Parent)
		printFiles("Directories", batch.Dirs)
		printFiles("Regular files", batch.RegFiles)
		printFiles("Symlinks", batch.Symlinks)
		printFiles("Others", batch.Others)
	}
	fmt.Fprintln(w, bar18)
	fmt.Fprintln(w, "Total", count, "files.")
}
