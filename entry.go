package main

import (
	"runtime"
	"sort"

	"github.com/donyori/gotfp"
)

func getToRemove(roots ...string) (toRemove BatchList, err error) {
	workerNumber := runtime.GOMAXPROCS(0)
	var bl BatchList
	exitChan := make(chan struct{})
	errChan := make(chan error, workerNumber)
	h := makeBatchHandler(&bl, exitChan, errChan)
	doneChan := make(chan struct{})
	var workerErr error
	go func() {
		defer close(doneChan)
		e, ok := <-errChan
		if !ok {
			return
		}
		workerErr = e
		close(exitChan)
		for range errChan {
			// Drain errChan.
		}
	}()
	err = gotfp.TraverseBatches(h, workerNumber, errChan, 0, roots...)
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
