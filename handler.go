package main

import (
	"os"
	"path/filepath"

	"github.com/donyori/gotfp"
)

// Ensure batchChan != nil.
func makeBatchHandler(batchChan chan<- *Batch, exitChan <-chan struct{},
	errChan chan<- error) gotfp.BatchHandler {
	lazyLoadSettings()
	lazyLoadPatternBatches()
	h := func(batch gotfp.Batch, depth int) (
		action gotfp.Action, skipDirs map[string]bool) {
		// Handling batch.Parent is necessary because of roots.
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
		if batch.Parent.Info != nil {
			if batch.Parent.Info.IsDir() {
				if skipPatternBatch.MatchDir(&batch.Parent, nil) {
					action = gotfp.ActionSkipDir
					return
				}
				if removePatternBatch.MatchDir(&batch.Parent, nil) {
					b := &Batch{
						Parent: filepath.Dir(batch.Parent.Path),
						Dirs:   []string{batch.Parent.Info.Name()},
					}
					batchChan <- b
					action = gotfp.ActionSkipDir
					return
				}
			} else {
				var doesRemove bool
				mode := batch.Parent.Info.Mode()
				if mode.IsRegular() {
					doesRemove = !skipPatternBatch.MatchRegFile(&batch.Parent, nil) &&
						removePatternBatch.MatchRegFile(&batch.Parent, nil)
				} else if (mode & os.ModeSymlink) != 0 {
					doesRemove = !skipPatternBatch.MatchSymlink(&batch.Parent, nil) &&
						removePatternBatch.MatchSymlink(&batch.Parent, nil)
				} else {
					doesRemove = !skipPatternBatch.MatchOther(&batch.Parent, nil) &&
						removePatternBatch.MatchOther(&batch.Parent, nil)
				}
				if doesRemove {
					b := &Batch{
						Parent: filepath.Dir(batch.Parent.Path),
						Dirs:   []string{batch.Parent.Info.Name()},
					}
					batchChan <- b
				}
				action = gotfp.ActionContinue
				return
			}
		}

		var b *Batch
		for i := range batch.Dirs {
			if skipPatternBatch.MatchDir(&batch.Dirs[i], &batch) {
				if skipDirs == nil {
					skipDirs = make(map[string]bool)
				}
				skipDirs[batch.Dirs[i].Path] = true
			} else if removePatternBatch.MatchDir(&batch.Dirs[i], &batch) {
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
			if !skipPatternBatch.MatchRegFile(&batch.RegFiles[i], &batch) &&
				removePatternBatch.MatchRegFile(&batch.RegFiles[i], &batch) {
				if b == nil {
					b = &Batch{Parent: batch.Parent.Path}
				}
				b.RegFiles = append(b.RegFiles, batch.RegFiles[i].Info.Name())
			}
		}
		for i := range batch.Symlinks {
			if !skipPatternBatch.MatchSymlink(&batch.Symlinks[i], &batch) &&
				removePatternBatch.MatchRegFile(&batch.Symlinks[i], &batch) {
				if b == nil {
					b = &Batch{Parent: batch.Parent.Path}
				}
				b.Symlinks = append(b.Symlinks, batch.Symlinks[i].Info.Name())
			}
		}
		for i := range batch.Others {
			if !skipPatternBatch.MatchOther(&batch.Others[i], &batch) &&
				removePatternBatch.MatchOther(&batch.Others[i], &batch) {
				if b == nil {
					b = &Batch{Parent: batch.Parent.Path}
				}
				b.Others = append(b.Others, batch.Others[i].Info.Name())
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
