package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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
	fileExtensionString := ExtractFileExtension(mediaTypeString)

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not get video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "video user id does not match user id", err)
		return
	}

	byteSliceForFilePath := make([]byte, 32)
	_, err = rand.Read(byteSliceForFilePath)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"could not read byte slice",
			err,
		)
		return
	}

	filePathStringRelativeToAssets :=
		base64.RawURLEncoding.EncodeToString(byteSliceForFilePath) +
		"." +
		fileExtensionString

	filePathString := filepath.Join(cfg.assetsRoot, filePathStringRelativeToAssets)

	osFile, err := os.Create(filePathString)
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
		"http://localhost:%s/assets/%s",
		cfg.port,
		filePathStringRelativeToAssets,
	)

	video.ThumbnailURL = &thumbnailURLString
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}





