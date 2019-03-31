Utilities
=============
The channel package contains several functions to deal with channels interfaces. It creates interface channels for reflect channels, bidirectional channels with input from a readonly, broadcasts one input channel to multiple outputs, combines multiple input channels to one output channel, publishes a fixed value to a channel when available and sends messages on these channels.

## External packages
|Package Name|Description|Reason|
|------------|-----------|------|
|[context](https://golang.org/x/net/context)|Goroutine context management (included in standard library from Golang 1.7)|To stop the goroutine if the context is no longer valid|
