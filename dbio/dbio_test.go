package dbio

import (
	"testing"
)

func TestListSong(t *testing.T) {
	songs := ListSongs()
	expected := []string{
		"Pink Floyd -- Wish You Were Here",
		"Sting -- Englishman In New York",
		"The Rolling Stones -- Start Me Up",
	}
	same := true
	for i, song := range songs {
		if song != expected[i] {
			same = false
			break
		}
	}
	if !same {
		t.Errorf("Wrong results;\n  expected: %v\n  got %v", expected, songs)
	}
}

func TestGetSong(t *testing.T) {
	song := GetSong("Start Me Up")
	if song.Artist != "The Rolling Stones" {
		t.Errorf("Failed to fetch the song \"Start Me Up\", received %v", song.SongName)
	}
}
