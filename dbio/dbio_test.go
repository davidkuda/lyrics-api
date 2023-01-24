package dbio

import (
	"testing"
)

func TestGetSong(t *testing.T) {
	song := GetSong("Start Me Up")
	if song.Artist != "The Rolling Stones" {
		t.Errorf("Failed to fetch the song \"Start Me Up\", received %v", song.SongName)
	}
}
