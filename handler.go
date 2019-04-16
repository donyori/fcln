package main

import (
	"os"

	"github.com/donyori/gotfp"
)

// Ensure batchChan != nil.
func makeBatchHandler(batchChan chan<- *Batch, exitChan <-chan struct{},
	errChan chan<- error) gotfp.BatchHandler {
	lazyLoadSettings()
	h := func(batch gotfp.Batch, depth int) (
		action gotfp.Action, skipDirs map[string]bool) {
		// Handling batch.Parent is necessary because of roots.
		if batch.Parent.Info != nil {
			if batch.Parent.Info.IsDir() {
				if skipDirFilter(batch.Parent.Info) {
					action = gotfp.ActionSkipDir
					return
				}
			} else {
				action = gotfp.ActionContinue
				return
			}
		}
		if batch.Parent.Err != nil {
			if os.IsPermission(batch.Parent.Err) {
				switch settings.PermissionErrorHandling {
				case Ignore:
					// Skip all its contents.
					action = gotfp.ActionSkipDir
					return
				case Fatal:
					// Report error and exit.
					action = gotfp.ActionExit
				case Warn:
					// Report error and skip all its contents.
					fallthrough
				default:
					// Work as Warn.
					action = gotfp.ActionSkipDir
				}
			}
			if errChan != nil {
				errChan <- batch.Parent.Err
			}
			return
		}
		if len(batch.Errs) > 0 {
			for i := range batch.Errs {
				if os.IsPermission(batch.Errs[i].Err) {
					switch settings.PermissionErrorHandling {
					case Ignore:
						// Ignore this error.
						continue
					case Fatal:
						// Report this error and exit.
						action = gotfp.ActionExit
					case Warn:
						// Just report this error.
					default:
						// Work as Warn:
					}
				} else {
					action = gotfp.ActionExit
				}
				if errChan != nil {
					errChan <- batch.Errs[i].Err
				}
			} // End of for loop.
			if action == gotfp.ActionExit {
				return
			}
		}
		select {
		case <-exitChan:
			action = gotfp.ActionExit
			return
		default:
		}

		var b *Batch
		for i := range batch.Dirs {
			if skipDirFilter(batch.Dirs[i].Info) {
				if skipDirs == nil {
					skipDirs = make(map[string]bool)
				}
				skipDirs[batch.Dirs[i].Path] = true
			} else if removeDirFilter(batch.Dirs[i].Info) {
				if skipDirs == nil {
					skipDirs = make(map[string]bool)
				}
				skipDirs[batch.Dirs[i].Path] = true
				if b == nil {
					b = &Batch{Parent: batch.Parent.Path}
				}
				b.Dirs = append(b.Dirs, batch.Dirs[i].Info.Name())
			}
		}
		for i := range batch.RegFiles {
			if !skipRegFileFilter(batch.RegFiles[i].Info) &&
				removeRegFileFilter(batch.RegFiles[i].Info) {
				if b == nil {
					b = &Batch{Parent: batch.Parent.Path}
				}
				b.RegFiles = append(b.RegFiles, batch.RegFiles[i].Info.Name())
			}
		}
		if b != nil {
			batchChan <- b
		}

		if len(skipDirs) == 0 {
			action = gotfp.ActionContinue
		} else {
			action = gotfp.ActionSkipDir
		}
		return
	}
	return h
}
