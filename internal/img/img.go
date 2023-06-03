package img

import (
	"bytes"
	"encoding/base64"
	"fmt"
	log "github.com/sirupsen/logrus"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"
)

const (
	imgTypePng  = "png"
	imgTypeJpeg = "jpeg"
	imgTypeGif  = "gif"
)

func ParseB64(data string) (image.Image, error) {
	idx := strings.Index(data, ";base64,")
	if idx < 0 {
		return nil, fmt.Errorf("no base64 substring found")
	}
	imageType := data[11:idx]
	log.Infof("[IMG] detected image of %s", imageType)

	decoded, err := base64.StdEncoding.DecodeString(data[idx+8:])
	if err != nil {
		return nil, fmt.Errorf("decode b64 failed: %w", err)
	}

	r := bytes.NewReader(decoded)
	switch imageType {
	case imgTypePng:
		im, err := png.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("png parse failed: %w", err)
		}

		return im, nil
	case imgTypeJpeg:
		im, err := jpeg.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("jpeg parse failed: %w", err)
		}
		return im, nil
	case imgTypeGif:
		im, err := gif.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("gif parse failed: %w", err)
		}
		return im, nil
	default:
		return nil, fmt.Errorf("unknown type %s", imageType)
	}
}
