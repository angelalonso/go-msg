package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func (ma *MsgApp) Initialize() {
	// I use ADDRHOST to be able to use master and node on localhost for testing
	var ADDRHOST string
	if MASTERHOST == "" {
		ADDRHOST = "localhost"
	} else {
		ADDRHOST = MASTERHOST
	}
	if MASTERPORT == "" {
		MASTERPORT = DEFAULTPORT
	}
	var addr = flag.String("addr", ADDRHOST+":"+MASTERPORT, "http service address")
	if MASTERHOST == "" {
		ma.Mode = "master"
		log.Println("Running as MASTER on Port " + MASTERPORT)
		go runMaster(addr)
		runNode(addr)
	} else {
		log.Println("Running as NODE")
		ma.Mode = "node"
		runNode(addr)
	}
}

func runMaster(addr *string) {
	//TODO: move the log init to a function (Write function needed?)
	_, err := os.Create("master.log")
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.OpenFile("master.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	flag.Parse()
	log.SetFlags(0)

	http.HandleFunc("/msg", masterMsgManage)
	log.Fatal(http.ListenAndServe(*addr, nil))

}

func runNode(addr *string) {
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
