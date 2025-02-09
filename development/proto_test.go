package main

import (
	"fmt"
	"testing"
)

func TestProto(t *testing.T) {

	m := make(map[string][]byte)

	m["shubh"] = []byte("pandey")
	m["shubham"] = []byte("pandey")
	m["swami"] = []byte("pandey")

	bt := WriteJson(m)

	fmt.Println(string(bt))
}
