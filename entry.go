package main

import (
	"runtime"
	"sort"

	"github.com/donyori/gotfp"
)

func getToRemove(roots ...string) (toRemove BatchList, err error) {
	workerNumber := runtime.GOMAXPROCS(0) - 1
	if workerNumber <= 0 {
		workerNumber = 1
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
			if workerErr == nil {
				workerErr = e
			}
		}
	}()
	err = gotfp.TraverseBatches(h, workerNumber, errChan, 0, roots...)
	close(batchChan)
	close(errChan)
	<-doneChan
	if err == nil {
		err = workerErr
		if err == nil {
			sort.Sort(bl)
			toRemove = bl
		}
	}
	return
}
