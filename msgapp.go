package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

func (ma *MsgApp) Initialize() {
	// I use ADDRHOST to be able to use master and node on localhost for testing
	var ADDRHOST string
	if MASTERHOST == "" {
		ADDRHOST = "localhost"
	} else {
		ADDRHOST = MASTERHOST
	}
	var addr = flag.String("addr", ADDRHOST+":"+MASTERPORT, "http service address")
	if MASTERHOST == "" {
		ma.Mode = "master"
		runMaster(addr)
	} else {
		ma.Mode = "node"
		runNode(addr)
	}
}

func runMaster(addr *string) {
	log.Println("Running as MASTER")
	log.Println(MASTERPORT)
	flag.Parse()
	log.SetFlags(0)

	http.HandleFunc("/msg", masterMsgManage)
	log.Fatal(http.ListenAndServe(*addr, nil))

}

func runNode(addr *string) {
	log.Println("Running as NODE")
	if MASTERHOST == "" {
		MASTERHOST = "localhost"
	}
	if USER == "" {
		USER = getRandomName()
	}
	msgIn := make(chan message)
	msgOut := make(chan message)
	c := getConnection(addr)
	msgSend("system", "name "+USER, c)
	defer c.Close()
	fmt.Print(prompt)
	for {
		go channelReader(c, msgIn)
		go channelWriter(c, msgOut)
		select {
		case s1 := <-msgIn:
			msgInManage(s1)
		case s2 := <-msgOut:
			msgOutManage(s2, c)
		}
	}
}

func initMasterLog() {
	f, err := os.OpenFile("master.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
}

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

// ##################### NODE

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
