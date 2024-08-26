package userFiles

import (
	"encoding/json"
	"main/modules/userFiles/userDto"
	"net/http"
)

func CreateFile(w http.ResponseWriter, r *http.Request) {
	dto, err, status := userDto.FromRequest(r)
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
