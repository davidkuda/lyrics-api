package main

import (
	"fmt"
)

type Songs []Song

type Song struct {
	Artist    string
	SongName  string
	SongText  string
	Chords    string
	Copyright string
	Covers    []string // list of URLs to great covers, e.g. on YouTube
}

func main() {
	fmt.Println("Hello World")
}
