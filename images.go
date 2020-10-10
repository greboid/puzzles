package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"time"

	_ "github.com/oov/psd"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vector"
	_ "golang.org/x/image/vp8"
	_ "golang.org/x/image/vp8l"
	_ "golang.org/x/image/webp"

	"github.com/simpicapp/goexif/exif"
	"github.com/simpicapp/goexif/tiff"
)

func getExifData(data []byte) (*exif.Exif, error) {
	reader := bytes.NewReader(data)
	exifData, err := exif.Decode(reader)
	if err != nil {
		return nil, err
	}
	return exifData, nil
}

func parseExif(exifData *exif.Exif) *ExifResults {
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
	return results
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

type ImageInfo struct {
	Width     int          `json:"width"`
	Height    int          `json:"height"`
	ImageType string       `json:"type"`
	ExifData  *ExifResults `json:"exifData"`
}

func getImageSize(data []byte) (ImageInfo, error) {
	reader := bytes.NewReader(data)
	config, imageType, err := image.DecodeConfig(reader)
	if err != nil {
		return ImageInfo{}, err
	}
	imageInfo := ImageInfo{
		Width:     config.Width,
		Height:    config.Height,
		ImageType: imageType,
	}
	return imageInfo, err
}

func getImageInfo(data []byte) (ImageInfo, error) {
	imageInfo, err := getImageSize(data)
	if err != nil {
		return ImageInfo{}, err
	}
	exifData, err := getExifData(data)
	if err == nil {
		imageInfo.ExifData = parseExif(exifData)
		return imageInfo, err
	} else {
		imageInfo.ExifData = &ExifResults{}
	}
	return imageInfo, nil
}
