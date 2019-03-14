package network

import (
	"context"
	"errors"
	"reflect"
)

func reflectchan2interfacechan(ctx context.Context, r reflect.Value) (<-chan interface{}, error) {
	if r.Kind() != reflect.Chan {
		return nil, errors.New("Input value is not a channel type")
	}
	c := make(chan interface{})
	go func() {
		done := false
		for !done {
			//Get next value from channel
			if val, ok := r.Recv(); ok {
				select {
				case c <- val.Interface():
				case <-ctx.Done():
					done = true
				}
			} else {
				done = true
			}
		}
		//clean up resources
		close(c)
	}()
	return c, nil
}

func rchan2rwchan(ctx context.Context, inChan <-chan interface{}) chan interface{} {
	outChan := make(chan interface{})
	go func() {
		running := true
		for running {
			select {
			case <-ctx.Done():
				running = false
			case m, ok := <-inChan:
				outChan <- m
				running = ok
			}
		}
		close(outChan)
	}()
	return outChan
}
