# go-msg
CLI Chat that runs in master-node mode, written in Go

# Requirements
go get github.com/gorilla/websocket

# Run as a master
go build && ./go-msg

# Run as a node
go build && GOMSG_MASTER_HOST=localhost GOMSG_MASTER_PORT=8040 ./go-msg

# KNOWN ISSUES
Loads. Currently working on:
- Disconnection is not handled
- Name change is not propagated
- No feedback after message is sent
