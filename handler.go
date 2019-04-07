package main

import (
	"os"

	"github.com/donyori/gotfp"
)

// Ensure batchChan != nil.
func makeBatchHandler(batchChan chan<- *Batch, exitChan <-chan struct{},
	errChan chan<- error, doesIgnorePermissionError bool) gotfp.BatchHandler {
	h := func(batch gotfp.Batch, depth int) (
		action gotfp.Action, skipDirs map[string]bool) {
		// Handling batch.Parent is necessary because of roots.
		if batch.Parent.Info != nil {
			if batch.Parent.Info.IsDir() {
				if skipDirFilter(batch.Parent.Info) {
					return gotfp.ActionSkipDir, nil
				}
			} else {
				return gotfp.ActionContinue, nil
			}
		}
		if batch.Parent.Err != nil {
			if doesIgnorePermissionError && os.IsPermission(batch.Parent.Err) {
				// Skip all its content.
				return gotfp.ActionSkipDir, nil
			}
			if errChan != nil {
				errChan <- batch.Parent.Err
			}
			return gotfp.ActionExit, nil
		}
		if len(batch.Errs) > 0 {
			for i := range batch.Errs {
				if doesIgnorePermissionError &&
					os.IsPermission(batch.Errs[i].Err) {
					continue
				}
				if errChan != nil {
					errChan <- batch.Errs[i].Err
				}
				action = gotfp.ActionExit
			}
			if action == gotfp.ActionExit {
				return
			}
		}
		select {
		case <-exitChan:
			return gotfp.ActionExit, nil
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
