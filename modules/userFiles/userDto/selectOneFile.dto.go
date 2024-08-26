package userDto

import (
	"main/modules/common"
)

type SelectOneFileDTO struct {
	Filename  string              `json:"filename"`
	UserId    string              `json:"user_id"`
	Type      common.PlatformType `json:"type"`
	ProjectId *int                `json:"project_id"`
}
