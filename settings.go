package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/donyori/goctpf"
)

type Settings struct {
	Worker                  goctpf.WorkerSettings `json:"worker"`
	PermissionErrorHandling ErrorHandling         `json:"permission_error_handling,omitempty"`
}

var (
	settings         *Settings
	loadSettingsOnce sync.Once
)

func newSettings() *Settings {
	settings := &Settings{
		Worker:                  *goctpf.NewWorkerSettings(),
		PermissionErrorHandling: Warn,
	}
	if settings.Worker.Number > 1 {
		settings.Worker.Number--
	}
	return settings
}

func lazyLoadSettings() {
	loadSettingsOnce.Do(func() {
		data, err := ioutil.ReadFile(settingsPath)
		if err != nil {
			if os.IsNotExist(err) {
				settings = newSettings()
				return
			}
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
