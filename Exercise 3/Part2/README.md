# TCP

There are three common ways to format a message: (Though there are probably others)
 - 1: Always send fixed-sized messages
 - 2: Send the message size with each message (as part of a fixed-size header)
 - 3: Use some kind of marker to signify the end of a message

The TCP server can send you two of these:
 - Fixed size messages of size `1024`, if you connect to port `34933`
 - Delimited messages that use `\0` as the marker, if you connect to port `33546`

The server will read until it encounters the first `\0`, regardless.

Sharing a socket between threads should not be a problem, although reading from a socket in two threads will probably mean that only one of the threads get the message. If you are using blocking sockets, you could create a "receiving"-thread for each socket. Alternatively, you can use socket sets and the [`select()`](http://en.wikipedia.org/wiki/Select_%28Unix%29) function (or its equivalent).


## Connecting:
The IP address of the TCP server will be the same as the address the UDP server as spammed out on port 30000. Connect to the TCP server, using port `34933` for fixed-size messages, or port `33546` for 0-terminated messages. It will send you a welcome-message when you connect. 

The server will then echo anything you say to it back to you (as long as your message ends with `'\0'`). Try sending and receiving a few messages.

## Accepting connections:
 Tell the server to connect back to you, by sending a message of the form `Connect to: #.#.#.#:#\0` (IP of your machine and port you are listening to). You can find your own address by running `ifconfig` in the terminal.
 
 This new connection will behave the same way on the server-side, so you can send messages and receive echoes in the same way as before. You can even have it "Connect to" recursively (but please be nice... And no, the server will refuse requests to connect to itself).
