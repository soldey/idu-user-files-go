package userDto

import (
	"encoding/json"
	"main/modules/common"
	"net/http"
)

type PatchFileDTO struct {
	Filename  *string              `json:"filename"`
	Public    *bool                `json:"public"`
	Type      *common.PlatformType `json:"type"`
	ProjectId *int                 `json:"project_id"`
}

func (dto *PatchFileDTO) FromRequest(r *http.Request) (error, int) {
	err := json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		return err, http.StatusBadRequest
	}
	return nil, http.StatusOK
}
