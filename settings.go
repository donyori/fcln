package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"runtime"
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

func NewSettings() *Settings {
	settings := &Settings{
		Worker:                  *goctpf.NewWorkerSettings(),
		PermissionErrorHandling: Warn,
	}
	if settings.Worker.Number > 1 {
		settings.Worker.Number--
	}
	return settings
}

func LazyLoadSettings() {
	loadSettingsOnce.Do(func() {
		defer func() {
			if settings == nil || settings.Worker.Number != 0 {
				return
			}
			maxprocs := runtime.GOMAXPROCS(0)
			if maxprocs > 1 {
				settings.Worker.Number = uint32(maxprocs - 1)
			} else {
				settings.Worker.Number = 1
			}
		}()
		data, err := ioutil.ReadFile(settingsPath)
		if err != nil {
			if os.IsNotExist(err) {
				err = nil
				settings = NewSettings()
				return
			}
			panic(err)
		}
		s := NewSettings()
		err = json.Unmarshal(data, s)
		if err != nil {
			panic(err)
		}
		settings = s
	})
}
