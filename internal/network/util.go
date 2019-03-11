package network

import (
	"context"
	"reflect"
)

func reflect2chan(ctx context.Context, r reflect.Value, c chan<- interface{}) {
	for {
		//Get next value from channel
		if val, ok := r.Recv(); ok {
			select {
			case c <- val.Interface():
			case <-ctx.Done():
				return
			}
		} else {
			return
		}
	}
}
