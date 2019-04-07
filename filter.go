package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	pathSeps string = "/\\"

	skipPattern      *regexp.Regexp
	rmRegFilePattern *regexp.Regexp
	rmDirPattern     *regexp.Regexp
)

func init() {
	if !strings.ContainsRune(pathSeps, os.PathSeparator) {
		pathSeps += string(os.PathSeparator)
	}
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
	patternDir := filepath.Join(filepath.Dir(exePath), "patterns")
	skipPath := filepath.Join(patternDir, "skip.txt")
	rmPath := filepath.Join(patternDir, "remove.txt")
	var sbSkip, sbReg, sbDir strings.Builder
	var sb *strings.Builder

	skipFile, err := os.Open(skipPath)
	if err != nil {
		panic(err)
	}
	defer skipFile.Close() // Ignore error.
	scanner := bufio.NewScanner(skipFile)
	sb = &sbSkip
	for scanner.Scan() {
		t, _ := parseScanner(scanner)
		if t == "" {
			continue
		}
		sb.Grow(len(t) + 3)
		sb.WriteString("|(")
		sb.WriteString(t)
		sb.WriteRune(')')
	}
	err = scanner.Err()
	if err != nil {
		panic(err)
	}
	regex := sbSkip.String()
	if regex != "" {
		skipPattern = regexp.MustCompile(postProcessingRegexp(regex))
	}

	rmFile, err := os.Open(rmPath)
	if err != nil {
		panic(err)
	}
	defer rmFile.Close() // Ignore error.
	scanner = bufio.NewScanner(rmFile)
	for scanner.Scan() {
		t, isDir := parseScanner(scanner)
		if t == "" {
			continue
		}
		if isDir {
			sb = &sbDir
		} else {
			sb = &sbReg
		}
		sb.Grow(len(t) + 3)
		sb.WriteString("|(")
		sb.WriteString(t)
		sb.WriteRune(')')
	}
	err = scanner.Err()
	if err != nil {
		panic(err)
	}
	regex = sbReg.String()
	if regex != "" {
		rmRegFilePattern = regexp.MustCompile(postProcessingRegexp(regex))
	}
	regex = sbDir.String()
	if regex != "" {
		rmDirPattern = regexp.MustCompile(postProcessingRegexp(regex))
	}
}

// Ensure info != nil.
func skipFilter(info os.FileInfo) bool {
	if skipPattern == nil {
		return false
	}
	return skipPattern.MatchString(info.Name())
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

// Call it after scanner.Scan().
func parseScanner(scanner *bufio.Scanner) (text string, isDir bool) {
	t := strings.TrimSpace(scanner.Text())
	if t == "" || strings.HasPrefix(t, "//") {
		// Skip empty lines and comments.
		return "", false
	}
	// Remove comments:
	if idx := strings.Index(t, " //"); idx >= 0 {
		t = t[:idx]
	}
	t = strings.TrimSpace(t)
	isDir = strings.Contains(pathSeps, string(t[len(t)-1]))
	text = strings.Trim(t, pathSeps)
	return
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
