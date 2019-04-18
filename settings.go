package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/donyori/gocommfw"
)

type Settings struct {
	Worker                  gocommfw.WorkerSettings `json:"worker"`
	PermissionErrorHandling ErrorHandling           `json:"permission_error_handling"`
}

var (
	settings         *Settings
	loadSettingsOnce sync.Once
)

func newSettings() *Settings {
	settings := &Settings{
		Worker:                  *gocommfw.NewWorkerSettings(),
		PermissionErrorHandling: Warn,
	}
	if settings.Worker.Number > 1 {
		settings.Worker.Number--
	}
	return settings
}

func lazyLoadSettings() {
	loadSettingsOnce.Do(func() {
		file, err := os.Open(settingsPath)
		if err != nil {
			if err == os.ErrNotExist {
				settings = newSettings()
				return
			}
			panic(err)
		}
		defer file.Close() // Ignore error.
		data, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		s := newSettings()
		err = json.Unmarshal(data, s)
		if err != nil {
			panic(err)
		}
		settings = s
	})
}
