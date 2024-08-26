package userFiles

import (
	"encoding/json"
	"fmt"
	"main/modules/userFiles/userDto"
	"net/http"
	"strings"
)

func CreateFile(w http.ResponseWriter, r *http.Request) {
	dtoString := fmt.Sprintf(
		"{\"filename\": %s, \"public\": %s, \"owner\": %s, \"type\": %s, \"project_id\": %s}",
		ParseParam(r.URL.Query().Get("filename"), true),
		ParseParam(r.URL.Query().Get("public"), false),
		ParseParam(r.URL.Query().Get("owner"), true),
		ParseParam(r.URL.Query().Get("type"), true),
		ParseParam(r.URL.Query().Get("project_id"), false),
	)
	dto := new(userDto.CreateFileDTO)
	err := json.NewDecoder(strings.NewReader(dtoString)).Decode(&dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

func ParseParam(param string, isString bool) string {
	if param == "" {
		return "null"
	}
	if isString {
		return fmt.Sprintf("\"%s\"", param)
	}
	return param
}
