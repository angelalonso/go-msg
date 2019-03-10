package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

func masterMsgManage(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{} // use default options
	var newUser user
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error upgrading connection:", err)
		return
	}
	userAddress := c.RemoteAddr()
	newUser.Name = userAddress.String()
	newUser.Address = userAddress
	newUser.Conn = c
	users = append(users, newUser)
	log.Printf("New User connected: %s (%d)", newUser.Name, len(users))
	defer c.Close()
	for {
		mt, rawMsgIn, err := c.ReadMessage()
		if err != nil {
			log.Printf("ERROR reading: %s", err)
			break
		}
		masterMsgForward(rawMsgIn, mt, c)
	}
}

func masterMsgForward(rawMsgOut []byte, mt int, c *websocket.Conn) {
	var msgOut message
	err_nonjson_msg := json.Unmarshal(rawMsgOut, &msgOut)
	if err_nonjson_msg != nil {
		log.Fatal("Error converting byte to json:", err_nonjson_msg)
	}
	if msgOut.To == "system" || msgOut.To == "" {
		msg2System(msgOut, c)
	} else {
		log.Printf("New message: %s", rawMsgOut)
		for u := range users {
			if users[u].Name == msgOut.To {
				masterMsgSend(msgOut.From, msgOut.Content, users[u].Conn)
			}
		}
	}
}

func msg2System(m message, c *websocket.Conn) {
	var msgFrom = "system"
	switch {
	case m.Content == "help":
		masterMsgSend(msgFrom, funcHelp(), c)
	case m.Content == "userlist":
		masterMsgSend(msgFrom, funcGetUserList(), c)
	case m.Content == "whoami":
		masterMsgSend(msgFrom, funcGetName(m), c)
	case strings.HasPrefix(m.Content, "name"):
		masterMsgSend(msgFrom, funcSetName(m, c), c)
	default:
		masterMsgSend(msgFrom, "\n        - Command not recognized:\n"+m.Content, c)
	}
}

func masterMsgSend(fr string, cnt string, c *websocket.Conn) {
	var msgOut = message{}
	msgOut.From = fr
	msgOut.Content = cnt
	json_msg, err_json_msg := json.Marshal(msgOut)
	if err_json_msg != nil {
		log.Printf("ERROR on building JSON: %s", err_json_msg)
	}
	err := c.WriteMessage(websocket.TextMessage, []byte(json_msg))
	if err != nil {
		log.Printf("ERROR on writing: %s", err)
		return
	}
}

// Returns an overview of the possible commands
func funcHelp() string {
	result := "\n        - Supported commands -\n" +
		"  \\help              shows this help\n" +
		"  \\userlist          lists current active users\n\n" +
		"  \\whoami            get users' own name\n\n" +
		"  \\name <name>       set users' own name\n\n" +
		"  @<user> <message>  sends message to user\n" +
		"  \\@<user>           sends all following messages to user\n"
	return result
}

func funcSetName(m message, c *websocket.Conn) string {
	result := "User Name could not be changed"
	//TODO: check that the name does not yet exist
	//TODO: inform any communications open of name change
	newName := strings.Split(m.Content, " ")[1]
	for u := range users {
		if users[u].Address == c.RemoteAddr() {
			users[u].Name = newName
			log.Printf("User %s changed name to %s", c.RemoteAddr().String(), users[u].Name)
			result = "\n        - Your username is now " + users[u].Name + " (was " + c.RemoteAddr().String() + ")"
		}
	}
	return result
}

// Returns the name of the user "calling" the function
func funcGetName(m message) string {
	result := "\n        - Your username is " + m.From
	return result
}

// Returns a list of the currently registered users
func funcGetUserList() string {
	var result []string
	for u := range users {
		result = append(result, users[u].Name)
	}
	return "\n        - Current active users are :\n" + strings.Join(result, "\n")
}
