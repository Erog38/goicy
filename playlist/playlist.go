package playlist

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"git.philgore.net/CS497/Federation/Enterprise/config"
	"git.philgore.net/CS497/Federation/Enterprise/logger"
	pq "github.com/erog38/go-priority-queue"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

type Playlist struct {
	Playing   Track
	NextTrack Track
	Queue     pq.PriorityQueue
	db        *gorm.DB
}

func InitPlaylist() Playlist {
	pl := Playlist{}
	db, err := gorm.Open("sqlite3", config.Cfg.Playlist+"?_busy_timeout=5000")
	pl.db = db
	pl.Queue = pq.New()
	if err != nil {
		logger.Term("could not open database file"+err.Error(), logger.LOG_ERROR)
	}

	os.Mkdir("/dev/shm/goicy", 0660)

	pl.loadDB()
	go RunApi(&pl)
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

func (pl Playlist) First() string {
	t := pl.randomTrack()
	pl.Playing = t
	saveTrack("/dev/shm/goicy/current.mp3", t)
	go pl.saveNext()
	return "/dev/shm/goicy/current.mp3"
}

func (pl Playlist) saveNext() {
	var count int
	pl.db.Model(&Track{}).Count(&count)
	if count == 0 {
		logger.Term("No records in the database!", logger.LOG_ERROR)
	}

	var t Track
	if pl.Queue.Len() == 0 {
		t = pl.randomTrack()
	} else {
		tr, _ := pl.Queue.Pop()
		t = tr.(Track)
	}
	pl.NextTrack = t
	saveTrack("/dev/shm/goicy/next.mp3", t)
}

func saveTrack(fileName string, t Track) error {
	if t.TrackURL == "" {
		return errors.New("must have a track url in the track!")
	}
	trackURL := t.TrackURL
	trackResp, _ := http.Get(trackURL + "/download")
	defer trackResp.Body.Close()
	track, _ := ioutil.ReadAll(trackResp.Body)
	_ = ioutil.WriteFile(fileName, track, 0660)
	return nil
}

//next should:
// check if the "next.mp3" file exists
// if it doesn't, download one
// next point the
func (pl Playlist) Next() string {

	_, err := os.Stat("/dev/shm/goicy/next.mp3")

	if err == os.ErrNotExist {
		//file does not exist
		//buffer the next file
		pl.saveNext()
	} else if err != nil {
		logger.Log("Unknown file error "+err.Error(), logger.LOG_INFO)
	}
	//move the file from next to current
	//buffer the next file
	os.Rename("/dev/shm/goicy/next.mp3", "/dev/shm/goicy/current.mp3")
	pl.Playing = pl.NextTrack
	go pl.saveNext()
	return "/dev/shm/goicy/current.mp3"
}

func (pl Playlist) randomTrack() Track {

	var count int
	//Set the current track
	pl.db.Model(&Track{}).Count(&count)
	row := uint(count) % uint(rand.Uint64())
	var t Track
	pl.db.Where(Track{ID: row}).First(&t)
	pl.Playing = t
	return t
}

func (pl Playlist) Add(tID string) error {
	var t Track
	pl.db.Where(&Track{TrackID: tID}).First(&t)

	if t.ID == 0 {
		return errors.New("No track in DB with this ID!")
	}
	//Check if the playlist already has the file, if so increase it's priority,
	//else just add it to the end of the playlist.
	priority, contains := pl.Queue.Contains(t)
	if contains {
		pl.Queue.UpdatePriority(t, priority+1)
	} else {
		pl.Queue.Insert(t, 100)
	}
	return nil
}

func (pl Playlist) Close() {
	pl.db.Close()
}
