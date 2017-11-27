package playlist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"git.philgore.net/CS497/Federation/Enterprise/config"
	"git.philgore.net/CS497/Federation/Enterprise/logger"
	"github.com/gin-gonic/gin"
)

var g *gin.Engine

func RunApi() {
	gin.SetMode(gin.ReleaseMode)
	f, err := os.Create(config.Cfg.ApiLog)
	if err != nil {
		fmt.Println(err)
	}
	gin.DefaultWriter = io.MultiWriter(f)

	g = gin.Default()

	initRoutes()
	logger.Log("Starting api on port "+config.Cfg.ApiPort, logger.LOG_DEBUG)
	port := ":" + config.Cfg.ApiPort
	g.Run(port)
}

func initRoutes() {
	g.POST("/add", addHandler)
	//	g.POST("/remove", removeHandler)
	g.GET("/list", listHandler)
	//	g.GET("/album", albumHandler)
	//	g.Get("/track", trackHandler)
	g.GET("/current", currentHandler)
	g.GET("/history", historyHandler)
}

func listHandler(c *gin.Context) {
	p, _ := c.GetQuery("page")
	page, _ := strconv.Atoi(p)

	var resp ListResponse
	var albums []Album
	var count int
	pl.db.Table("albums").Count(&count)

	resp.Total = count
	pl.db.Offset(10 * page).Limit(10).Find(&albums)
	for _, a := range albums {

		var tracks []Track
		pl.db.Model(&Track{}).Where(&Track{AlbumID: a.AlbumID}).Find(&tracks)

		var apiTracks []ApiTrack

		for _, t := range tracks {
			apiTracks = append(apiTracks,
				ApiTrack{
					ID:       t.TrackID,
					Title:    t.TrackTitle,
					Artist:   t.ArtistName,
					Album:    t.AlbumTitle,
					Duration: t.TrackDuration,
					URL:      t.TrackURL,
					AlbumArt: t.TrackImageFile,
				})
		}

		resp.Albums = append(resp.Albums,
			ApiAlbum{
				ID:       a.AlbumID,
				Title:    a.AlbumTitle,
				Artist:   a.ArtistName,
				URL:      a.AlbumURL,
				AlbumArt: a.AlbumImageFile,
				Tracks:   apiTracks,
			})
	}
	resp.Success = true
	c.JSON(http.StatusOK, resp)
}

func addHandler(c *gin.Context) {

	var t Track
	var resp Response
	c.BindJSON(&t)

	if t.TrackID == "" {
		resp.Success = false
		resp.Err = "Missing the necessary Track ID!"
		str, _ := json.Marshal(resp)
		c.JSON(http.StatusBadRequest, str)
	}
	pl.mutex.Lock()
	err := Add(t.TrackID)
	pl.mutex.Unlock()
	if err != nil {
		resp.Success = false
		resp.Err = "Invalid Track ID!"
		str, _ := json.Marshal(resp)
		c.JSON(http.StatusBadRequest, str)
	}
	resp.Success = true
	str, _ := json.Marshal(resp)
	c.JSON(http.StatusOK, str)

}

func historyHandler(c *gin.Context) {
	var resp HistoryResponse
	pl.mutex.Lock()
	history := pl.History
	pl.mutex.Unlock()

	for _, t := range history {
		resp.History = append(resp.History,
			ApiTrack{
				ID:       t.TrackID,
				Title:    t.TrackTitle,
				Artist:   t.ArtistName,
				Album:    t.AlbumTitle,
				Duration: t.TrackDuration,
				URL:      t.TrackURL,
				AlbumArt: t.TrackImageFile,
			})
	}

	resp.Success = true
	c.JSON(http.StatusOK, resp)
}

func currentHandler(c *gin.Context) {
	var resp CurrentResponse
	pl.mutex.Lock()
	current := pl.History[len(pl.History)-1]
	pl.mutex.Unlock()
	if current.TrackID == "" {
		resp.Success = false
		resp.Err = "No content playing!"
		c.JSON(http.StatusNoContent, resp)
	}

	resp.Track = ApiTrack{
		ID:       current.TrackID,
		Title:    current.TrackTitle,
		Artist:   current.ArtistName,
		Album:    current.AlbumTitle,
		Duration: current.TrackDuration,
		URL:      current.TrackURL,
		AlbumArt: current.TrackImageFile,
	}

	resp.Success = true

	c.JSON(http.StatusOK, resp)
}
