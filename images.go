package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

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

func parseExif(exifData *exif.Exif) *OutputString {
	var data []string

	values := make(map[string]string)
	walker := &walker{values}
	_ = exifData.Walk(walker)
	for key, value := range values {
		data = append(data, key+": "+value)
	}
	sort.Strings(data)
	data = append([]string{"----Raw Values----"}, data...)
	datetime, err := exifData.DateTime()
	if err == nil {
		data = append([]string{"Date: " + datetime.String()}, data...)
	}
	lat, long, err := exifData.LatLong()
	if err == nil {
		data = append([]string{fmt.Sprintf("Maps Link: https://www.google.com/maps/search/?api=1&query=%f,%f", lat, long)}, data...)
	}
	comment, err := exifData.Get("usercomment")
	if err == nil {
		data = append([]string{"Comment: " + comment.String()}, data...)
	}
	result, _ := json.Marshal(data)
	return &OutputString{
		Success: true,
		Result:  string(result),
	}
}

type walker struct {
	values map[string]string
}

func (e *walker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	e.values[string(name)] = tag.String()
	return nil
}
