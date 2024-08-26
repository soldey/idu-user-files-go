package userDto

import (
	"encoding/json"
	"fmt"
	"main/modules/common"
	"net/http"
	"strings"
)

type CreateFileDTO struct {
	Filename  string              `json:"filename"`
	Public    bool                `json:"public"`
	Owner     string              `json:"owner"`
	Type      common.PlatformType `json:"type"`
	ProjectId *int                `json:"project_id"`
}

func FromRequest(r *http.Request) (*CreateFileDTO, error, int) {
	dtoString := fmt.Sprintf(
		"{\"filename\": %s, \"public\": %s, \"owner\": %s, \"type\": %s, \"project_id\": %s}",
		common.ParseParam(r.URL.Query().Get("filename"), true),
		common.ParseParam(r.URL.Query().Get("public"), false),
		common.ParseParam(r.URL.Query().Get("owner"), true),
		common.ParseParam(r.URL.Query().Get("type"), true),
		common.ParseParam(r.URL.Query().Get("project_id"), false),
	)
	dto := new(CreateFileDTO)
	err := json.NewDecoder(strings.NewReader(dtoString)).Decode(&dto)
	if err != nil {
		return nil, err, http.StatusBadRequest
	}
	return dto, nil, http.StatusOK
}
