package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

type Settings struct {
	PermissionErrorHandling ErrorHandling `json:"permission_error_handling"`
}

var (
	settings         *Settings
	loadSettingsOnce sync.Once
)

func lazyLoadSettings() {
	loadSettingsOnce.Do(func() {
		file, err := os.Open(settingsPath)
		if err != nil {
			if err == os.ErrNotExist {
				settings = &Settings{PermissionErrorHandling: Warn}
				return
			}
			panic(err)
		}
		defer file.Close() // Ignore error.
		data, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		s := &Settings{PermissionErrorHandling: Warn}
		err = json.Unmarshal(data, s)
		if err != nil {
			panic(err)
		}
		settings = s
	})
}
