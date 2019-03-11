# UDP

We have set up a server on the real time lab that you're going to communicate with in this exercise. If you cannot connect to it, it might be down. Ask a student assistant to turn it on for you.

### Finding the server IP:
The server broadcasts it's own ip on port `30000`, listen for messages on this port to find the IP. You should write down the IP address as you will need it for later exercises.

### Sending UDP packets:
The server will respond to any message you send to it. Try sending a message to the server ip on port `20000 + n` where `n` is the number of the workspace you're sitting at. Listen to messages from the server and print them to a terminal to confirm that the messages are in fact beeing responded to.

#### Tips&Tricks
- The server will act the same way if you send a broadcast (`#.#.#.255`) instead of sending directly to the server.
- You are free to mess with the people around you by using the same port as them, but they may not appreciate it.
- It may be easier to use two sockets: one for sending and one for receiving.
- If you use broadcast, the messages will loop back to you! The server prefixes its reply with "You said: ", so you can tell if you are getting a reply from the server or if you are just listening to yourself.
