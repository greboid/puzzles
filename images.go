package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/simpicapp/goexif/exif"
	"github.com/simpicapp/goexif/tiff"
)

func getExifData(input io.Reader) (*exif.Exif, error) {
	exifData, err := exif.Decode(input)
	if err != nil {
		return nil, err
	}
	return exifData, nil
}

func parseExif(exifData *exif.Exif) []byte {
	results := &ExifResults{}

	values := make(map[string]string)
	walker := &walker{values}
	_ = exifData.Walk(walker)

	results.RawValues = &values

	datetime, err := exifData.DateTime()
	if err == nil {
		results.DateTime = datetime
	}

	lat, long, err := exifData.LatLong()
	if err == nil {
		results.MapLink = fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%f,%f", lat, long)
	}
	comment, err := exifData.Get("usercomment")
	if err == nil {
		results.Comment = comment.String()
	}
	resultsJson, _ := json.Marshal(results)
	result, _ := json.Marshal(&OutputString{
		Success: true,
		Result:  string(resultsJson),
	})
	return result
}

type walker struct {
	values map[string]string
}

func (e *walker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	e.values[string(name)] = tag.String()
	return nil
}

type ExifResults struct {
	MapLink   string             `json:"mapLink,omitempty"`
	DateTime  time.Time          `json:"datetime,omitempty"`
	Comment   string             `json:"comment,omitempty"`
	RawValues *map[string]string `json:"rawValues"`
}
