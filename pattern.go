package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

func loadPatterns(filename string) (regFilePattern, dirPattern *regexp.Regexp,
	err error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	var sbReg, sbDir strings.Builder
	var sb *strings.Builder
	scanner := bufio.NewScanner(file)
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
		if strings.Contains(pathSeps, string(t[len(t)-1])) {
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
		return nil, nil, err
	}
	regex := sbReg.String()
	if regex != "" {
		regFilePattern = regexp.MustCompile(postProcessingRegexp(regex))
	}
	regex = sbDir.String()
	if regex != "" {
		dirPattern = regexp.MustCompile(postProcessingRegexp(regex))
	}
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
