package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/donyori/gotfp"
)

func GetToBeRemovedFiles(roots ...string) (toRemove BatchList, err error) {
	LazyLoadSettings()
	if len(roots) == 0 {
		return
	}
	infosMap := make(map[string][]*FileInfo) // "s" stands for "slice".
	infoChan := make(chan *FileInfo, settings.Worker.Number)
	exitChan := make(chan struct{})
	errChan := make(chan error, settings.Worker.Number)
	h := MakeFileWithBatchHandler(infoChan, exitChan, errChan)
	doneChan := make(chan struct{})
	var workerErr, e error
	var info *FileInfo
	var ok bool
	go func() {
		defer close(doneChan)
		doesContinue := true
		for doesContinue {
			select {
			case info, ok = <-infoChan:
				if !ok {
					doesContinue = false
					break
				}
				if workerErr != nil {
					// Stop adding info to infosMap.
					break
				}
				// info cannot be nil.
				infosMap[info.Dir] = append(infosMap[info.Dir], info)
			case e = <-errChan:
				fmt.Fprintln(os.Stderr, time.Now(), e)
				if os.IsPermission(e) &&
					settings.PermissionErrorHandling != Fatal {
					break
				}
				// To make sure to close exitChan and set workerErr only once.
				if exitChan == nil {
					break
				}
				close(exitChan)
				exitChan = nil
				workerErr = e
			}
		}
		for e = range errChan {
			// Drain errChan.
			fmt.Fprintln(os.Stderr, time.Now(), e)
			if workerErr == nil && (!os.IsPermission(e) ||
				settings.PermissionErrorHandling == Fatal) {
				workerErr = e
			}
		}
	}()
	gotfp.TraverseFilesWithBatch(h, settings.Worker, errChan, roots...)
	close(infoChan)
	close(errChan)
	<-doneChan
	err = workerErr
	if err != nil {
		return
	}
	toRemove = make([]*Batch, 0, len(infosMap))
	for dir, infos := range infosMap {
		if len(infos) == 0 {
			continue
		}
		batch := &Batch{Parent: dir}
		for _, info := range infos {
			switch info.Cat {
			case gotfp.Directory:
				batch.Dirs = append(batch.Dirs, info.Base)
			case gotfp.RegularFile:
				batch.RegFiles = append(batch.RegFiles, info.Base)
			case gotfp.Symlink:
				batch.Symlinks = append(batch.Symlinks, info.Base)
			case gotfp.OtherFile:
				batch.Others = append(batch.Others, info.Base)
			default:
				panic(fmt.Errorf(
					"fcln: got file category %q in BatchList, which should NOT occur",
					info.Cat))
			}
		}
		if len(batch.Dirs) > 0 {
			sort.Strings(batch.Dirs)
		}
		if len(batch.RegFiles) > 0 {
			sort.Strings(batch.RegFiles)
		}
		if len(batch.Symlinks) > 0 {
			sort.Strings(batch.Symlinks)
		}
		if len(batch.Others) > 0 {
			sort.Strings(batch.Others)
		}
		toRemove = append(toRemove, batch)
	}
	sort.Sort(toRemove)
	return
}

func PrintBatchList(bl BatchList) {
	bl.Print(os.Stdout)
}
