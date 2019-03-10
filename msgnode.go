package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

func getConnection(addr *string) *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/msg"}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	return c
}

func channelReader(c *websocket.Conn, m chan message) {
	var msgIn = message{}
	_, rcv_message, _ := c.ReadMessage()
	err_nonjson_msg := json.Unmarshal(rcv_message, &msgIn)
	if err_nonjson_msg != nil {
		log.Printf("ERROR on reading JSON: %s", err_nonjson_msg)
	}
	m <- msgIn
}

// Reads any messages that come from stdin
func channelWriter(c *websocket.Conn, m chan message) {
	var msgOut = message{}
	var ioReader = bufio.NewReader(os.Stdin)
	if content, _ := ioReader.ReadString('\n'); content != "\n" {
		msgOut.Content = strings.TrimSpace(content)
		msgOut.From = USER
		msgOut.To = userTo
	}
	m <- msgOut
}

// Shows the messages coming from the connection
func msgInManage(m message) {
	fmt.Print("\r\r ")
	fmt.Print(m.From)
	fmt.Print(" says: ")
	fmt.Println(m.Content)
	fmt.Print(prompt)
}

// Prepares the messages coming from stdin
// - if they start with \, they are treated as messages to the system
//   - some of these are handled here and some others at the server
// - if a message starts with @, it marks the user the message goes to
// - otherwise it is sent to the server
func msgOutManage(m message, c *websocket.Conn) {
	if len(m.Content) > 0 {
		if m.Content[0] == '\u005c' {
			if m.Content[1] == '@' {
				userTo = strings.Split(strings.Split(m.Content, " ")[0], "@")[1]
				prompt = "@" + userTo + ">>"
			} else {
				t := "system"
				cnt := strings.Split(m.Content, "\\")[1]
				msgSend(t, cnt, c)
			}
		} else if m.Content[0] == '@' {
			t := strings.Split(strings.Split(m.Content, " ")[0], "@")[1]
			cnt := strings.Replace(m.Content, "@"+t+" ", "", 1)
			msgSend(t, cnt, c)
		} else {
			msgSend(m.To, m.Content, c)
		}
	}
	fmt.Print(prompt)
}

// Fills up all fields on the message struct and sends the message
func msgSend(to string, cnt string, c *websocket.Conn) {
	var msgOut = message{}
	msgOut.From = USER
	msgOut.To = to
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
