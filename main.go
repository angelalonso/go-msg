package main

import (
	"net"
	"os"

	"github.com/gorilla/websocket"
)

type MsgApp struct {
	Mode string
}

type user struct {
	Name    string
	Address net.Addr
	Conn    *websocket.Conn
}

type message struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
}

var MASTERHOST = os.Getenv("GOMSG_MASTER_HOST")
var MASTERPORT = os.Getenv("GOMSG_MASTER_PORT")
var USER = os.Getenv("GOMSG_USER")
var prompt = "@" + userTo + ">>"
var userTo = "system"
var users []user

func main() {
	ma := MsgApp{}
	ma.Initialize()
}
