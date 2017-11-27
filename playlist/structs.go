package playlist

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

type FMAResponse struct {
	Errors     []string `json:"errors"`
	Limit      int      `json:"limit"`
	Message    string   `json:"message"`
	Page       int      `json:"page"`
	Title      string   `json:"title"`
	Total      string   `json:"total"`
	TotalPages int      `json:"total_pages"`
}

type AlbumDataset struct {
	Albums []Album `json:"dataset"`
}

type AlbumResponse struct {
	FMAResponse
	AlbumDataset
}

func (ar *FMAResponse) UnmarshalJSON(b []byte) error {
	tmp := make(map[string]interface{})
	json.Unmarshal(b, &tmp)
	for k, v := range tmp {
		switch k {
		case "errors":
			for _, err := range v.([]interface{}) {
				ar.Errors = append(ar.Errors, err.(string))
			}
			break
		case "limit":
			ar.Limit = int(v.(float64))
			break
		case "message":
			ar.Message = v.(string)
			break
		case "page":
			switch t := v.(type) {
			case int:
				ar.Page = t
			case string:
				ar.Page, _ = strconv.Atoi(t)
			}
			break
		case "title":
			ar.Title = v.(string)
			break
		case "total":
			ar.Total = v.(string)
			break
		case "total_pages":
			ar.TotalPages = int(v.(float64))
			break
		}
	}
	return nil
}

func (ar *AlbumResponse) UnmarshalJSON(b []byte) error {
	var fmar FMAResponse
	var ad AlbumDataset
	json.Unmarshal(b, &fmar)
	json.Unmarshal(b, &ad)
	ar.FMAResponse = fmar
	ar.AlbumDataset = ad

	return nil
}
func (ar *TrackResponse) UnmarshalJSON(b []byte) error {
	var fmar FMAResponse
	var td TrackDataset
	json.Unmarshal(b, &fmar)
	json.Unmarshal(b, &td)
	ar.FMAResponse = fmar
	ar.TrackDataset = td
	return nil
}

type TrackDataset struct {
	Tracks []Track `json:"dataset"`
}

type TrackResponse struct {
	FMAResponse
	TrackDataset
}

type TrackGenre struct {
	GenreID    string `json:"genre_id"`
	GenreTitle string `json:"genre_title"`
	GenreURL   string `json:"genre_url"`
}

type Track struct {
	ID                    uint `gorm:"primary_key"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             *time.Time `sql:"index"`
	AlbumID               string     `json:"album_id"`
	AlbumTitle            string     `json:"album_title"`
	AlbumURL              string     `json:"album_url"`
	ArtistID              string     `json:"artist_id"`
	ArtistName            string     `json:"artist_name"`
	ArtistURL             string     `json:"artist_url"`
	ArtistWebsite         string     `json:"artist_website"`
	LicenseImageFile      string     `json:"license_image_file"`
	LicenseImageFileLarge string     `json:"license_image_file_large"`
	LicenseParentID       string     `json:"license_parent_id"`
	LicenseTitle          string     `json:"license_title"`
	LicenseURL            string     `json:"license_url"`
	TrackBitRate          string     `json:"track_bit_rate"`
	TrackComments         string     `json:"track_comments"`
	TrackComposer         string     `json:"track_composer"`
	TrackCopyrightC       string     `json:"track_copyright_c"`
	TrackCopyrightP       string     `json:"track_copyright_p"`
	TrackDateCreated      string     `json:"track_date_created"`
	TrackDateRecorded     string     `json:"track_date_recorded"`
	TrackDiscNumber       string     `json:"track_disc_number"`
	TrackDuration         string     `json:"track_duration"`
	TrackExplicit         string     `json:"track_explicit"`
	TrackExplicitNotes    string     `json:"track_explicit_notes"`
	TrackFavorites        string     `json:"track_favorites"`
	TrackFile             string     `json:"track_file"`
	TrackID               string     `json:"track_id"`
	TrackImageFile        string     `json:"track_image_file"`
	TrackInformation      string     `json:"track_information"`
	TrackInstrumental     string     `json:"track_instrumental"`
	TrackInterest         string     `json:"track_interest"`
	TrackLanguageCode     string     `json:"track_language_code"`
	TrackListens          string     `json:"track_listens"`
	TrackLyricist         string     `json:"track_lyricist"`
	TrackNumber           string     `json:"track_number"`
	TrackPublisher        string     `json:"track_publisher"`
	TrackTitle            string     `json:"track_title"`
	TrackURL              string     `json:"track_url"`
}

type AlbumImages struct {
	AlbumID        string `json:"album_id"`
	ArtistID       string `json:"artist_id"`
	CuratorID      string `json:"curator_id"`
	ImageCaption   string `json:"image_caption"`
	ImageCopyright string `json:"image_copyright"`
	ImageFile      string `json:"image_file"`
	ImageID        string `json:"image_id"`
	ImageOrder     string `json:"image_order"`
	ImageSource    string `json:"image_source"`
	ImageTitle     string `json:"image_title"`
	UserID         string `json:"user_id"`
}

type Album struct {
	gorm.Model
	AlbumComments     string `json:"album_comments"`
	AlbumDateCreated  string `json:"album_date_created"`
	AlbumDateReleased string `json:"album_date_released"`
	AlbumEngineer     string `json:"album_engineer"`
	AlbumFavorites    string `json:"album_favorites"`
	AlbumHandle       string `json:"album_handle"`
	AlbumID           string `json:"album_id"`
	AlbumImageFile    string `json:"album_image_file"`
	AlbumInformation  string `json:"album_information"`
	AlbumListens      string `json:"album_listens"`
	AlbumProducer     string `json:"album_producer"`
	AlbumTitle        string `json:"album_title"`
	AlbumTracks       string `json:"album_tracks"`
	AlbumType         string `json:"album_type"`
	AlbumURL          string `json:"album_url"`
	ArtistName        string `json:"artist_name"`
	ArtistURL         string `json:"artist_url"`
}

//api

type ApiTrack struct {
	ID       string `json:"track_id"`
	Title    string `json:"title, omitempty"`
	Artist   string `json:"artist, omitempty"`
	Album    string `json:"album, omitempty"`
	Duration string `json:"duration, omitempty"`
	AlbumArt string `json:"album_art, omitempty"`
	URL      string `json:"url, omitempty"`
}

type ApiAlbum struct {
	ID       string     `json:"album_id"`
	Title    string     `json:"title, omitempty"`
	Artist   string     `json:"artist, omitempty"`
	AlbumArt string     `json:"album_art, omitempty"`
	Tracks   []ApiTrack `json:"tracks"`
	URL      string     `json:"url, omitempty"`
}

type PageOpts struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}

type Response struct {
	Success bool   `json:"success"`
	Err     string `json:"error"`
}

type ListResponse struct {
	Response
	Total  int        `json:"total"`
	Albums []ApiAlbum `json:"albums"`
}

type CurrentResponse struct {
	Response
	Track ApiTrack `json:"track"`
}

type HistoryResponse struct {
	Response
	History []ApiTrack `json:"history"`
}

type AddResponse struct {
	Response
}

type RemoveResponse struct {
	Response
}
