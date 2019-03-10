package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
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
