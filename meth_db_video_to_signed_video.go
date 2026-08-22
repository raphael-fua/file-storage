package main

import (
	"errors"
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}
	bucket, key, found := strings.Cut(*video.VideoURL, ",")
	if !found {
		return video, errors.New("was expecting a comma in video url")
	}

	url, err := generatePresignedURL(cfg.s3Client, bucket, key, time.Minute)
	if err != nil {
		return video, err
	}

	video.VideoURL = &url
	return video, nil
}
