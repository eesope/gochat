package main

//As before, the user connects to the system via TCP and issues commands, with the supported commands being:

// • /NICK <nickname>
// • /LIST
// • /MSG <recipients> <message>
// These three commands have the same syntax and semantics as in assignment 1. Refer to the write-up of that assignment
// for details. Note that there are abbreviated forms of these commands.
// The design of the system is up to you. One possibility is to mimic the Elixir version. This means having a separate
// goroutine to act as the server responsible for handling nicknames and dispatching messages to their recipients. (This
// goroutine corresponds to the "chat server" in the Elixir version.) Then, for each external TCP client, the program creates
// two goroutines (corresponding to a proxy in the Elixir version) to handle requests. The chat server and the proxies will
// need to communicate in some way, for example, via channels. Although this design can be made to work, there is an
// alternative: since all Go routines created by the application run in the same process, it is possible to share data.
// Use as few top-level variables (variables declared outside functions) as possible. Your system should work with the
// Java/Elixir "dumb" client you implemented in assignment 1. Re-submit that Java/Elixir client, suitably modified if
// necessary. In addition, you'll also need to implement a Go client. Note that the chat program needs to remove the
// relevant nickname when a client terminates.
