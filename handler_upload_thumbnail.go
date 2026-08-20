package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	// "thumbnail" should match the HTML form input name
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse from file", err)
		return
	}
	defer file.Close()

	mediaTypeString := header.Header.Get("Content-Type")
	words := strings.Split(mediaTypeString, "/")
	fileExtensionString := words[len(words) - 1]

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not get video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "video user id does not match user id", err)
		return
	}

	// `file` is an `io.Reader` that we can read from to get the image data
	// imageAsByteSlice, err := io.ReadAll(file)
	// if err != nil {
	// 	respondWithError(w, http.StatusBadRequest, "could not read thumbnail", err)
	// 	return
	// }

	filePath := filepath.Join(cfg.assetsRoot, videoIDString + "." + fileExtensionString)
    // fmt.Sprintf("assets/%s.%s", videoIDString, fileExtensionString)
	osFile, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not create file", err)
		return
	}
	_, err = io.Copy(osFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not copy to new file", err)
		return
	}

	thumbnailURLString := fmt.Sprintf(
		"http://localhost:%s/assets/%s.%s",
		cfg.port,
		videoIDString,
		fileExtensionString,
	)

	// imageSQLFormat := base64.StdEncoding.EncodeToString(imageAsByteSlice)
	// dataURL := fmt.Sprintf("data:%s;base64,%s", mediaTypeString, imageSQLFormat)

	video.ThumbnailURL = &thumbnailURLString
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}





