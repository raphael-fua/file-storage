package main

import (
	"bytes"
	"encoding/json"
	// "errors"
	// "net/http"
	"os/exec"
)


type Stream struct {
	WidthsAndHeights []WidthAndHeight `json:"streams"`
}

type WidthAndHeight struct {
	Width int `json:"width"`
	Height int `json:"height"`
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v",
		"error",
		"-print_format",
		"json",
		"-show_streams",
		filePath,
	)

	buf := bytes.NewBuffer([]byte{})
	cmd.Stdout = buf

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	// stream := Stream {
	// 	WidthsAndHeights: []WidthAndHeight{},
	// }
	// var widthAndHeight []WidthAndHeight
	var stream Stream
	err = json.Unmarshal(buf.Bytes(), &stream)
	if err != nil {
		return "", err
	}

	if stream.WidthsAndHeights[0].Width < stream.WidthsAndHeights[0].Height {
		return "9:16", nil
	}
	return "16:9", nil
}
