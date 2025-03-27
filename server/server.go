package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type Command interface{} // to manage client commands in chatserver

type SetNick struct {
	ClientID string // client identifier
	Nick     string
	Reply    chan Response // channel to receive response after process command
}

type List struct {
	Reply chan Response
}

type Msg struct {
	Sender     string // nick name
	Recipients string // * || string list
	Message    string
	Reply      chan Response
}

type Client struct {
	Nickname string
	Conn     net.Conn // tcp connection info
}

var (
	clientMap = make(map[string]*Client) // to update nickname
	mutex     sync.Mutex
)

type Response struct {
	Success bool
	Message string
}

type ChatServer struct {
	commands chan Command // channel for getting client command
	clients  map[string]net.Conn // key: nickname - val: TCP conn of each client
	mu       sync.Mutex          // to protect client so that not mixed up channel
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		commands: make(chan Command),
		clients:  make(map[string]net.Conn),
	}
}

// ---------------------------------------
// main server
// ---------------------------------------
// drive the program
func (s *ChatServer) Start() {
	for cmd := range s.commands {
		switch c := cmd.(type) {
		case SetNick:
			s.handleSetNick(c)

		case List:
			s.mu.Lock()
			var nicks []string
			for nick := range s.clients {
				nicks = append(nicks, nick)
			}
			s.mu.Unlock()
			c.Reply <- Response{Success: true, Message: "Users: " + strings.Join(nicks, ", ")}
		
		case Msg:
			s.mu.Lock()
			if c.Recipients == "*" {
				// everyone but not sender
				for nick, conn := range s.clients {
					if nick != c.Sender && conn != nil {
						conn.Write([]byte(fmt.Sprintf("[%s]: %s\n", c.Sender, c.Message)))
					}
				}
			} else {
				recipients := strings.Split(c.Recipients, ",")
				for _, nick := range recipients {
					nick = strings.TrimSpace(nick)
					if conn, exists := s.clients[nick]; exists && conn != nil {
						conn.Write([]byte(fmt.Sprintf("[%s]: %s\n", c.Sender, c.Message)))
					}
				}
			}
			s.mu.Unlock()
			c.Reply <- Response{Success: true, Message: "Message sent"}
		}
	}
}

func (s *ChatServer) handleSetNick(c SetNick) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// check 2 condition:
	// 1. existingConn, exists := s.clients[c.Nick] -> map lookup -> true if exists && assign that value to existingConn
	// 2. if ^ is true, once again check existingConn != nil

	if existingConn, exists := s.clients[c.Nick]; exists && existingConn != nil {
		c.Reply <- Response{Success: false, Message: "Nickname already in use"}
		return
	}

	// set oldNick to manage if exist
	var oldNick string
	for nick, conn := range s.clients {
		if conn != nil && conn.RemoteAddr().String() == c.ClientID {
			oldNick = nick
			break
		}
	}

	if oldNick != "" {
		delete(s.clients, oldNick)
		fmt.Printf("Client %s: Nickname updated from %s to %s\n", c.ClientID, oldNick, c.Nick)
	} else {
		fmt.Printf("Client %s: Nickname set to %s\n", c.ClientID, c.Nick)
	}

	c.Reply <- Response{Success: true, Message: "Nickname set to " + c.Nick}
}

// proxy offers connection for each client
func (s *ChatServer) RegisterClient(nick string, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[nick] = conn
}

// ---------------------------------------
// TCP connection: proxy
// ---------------------------------------

func main() {
	// start server; new goroutine
	server := NewChatServer()
	go server.Start()

	ln, err := net.Listen("tcp", ":6666")
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
		os.Exit(1)
	}
	defer ln.Close()

	fmt.Println("Go Chat Server started on port 6666...")

	// each client has goroutine
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go handleClient(conn, server)
	}
}

// proxy; conn TCP
func handleClient(conn net.Conn, server *ChatServer) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	var myNick string

	// each client has channel
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		input = strings.TrimSpace(input)
		parts := strings.Fields(input)
		if len(parts) == 0 {
			conn.Write([]byte("Invalid command\n"))
			continue
		}

		switch parts[0] {
		case "/NICK", "/N":
			if len(parts) < 2 {
				conn.Write([]byte("Usage: /NICK <nickname>\n"))
				continue
			}
			nick := parts[1]
			replyCh := make(chan Response)
			// send SetNick command to server
			server.commands <- SetNick{ClientID: conn.RemoteAddr().String(), Nick: nick, Reply: replyCh}
			resp := <-replyCh
			conn.Write([]byte(resp.Message + "\n"))
			if resp.Success {
				myNick = nick
				// proxy -(client data)-> server
				server.RegisterClient(nick, conn)
			}

		case "/LIST", "/L":
			replyCh := make(chan Response)
			server.commands <- List{Reply: replyCh}
			resp := <-replyCh
			conn.Write([]byte(resp.Message + "\n"))

		case "/MSG", "/M":
			if myNick == "" {
				conn.Write([]byte("Set a nickname first using /NICK\n"))
				continue
			}
			if len(parts) < 3 {
				conn.Write([]byte("Usage: /MSG <recipient(s)> <message>\n"))
				continue
			}
			// parts[1] recipient, parts[2:] message
			targets := parts[1]
			message := strings.Join(parts[2:], " ")
			replyCh := make(chan Response)
			server.commands <- Msg{Sender: myNick, Recipients: targets, Message: message, Reply: replyCh}
			resp := <-replyCh
			conn.Write([]byte(resp.Message + "\n"))
		default:
			conn.Write([]byte("Unknown command\n"))
		}
	}
}
