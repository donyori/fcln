package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	pathSeps string = "/\\"

	skpRegFilePattern, skpDirPattern *regexp.Regexp
	rmRegFilePattern, rmDirPattern   *regexp.Regexp
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
	skpPath := filepath.Join(patternDir, "skip.txt")
	rmPath := filepath.Join(patternDir, "remove.txt")
	skpRegFilePattern, skpDirPattern, err = loadPatterns(skpPath)
	if err != nil {
		panic(err)
	}
	rmRegFilePattern, rmDirPattern, err = loadPatterns(rmPath)
	if err != nil {
		panic(err)
	}
}

// Ensure info != nil and !info.IsDir().
func skipRegFileFilter(info os.FileInfo) bool {
	if skpRegFilePattern == nil {
		return false
	}
	return skpRegFilePattern.MatchString(info.Name())
}

// Ensure info != nil && info.IsDir().
func skipDirFilter(info os.FileInfo) bool {
	if skpDirPattern == nil {
		return false
	}
	return skpDirPattern.MatchString(info.Name())
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
