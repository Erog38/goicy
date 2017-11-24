package playlist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"git.philgore.net/CS497/Federation/Enterprise/config"
	"github.com/gin-gonic/gin"
)

var g *gin.Engine
var pl *Playlist

func RunApi(playlist *Playlist) {

	pl = playlist
	gin.SetMode(gin.ReleaseMode)
	f, err := os.Create(config.Cfg.ApiLog)
	if err != nil {
		fmt.Println(err)
	}
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	g = gin.Default()

	initRoutes()

	g.Run(":" + config.Cfg.ApiPort)
}

func initRoutes() {
	g.POST("/add", addHandler)
	//	g.POST("/remove", removeHandler)
	g.GET("/current", currentHandler)
	//	g.GET("/playlist", playlistHandler)
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
	err := pl.Add(t.TrackID)
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

func currentHandler(c *gin.Context) {
	var resp CurrentResponse

	if pl.Playing.TrackID == "" {
		resp.Success = false
		resp.Err = "No content playing!"
		str, _ := json.Marshal(resp)
		c.JSON(http.StatusNoContent, str)
	}

	curr := pl.Playing

	resp.Track = ApiTrack{Title: curr.TrackTitle,
		Artist:   curr.ArtistName,
		Album:    curr.AlbumTitle,
		Duration: curr.TrackDuration,
		URL:      curr.TrackURL,
		AlbumArt: curr.TrackImageFile,
	}

	str, _ := json.Marshal(resp)
	c.JSON(http.StatusOK, str)
}
