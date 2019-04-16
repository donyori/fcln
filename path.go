package main

import (
	"os"
	"path/filepath"
)

var homeDir, exePath, settingsPath, patternsDir string

func init() {
	var err error
	exePath, err = os.Executable()
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
	homeDir = filepath.Dir(exePath)
	settingsPath = filepath.Join(homeDir, "settings", "settings.json")
	patternsDir = filepath.Join(homeDir, "patterns")
}
