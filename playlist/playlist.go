package playlist

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.philgore.net/CS497/Federation/Enterprise/config"
	"git.philgore.net/CS497/Federation/Enterprise/logger"
	pq "github.com/erog38/go-priority-queue"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

type Playlist struct {
	NextTrack Track
	History   []Track
	Queue     pq.PriorityQueue
	db        *gorm.DB
	mutex     *sync.Mutex
}

var pl *Playlist

func InitPlaylist() {
	pl = &Playlist{}
	pl.mutex = &sync.Mutex{}
	db, err := gorm.Open("sqlite3", config.Cfg.Playlist+"?_busy_timeout=5000")
	logger.Log("Opening database", logger.LOG_DEBUG)
	pl.db = db
	pl.Queue = pq.New()
	if err != nil {
		logger.Term("could not open database file"+err.Error(), logger.LOG_ERROR)
	}

	_, err = os.Stat("/dev/shm/goicy")
	if os.IsNotExist(err) {
		os.Mkdir("/dev/shm/goicy", os.ModePerm)
	} else if err != nil {
		logger.TermLn("Could not create /dev/shm/goicy "+err.Error(), logger.LOG_ERROR)
	} else {
		os.RemoveAll("/dev/shm/goicy")
		os.Mkdir("/dev/shm/goicy", os.ModePerm)
	}
	if config.Cfg.ReloadDB {
		loadDB()
	}
	rand.Seed(time.Now().UTC().UnixNano())
	go RunApi()
}

func getAlbumList(page int, fmar *AlbumResponse) {
	req := "https://freemusicarchive.org/api/get/albums.json?" +
		"api_key=" + config.Cfg.ApiKey + "&curator_handle=" + config.Cfg.Curator +
		"&page=" + strconv.Itoa(page)
	resp, err := http.Get(req)

	if err != nil {
		logger.Log("could not get Free Music Archive information "+
			err.Error(), logger.LOG_ERROR)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &fmar)
	if err != nil {
		logger.Log("could not unmarshal free music archive response "+
			err.Error(), logger.LOG_ERROR)
	}
}

func loadDB() {
	pl.db.AutoMigrate(Album{})
	pl.db.AutoMigrate(Track{})
	logger.Log("Loading the database with albums and tracks...", logger.LOG_INFO)
	fmar := AlbumResponse{}
	for page := 0; page <= fmar.TotalPages; page++ {
		getAlbumList(page, &fmar)
		for _, album := range fmar.Albums {

			tr := TrackResponse{}
			logger.Log("Downloading Album "+album.AlbumTitle+
				" from the Free Music Archive.", logger.LOG_INFO)

			for i := 0; i <= tr.TotalPages; i++ {
				r, e := http.Get("https://freemusicarchive.org/api/get/tracks.json?" +
					"api_key=" + config.Cfg.ApiKey + "&album_id=" + album.AlbumID +
					"&page=" + strconv.Itoa(i))
				if e != nil {
					logger.Log("could not unmarshal free music archive response"+
						e.Error(), logger.LOG_ERROR)
				}
				defer r.Body.Close()

				body, _ := ioutil.ReadAll(r.Body)
				err := json.Unmarshal(body, &tr)
				if err != nil {
					logger.Log("could not unmarshal free music archive response"+
						err.Error(), logger.LOG_ERROR)
				}
				for _, track := range tr.Tracks {
					if track.TrackURL != "" {
						pl.db.Where(Track{TrackID: track.TrackID}).Assign(&track).FirstOrCreate(&Track{})
					}
				}

			}

			pl.db.Where(Album{AlbumID: album.AlbumID}).Assign(&album).FirstOrCreate(&Album{})
		}
	}

}

func appendHistory(t Track) {

	pl.mutex.Lock()
	pl.History = append(pl.History, t)
	if len(pl.History) > 10 {
		pl.History = pl.History[len(pl.History)-10:]
	}
	pl.mutex.Unlock()
}

func First() string {
	t := randomTrack()
	appendHistory(t)
	saveTrack("/dev/shm/goicy/current.mp3", t)
	saveNext()
	return "/dev/shm/goicy/current.mp3"
}

func saveNext() {
	var count int
	pl.db.Model(&Track{}).Count(&count)
	if count == 0 {
		logger.Term("No records in the database!", logger.LOG_ERROR)
	}

	var t Track
	if pl.Queue.Len() == 0 {
		t = randomTrack()
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
	trackURL := strings.TrimSpace(t.TrackURL) + "/download"
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
func Next() string {

	_, err := os.Stat("/dev/shm/goicy/next.mp3")

	if os.IsNotExist(err) {
		//file does not exist
		//buffer the next file
		saveNext()
	} else if err != nil {
		logger.Log("Unknown file error "+err.Error(), logger.LOG_INFO)
	}
	//move the file from next to current
	//buffer the next file
	os.Rename("/dev/shm/goicy/next.mp3", "/dev/shm/goicy/current.mp3")
	appendHistory(pl.NextTrack)
	go saveNext()
	return "/dev/shm/goicy/current.mp3"
}

func randomTrack() Track {

	var count int
	//Set the current track
	pl.db.Model(&Track{}).Count(&count)
	row := uint(rand.Uint64()) % uint(count)
	var t Track
	pl.db.Where(Track{ID: row}).First(&t)
	return t
}

func Add(tID string) error {
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

func Close() {
	pl.db.Close()
}
