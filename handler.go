package main

import "github.com/donyori/gotfp"

func makeBatchHandler(toRemove *BatchList, exitChan <-chan struct{},
	errChan chan<- error) gotfp.BatchHandler {
	h := func(batch gotfp.Batch, depth int) (
		action gotfp.Action, skipDirs map[string]bool) {
		if batch.Parent.Err != nil {
			if errChan != nil {
				errChan <- batch.Parent.Err
			}
			return gotfp.ActionExit, nil
		}
		if len(batch.Errs) > 0 {
			if errChan != nil {
				for i := range batch.Errs {
					errChan <- batch.Errs[i].Err
				}
			}
			return gotfp.ActionExit, nil
		}
		select {
		case <-exitChan:
			return gotfp.ActionExit, nil
		default:
		}

		var b *Batch
		for i := range batch.Dirs {
			if removeDirFilter(batch.Dirs[i].Info) {
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
			if removeRegFileFilter(batch.RegFiles[i].Info) {
				if b == nil {
					b = &Batch{Parent: batch.Parent.Path}
				}
				b.RegFiles = append(b.RegFiles, batch.RegFiles[i].Info.Name())
			}
		}
		if b != nil {
			*toRemove = append(*toRemove, b)
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
