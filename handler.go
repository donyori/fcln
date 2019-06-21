package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/donyori/gotfp"
)

func MakeFileWithBatchHandler(infoChan chan<- *FileInfo,
	exitChan <-chan struct{}, errChan chan<- error) gotfp.FileWithBatchHandler {
	if infoChan == nil {
		panic(errors.New("fcln: infoChan is nil"))
	}
	LazyLoadSettings()
	LazyLoadPatternBatches()
	h := func(info gotfp.FileInfo, lctn *gotfp.LocationBatchInfo,
		depth int) (action gotfp.Action) {
		action = gotfp.ActionContinue
		select {
		case <-exitChan:
			return gotfp.ActionExit
		default:
		}
		if info.Cat == gotfp.ErrorFile {
			if os.IsPermission(info.Err) {
				switch settings.PermissionErrorHandling {
				case Ignore:
					// Skip this file, and all its contents if it is a directory.
					return gotfp.ActionSkip
				case Fatal:
					// Report error and exit.
					action = gotfp.ActionExit
				case Warn:
					// Report error and skip all its contents if it is a directory.
					fallthrough
				default:
					// Work as Warn.
					action = gotfp.ActionSkip
				}
			}
			if errChan != nil {
				errChan <- info.Err
			}
			return
		}
		if skipPatternBatch.Match(&info, lctn) {
			return gotfp.ActionSkip
		}
		if removePatternBatch.Match(&info, lctn) {
			dir, base := filepath.Split(info.Path)
			infoChan <- &FileInfo{
				Dir:  dir,
				Base: base,
				Cat:  info.Cat,
			}
			return gotfp.ActionSkip
		}
		return
	}
	return h
}
