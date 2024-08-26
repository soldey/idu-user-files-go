package userDto

import (
	"encoding/json"
	"fmt"
	"main/modules/common"
	"net/http"
	"strings"
)

type SelectManyFilesDTO struct {
	UserId    string              `json:"user_id"`
	Type      common.PlatformType `json:"type"`
	ProjectId *int                `json:"project_id"`
}

func (dto *SelectManyFilesDTO) FromRequest(r *http.Request) (error, int) {
	dtoString := fmt.Sprintf(
		"{\"user_id\": %s, \"type\": %s, \"project_id\": %s}",
		common.ParseParam(r.URL.Query().Get("user_id"), true),
		common.ParseParam(r.URL.Query().Get("type"), true),
		common.ParseParam(r.URL.Query().Get("project_id"), false),
	)
	err := json.NewDecoder(strings.NewReader(dtoString)).Decode(&dto)
	if err != nil {
		return err, http.StatusBadRequest
	}
	return nil, http.StatusOK
}
