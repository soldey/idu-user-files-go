package userFiles

import (
	"encoding/json"
	"main/modules/common"
	"main/modules/userFiles/userDto"
	"net/http"
)

func CreateFile(w http.ResponseWriter, r *http.Request) {
	dto := new(userDto.CreateFileDTO)
	err, status := dto.FromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	ctx := r.Context()
	entity, err, status := Service.CreateFile(dto, file, header, &ctx)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(entity)
}

func SelectFile(w http.ResponseWriter, r *http.Request) {
	dto := new(userDto.SelectOneFileDTO)
	err, status := dto.FromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	ctx := r.Context()
	bytes, err, status := Service.DownloadOneFile(dto, &ctx)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	_, ext := common.PrepareFilename(dto.Filename)
	w.Header().Set("Content-Type", common.GetMediaType(ext))
	w.WriteHeader(status)
	w.Write(*bytes)
}
