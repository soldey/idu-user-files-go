package userFiles

import (
	"github.com/uptrace/bun"
	"main/modules/common"
	"main/modules/userFiles/userDto"
	"time"
)

type UserFile struct {
	bun.BaseModel `bun:"table:user_files"`

	Id        int64               `bun:"id,pk,autoincrement" json:"id"`
	Filename  string              `bun:"filename,notnull" json:"filename"`
	Ext       string              `bun:"ext,notnull" json:"ext"`
	Public    bool                `bun:"public,notnull" json:"public"`
	Owner     string              `bun:"owner,notnull" json:"owner"`
	Type      common.PlatformType `bun:"column:type,notnull" json:"type"`
	ProjectId *int                `bun:"project_id,nullzero" json:"project_id"`
	CreatedAt time.Time           `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time           `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	IsDeleted bool                `bun:"is_deleted,notnull" json:"is_deleted"`
}

func EntityFromDTO(dto *userDto.CreateFileDTO) *UserFile {
	filename, ext := common.PrepareFilename(dto.Filename)
	return &UserFile{
		Filename:  filename,
		Ext:       ext,
		Public:    dto.Public,
		Owner:     dto.Owner,
		Type:      dto.Type,
		ProjectId: dto.ProjectId,
		IsDeleted: false,
	}
}

func (entity *UserFile) FillEntityFromPatchDTO(dto *userDto.PatchFileDTO) {
	if dto.Filename != nil {
		entity.Filename = *dto.Filename
	}
	if dto.Public != nil {
		entity.Public = *dto.Public
	}
	if dto.Type != nil {
		entity.Type = *dto.Type
	}
	if dto.ProjectId != nil {
		entity.ProjectId = dto.ProjectId
	}
	return
}
