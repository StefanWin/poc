package api

import (
	"net/http"

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
	// TODO : write file to input/<id>
	id := uuid.New().String()
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
	api.RequestChannel <- req
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
