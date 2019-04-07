package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	rmRegFilePattern *regexp.Regexp
	rmDirPattern     *regexp.Regexp
)

func init() {
	// Compile patterns:
	exePath, err := os.Executable()
	if err != nil {
		exePath, err = filepath.Abs(os.Args[0])
		if err != nil {
			panic(err)
		}
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		panic(err)
	}
	rmPath := filepath.Join(filepath.Dir(exePath),
		"patterns", "remove.txt")
	rmFile, err := os.Open(rmPath)
	if err != nil {
		panic(err)
	}
	defer rmFile.Close() // Ignore error.
	var sbReg, sbDir strings.Builder
	scanner := bufio.NewScanner(rmFile)
	var sb *strings.Builder
	pathSeps := "/\\"
	if !strings.ContainsRune(pathSeps, os.PathSeparator) {
		pathSeps += string(os.PathSeparator)
	}
	for scanner.Scan() {
		t := strings.TrimSpace(scanner.Text())
		if t == "" || strings.HasPrefix(t, "//") {
			// Skip empty lines and comments.
			continue
		}
		// Remove comments:
		if idx := strings.Index(t, " //"); idx >= 0 {
			t = t[:idx]
		}
		t = strings.TrimSpace(t)
		if last := string(t[len(t)-1]); strings.Contains(pathSeps, last) {
			sb = &sbDir
		} else {
			sb = &sbReg
		}
		t = strings.Trim(t, pathSeps)
		sb.Grow(len(t) + 3)
		sb.WriteString("|(")
		sb.WriteString(t)
		sb.WriteRune(')')
	}
	err = scanner.Err()
	if err != nil {
		panic(err)
	}
	regex := sbReg.String()
	if regex != "" {
		rmRegFilePattern = regexp.MustCompile(postProcessingRegexp(regex))
	}
	regex = sbDir.String()
	if regex != "" {
		rmDirPattern = regexp.MustCompile(postProcessingRegexp(regex))
	}
}

// Ensure info != nil && !info.IsDir().
func removeRegFileFilter(info os.FileInfo) bool {
	if rmRegFilePattern == nil {
		return false
	}
	return rmRegFilePattern.MatchString(info.Name())
}

// Ensure info != nil && info.IsDir().
func removeDirFilter(info os.FileInfo) bool {
	if rmDirPattern == nil {
		return false
	}
	return rmDirPattern.MatchString(info.Name())
}

func postProcessingRegexp(regex string) string {
	begin, end := 1, len(regex) // begin is 1 to remove first '|'
	if !strings.Contains(regex, ")|(") {
		// Remove the only pair of parentheses added by string builder.
		begin += 1
		end -= 1
	}
	regex = regex[begin:end]
	return regex
}
