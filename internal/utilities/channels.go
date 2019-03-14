package utilities

import (
	"context"
	"errors"
	"log"
	"reflect"
)

//ReflectChan2InterfaceChan creates a interface channel for any reflect channel
//Panics if types is incompatible
func ReflectChan2InterfaceChan(ctx context.Context, r reflect.Value) (<-chan interface{}, error) {
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

//RChan2RWChan creates a new bidirectional channel and use input from a readonly
func RChan2RWChan(ctx context.Context, inChan <-chan interface{}) chan interface{} {
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

//OneToManyChan broadcast one input chan to multiple outputs
//Panics if types do not match
func OneToManyChan(ctx context.Context, in interface{}, out ...interface{}) {
	T := reflect.TypeOf(in).Elem()
	reflectChans := make([]reflect.Value, len(out))
	for i, c := range out {
		if reflect.TypeOf(c).Elem() != T {
			log.Panicln("Incompatible channel types")
		}
		reflectChans[i] = reflect.ValueOf(c)
	}
	selectCases := make([]reflect.SelectCase, 2)
	for {
		selectCases[0] = reflect.SelectCase{
			Chan: reflect.ValueOf(ctx.Done()),
			Dir:  reflect.SelectRecv,
		}

		selectCases[1] = reflect.SelectCase{
			Chan: reflect.ValueOf(in),
			Dir:  reflect.SelectRecv,
		}
		i, val, ok := reflect.Select(selectCases)
		if i == 0 || !ok {
			return
		}

		for _, c := range reflectChans {
			c.Send(val)
		}
	}
}

//ManyToOneChan combines multiple input channels into one output channel
//Panics if type do not match
func ManyToOneChan(ctx context.Context, out interface{}, in ...interface{}) {
	T := reflect.TypeOf(in).Elem()
	selectCases := make([]reflect.SelectCase, len(in)+1)
	outChan := reflect.ValueOf(out)
	for i, c := range in {
		if reflect.TypeOf(c).Elem() != T {
			log.Panicln("Incompatible channel types")
		}
		selectCases[i+1] = reflect.SelectCase{
			Chan: reflect.ValueOf(c),
			Dir:  reflect.SelectRecv,
		}
	}
	for {
		selectCases[0] = reflect.SelectCase{
			Chan: reflect.ValueOf(ctx.Done()),
			Dir:  reflect.SelectRecv,
		}
		i, val, ok := reflect.Select(selectCases)
		if i == 0 || !ok {
			return
		}
		outChan.Send(val)
	}
}
