package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func (api *Api) ConvertFileHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(50 * 1024 * 1024) // 50 MB limit
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	srcFmt := r.FormValue("originalFormat")
	dstFmt := r.FormValue("targetFormat")
	file, fh, err := r.FormFile("conversionFile")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	defer file.Close()
	id := uuid.New().String()

	// TODO : write file to input/<id>
	path := filepath.Join("./input", fmt.Sprintf("%s%s", id, srcFmt))
	tFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	defer tFile.Close()
	_, err = io.Copy(tFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	req := ConversionRequest{
		ConversionRequestBody: ConversionRequestBody{
			File:           []byte{}, // TODO : remove
			Filename:       fh.Filename,
			OriginalFormat: srcFmt,
			TargetFormat:   dstFmt,
		},
		ConversionStatus:     "in queue",
		ExternalConversionID: id,
		WorkerConversionID:   "",
	}
	// Send the request to the channel
	api.RequestChannel <- &req
	respondWithJSON(w, http.StatusOK, ConversionProcessingResponse{
		ConversionID: id,
	})
}

func (api *Api) ConversionQueueStatusHandler(w http.ResponseWriter, r *http.Request) {
	// TODO : implement
}

func (api *Api) GetConvertedFileHandler(w http.ResponseWriter, r *http.Request) {
	// TODO : implement
}

func (api *Api) GetConvertedFileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	// TODO : implement
}
