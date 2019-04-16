package main

import "os"

// Ensure info != nil and !info.IsDir().
func skipRegFileFilter(info os.FileInfo) bool {
	lazyLoadPatterns()
	if skpRegFilePattern == nil {
		return false
	}
	return skpRegFilePattern.MatchString(info.Name())
}

// Ensure info != nil && info.IsDir().
func skipDirFilter(info os.FileInfo) bool {
	lazyLoadPatterns()
	if skpDirPattern == nil {
		return false
	}
	return skpDirPattern.MatchString(info.Name())
}

// Ensure info != nil && !info.IsDir().
func removeRegFileFilter(info os.FileInfo) bool {
	lazyLoadPatterns()
	if rmRegFilePattern == nil {
		return false
	}
	return rmRegFilePattern.MatchString(info.Name())
}

// Ensure info != nil && info.IsDir().
func removeDirFilter(info os.FileInfo) bool {
	lazyLoadPatterns()
	if rmDirPattern == nil {
		return false
	}
	return rmDirPattern.MatchString(info.Name())
}
