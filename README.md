### server structure 
|기능|설명|
|---|---|
TCP client | connect with proxies|
Goroutine | listening channel & receiving channel|
process command | /NICK, /LIST, /MSG |

### client feature
- connect to the system via TCP 
- issues commands:
    - /NICK <nickname> || /N <- erase if client terminate the program 
    - /LIST || /L
    - /MSG <recipients || *> <message> || /M

### how to test

```bash
# server start
go run main.go

# client start
elixir dumb_client.ex localhost 6666
```

### notes
- chat program has ping-pong-ing back and forth which can be simpler implement in channel & goroutine rather than shared data & mutex
