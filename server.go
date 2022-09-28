package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type server struct {
	channels map[string]*channel
	commands chan command
}

func newServer() *server {
	return &server{
		channels: make(map[string]*channel),
		commands: make(chan command),
	}
}

func (s *server) run() {
	for cmd := range s.commands {
		switch cmd.id {
		case CMD_NICK:
			s.nick(cmd.client, cmd.args)
		case CMD_JOIN:
			s.join(cmd.client, cmd.args)
		case CMD_CHANNELS:
			s.listChannels(cmd.client)
		case CMD_MSG:
			s.msg(cmd.client, cmd.args)
		case CMD_QUIT:
			s.quit(cmd.client)
		case CMD_SEND_FILE:
			s.createEmptyFile(cmd.client, cmd.args)
		}
	}
}

func (s *server) newClient(conn net.Conn) *client {
	log.Printf("new client has joined: %s", conn.RemoteAddr().String())

	return &client{
		conn:     conn,
		nick:     "anonymous",
		commands: s.commands,
	}
}

func (s *server) nick(c *client, args []string) {
	log.Print("Test: ", args)

	if len(args) < 2 {
		c.msg("nick is required. usage: /nick NAME")
		return
	}

	c.nick = args[1]
	c.msg(fmt.Sprintf("all right, I will call you %s", c.nick))
}

func (s *server) join(c *client, args []string) {
	if len(args) < 2 {
		c.msg("channel name is required. usage: /join CHANNEL_NAME")
		return
	}

	channelName := args[1]

	r, ok := s.channels[channelName]
	if !ok {
		r = &channel{
			name:    channelName,
			members: make(map[net.Addr]*client),
		}
		s.channels[channelName] = r
	}
	r.members[c.conn.RemoteAddr()] = c

	s.quitCurrentChannel(c)
	c.channel = r

	r.broadcast(c, fmt.Sprintf("%s joined the channel", c.nick))

	c.msg(fmt.Sprintf("welcome to %s", channelName))
}

func (s *server) listChannels(c *client) {
	var channels []string
	for name := range s.channels {
		channels = append(channels, name)
	}

	c.msg(fmt.Sprintf("available channels: %s", strings.Join(channels, ", ")))
}

func (s *server) msg(c *client, args []string) {

	// Check if exist a channel if s.channels

	if len(args) < 2 {
		c.msg("message is required, usage: /msg MSG")
		return
	}

	msg := strings.Join(args[1:], " ")
	c.channel.broadcast(c, c.nick+": "+msg)
}

func (s *server) quit(c *client) {
	log.Printf("client has left the chat: %s", c.conn.RemoteAddr().String())

	s.quitCurrentChannel(c)

	c.msg("sad to see you go =(")
	c.conn.Close()
}

func (s *server) quitCurrentChannel(c *client) {
	if c.channel != nil {
		oldChannel := s.channels[c.channel.name]
		delete(s.channels[c.channel.name].members, c.conn.RemoteAddr())
		oldChannel.broadcast(c, fmt.Sprintf("%s has left the channel", c.nick))
	}
}

func (s *server) createEmptyFile(c *client, args []string) {

	log.Print("Test arguments: ")
	log.Print(args)

	if len(args) < 2 {

		c.msg("Name and content file is required, usage: /send_file [namefile] [content]")
		return

	}
	fileName := args[1]
	fileContent := strings.Join(args[2:], " ")

	myFile, err := os.Create(fileName)
	if err != nil {
		log.Fatal("ERROR! ", err)
		c.msg("Error")
	}
	log.Println("Empty file created successfully. ", myFile)
	log.Print("File created successfully")

	_, err2 := myFile.WriteString(fileContent)

	if err2 != nil {
		log.Fatal(err2)
	}

	fmt.Println("done")

	myFile.Close()


	c.channel.broadcast(c, c.nick + ": Sent you this file: " + fileName)
	c.channel.broadcast(c, "The file contains: " + fileContent)
}


