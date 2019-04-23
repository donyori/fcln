package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/donyori/gotfp"
)

func getToRemove(roots ...string) (toRemove BatchList, err error) {
	lazyLoadSettings()
	workerNumber := settings.Worker.Number
	if workerNumber == 0 {
		panic(errors.New("fcln: worker number is 0"))
	}
	var bl BatchList
	batchChan := make(chan *Batch, workerNumber)
	exitChan := make(chan struct{})
	errChan := make(chan error, workerNumber)
	h := makeBatchHandler(batchChan, exitChan, errChan)
	doneChan := make(chan struct{})
	var workerErr, e error
	var b *Batch
	var ok bool
	go func() {
		defer close(doneChan)
		doesContinue := true
		for doesContinue {
			select {
			case b, ok = <-batchChan:
				if !ok {
					doesContinue = false
					break
				}
				if workerErr != nil {
					// Stop appending b to bl.
					break
				}
				bl = append(bl, b)
			case e = <-errChan:
				fmt.Fprintln(os.Stderr, time.Now(), e)
				if os.IsPermission(e) &&
					settings.PermissionErrorHandling != Fatal {
					break
				}
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
	gotfp.TraverseBatches(h, settings.Worker, errChan, roots...)
	close(batchChan)
	close(errChan)
	<-doneChan
	err = workerErr
	if err == nil {
		sort.Sort(bl)
		toRemove = bl
	}
	return
}

func printBatchList(bl BatchList) {
	bl.Print(os.Stdout)
}
