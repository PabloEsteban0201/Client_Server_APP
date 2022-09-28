package main

import (
	"net"
)

type channel struct {
	name    string
	members map[net.Addr]*client
}

func (r *channel) broadcast(sender *client, msg string) {
	for addr, m := range r.members {
		if sender.conn.RemoteAddr() != addr {
			m.msg(msg)
		}
	}
}