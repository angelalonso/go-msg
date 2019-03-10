package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

func getRandomName() string {
	r_source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(r_source)

	adjFile, err := ioutil.ReadFile("files/generator_adjs")
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}
	adjStrings := strings.Split(string(adjFile), "\n")

	nameFile, err := ioutil.ReadFile("files/generator_names")
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}
	nameStrings := strings.Split(string(nameFile), "\n")

	r_name := r.Int()
	r_adj := r.Int()
	nameIndex := int(math.Mod(float64(r_name), float64(len(nameStrings))))
	adjIndex := int(math.Mod(float64(r_adj), float64(len(adjStrings))))

	return adjStrings[adjIndex] + "_" + nameStrings[nameIndex]
}
