package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
    const fileNameWithoutExtension = "tubely-upload"
	const uploadLimitInt = 1 << 30 // 1GB
	mimeTypeString := "video/mp4"
	
	r.Body = http.MaxBytesReader(w, r.Body, uploadLimitInt)

	videoID, err := uuid.Parse(r.PathValue("videoID"))

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "could not find JWT", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "could not validate JWT", err)
		return
	}

	videoMeta, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(
			w,
			http.StatusNotFound,
			"could not get video meta data from database",
			err,
		)
		return
	}
	if videoMeta.UserID != userID {
		respondWithError(
			w,
			http.StatusUnauthorized,
			"video user id does not match user id",
			err,
		)
		return
	}

	multipartFile, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse from file", err)
		return
	}
	defer multipartFile.Close()

	mediaTypeString, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "cannot parse media type", err)
		return
	}
	if mediaTypeString != mimeTypeString {
		respondWithError(
			w,
			http.StatusBadRequest,
			fmt.Sprintf("uploaded file must have mime type `%s`", mimeTypeString),
			err,
		)
		return
	}

	osTempFile, err := os.CreateTemp(
		"",
		fileNameWithoutExtension + 
		"-" +
		"*" +
		"." +
		ExtractFileExtension(mimeTypeString),
	)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"could not create temporary file",
			err,
		)
		return
	}
    defer os.Remove(osTempFile.Name())
	defer osTempFile.Close()

	_, err = io.Copy(osTempFile, multipartFile)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"could not copy to the temporary file",
			err,
		)
		return
	}      

	_, err = osTempFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"could not reset offset of temporary file",
			err,
		)
		return
	}

	byteSliceForKey := make([]byte, 32)
	_, err = rand.Read(byteSliceForKey)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not read byte slice", err)
		return
	}
	aspectRatio, err := getVideoAspectRatio(osTempFile.Name())
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"could not get video's aspect ratio",
			err,
		)
		return
	}
	key := hex.EncodeToString(byteSliceForKey) + "." + ExtractFileExtension(mimeTypeString)
	if aspectRatio == "9:16" {
		key = "portrait" + "/" + key
	} else if aspectRatio == "16:9" {
		key = "landscape" + "/" + key
	}



   _, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: &cfg.s3Bucket,
		Key: &key,    
		Body: osTempFile,
		ContentType: &mimeTypeString,
	})
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"could not put object into S3",
			err,
		)
		return
	}

	videoURL := fmt.Sprintf(
		"https://%s.s3.%s.amazonaws.com/%s",
		cfg.s3Bucket,
		cfg.s3Region,
		key,
	)
	err = cfg.db.UpdateVideo(database.Video{
		ID: videoMeta.ID,
		CreatedAt: videoMeta.CreatedAt,
		UpdatedAt: videoMeta.UpdatedAt,
		ThumbnailURL: videoMeta.ThumbnailURL,
		VideoURL: &videoURL,
		CreateVideoParams: videoMeta.CreateVideoParams,
	})

}

























