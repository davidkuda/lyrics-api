package main

import (
	"fmt"
)

type Songs []Song

type Song struct {
	Artist    string `json:"artist"`
	SongName  string `json:"song_name"`
	SongText  string `json:"song_text,omitempty"`
	Chords    string `json:"chords,omitempty"`
	Copyright string `json:"copyright,omitempty"`
	// ? How to apply comments correctly?
	// Covers: list of URLs to great covers, e.g. on YouTube
	Covers []string `json:"covers,omitempty"`
}

func main() {
	fmt.Println("Hello World")
}
