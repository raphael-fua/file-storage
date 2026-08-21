package main

import (
	"strings"
)

//--------------------------------------------------------------
// PARAMETERS:
//   `mediaType`: a `string`
//     eg: `"image/png"`
// OUTPUT:
//   `fileExtension`: a `string`
//     eg: `"png"` if `mediaType` is `"image/png"`
func ExtractFileExtension(mediaType string) (fileExtension string) {
	words := strings.Split(mediaType, "/")
	return words[len(words) - 1]
}
//--------------------------------------------------------------






