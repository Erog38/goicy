package playlist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"

	"git.philgore.net/CS497/Federation/Enterprise/config"
	"git.philgore.net/CS497/Federation/Enterprise/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

type Playlist struct {
	Playing Track
	Queue   []Track
	db      *gorm.DB
}

func InitPlaylist() Playlist {
	pl := Playlist{}
	db, err := gorm.Open("sqlite3", config.Cfg.Playlist+"?_busy_timeout=5000")
	pl.db = db
	if err != nil {
		logger.Term("could not open database file"+err.Error(), logger.LOG_ERROR)
	}

	pl.loadDB()
	return pl
}

func (pl Playlist) loadDB() {
	fmt.Println("loading the DB")
	pl.db.AutoMigrate(&Album{})
	pl.db.AutoMigrate(&Track{})

	fmar := AlbumResponse{}
	for page := 0; page <= fmar.TotalPages; page++ {
		fmt.Println("Getting album list!")
		req := "https://freemusicarchive.org/api/get/albums.json?" +
			"api_key=" + config.Cfg.ApiKey + "&curator_handle=" + config.Cfg.Curator +
			"&page=" + strconv.Itoa(page)
		fmt.Println(req)
		resp, err := http.Get(req)
		if err != nil {
			logger.Log("could not get Free Music Archive information"+
				err.Error(), logger.LOG_ERROR)
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &fmar)
		if err != nil {
			logger.Log("could not unmarshal free music archive response"+
				err.Error(), logger.LOG_ERROR)
		}
		for _, album := range fmar.Albums {
			fmt.Println("Adding album..." + album.AlbumTitle)

			tr := TrackResponse{}
			for i := 0; i <= tr.TotalPages; i++ {
				r, e := http.Get("https://freemusicarchive.org/api/get/tracks.json?" +
					"api_key=" + config.Cfg.ApiKey + "&album_id=" + album.AlbumID +
					"&page=" + strconv.Itoa(i))
				if e != nil {
					logger.Log("could not unmarshal free music archive response"+
						err.Error(), logger.LOG_ERROR)
				}
				defer r.Body.Close()

				body, _ := ioutil.ReadAll(r.Body)
				err := json.Unmarshal(body, &tr)
				if err != nil {
					logger.Log("could not unmarshal free music archive response"+
						err.Error(), logger.LOG_ERROR)
				}
				for _, track := range tr.Tracks {
					fmt.Println("Adding track! " + track.TrackTitle)
					var count int
					pl.db.Model(&Track{}).Where(Track{TrackID: track.TrackID}).Count(&count)
					if count == 0 {
						pl.db.Create(&track)
					} else {
						pl.db.Where(Track{TrackID: track.TrackID}).Update(&track)
					}
				}
			}

			var count int
			pl.db.Model(Album{}).Where(Album{AlbumID: album.AlbumID}).Count(&count)
			if count == 0 {
				pl.db.Create(&album)
			} else {
				pl.db.Where(Album{AlbumID: album.AlbumID}).Update(&album)
			}
		}
	}

}

func (pl Playlist) First() {
	t = pl.randomTrack()
	return t
}

func (pl Playlist) Next() {

	var count int
	pl.db.Model(&Track{}).Count(&count)

	if count == 0 {
		logger.Term("No records in the database!", logger.LOG_ERROR)
	}
	if len(playlist) == 0 {
		t := pl.randomTrack()
		pl.Playing = t;
	} else {
		t := pl.Playlist[0]
		pl.Playlist = pl.Playlist[1:]
		pl.Playing = t
	}
}

func (pl Playlist) randomTrack() Track {

	var count int
	pl.db.Model(&Track{}).Count(&count)
	row := count % rand.Int()
	var t Track
	pl.db.Where(Track{ID: row}).First(&t)
	pl.Playing = t
	return t
}

func (pl Playlist) Add(tID string) {
	vat t Track
	pl.db.Where(&Track{TrackID: tID}).First(&t)
	//Check if the playlist already has the file, if so increase it's priority,
	//else just add it to the end of the playlist.
	pl.Playlist = append(pl.Playlist, t)
}

func (pl Playlist) Close() {
	pl.db.Close()
}
