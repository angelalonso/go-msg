package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestMain(t *testing.T) {
	// Create test server with the echo handler.
	s := httptest.NewServer(http.HandlerFunc(masterMsgManage))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// Send message to server, read response and check to see if it's what we expect.
	for i := 0; i < 10; i++ {
		//TODO: find a message that gives us a proper answer to start testing.
		//TODO:   or better yet, recode our application to make sense from a testing perspective
		if err := ws.WriteMessage(websocket.TextMessage, []byte("name testuser")); err != nil {
			t.Fatalf("%v", err)
		}
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p) != "hello" {
			fmt.Println(string(p))
			t.Fatalf("bad message")
		}
	}
}

// Convert into master if no master host and port is given
// Connect to master host and port
// Disconnect
// Disconnect on <ctrl>+c
