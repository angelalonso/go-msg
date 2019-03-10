package main

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	a = App{}
	a.Initialize()

	code := m.Run()

	os.Exit(code)
}

// Convert into master if no master host and port is given
// Connect to master host and port
// Disconnect
// Disconnect on <ctrl>+c
