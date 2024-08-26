package userDto

import (
	"main/modules/common"
)

type CreateFileDTO struct {
	Filename  string              `json:"filename"`
	Public    bool                `json:"public"`
	Owner     string              `json:"owner"`
	Type      common.PlatformType `json:"type"`
	ProjectId *int                `json:"project_id"`
}
