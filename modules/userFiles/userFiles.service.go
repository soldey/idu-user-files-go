package userFiles

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/uptrace/bun"
	"io"
	"main/modules/common"
	"main/modules/database"
	"main/modules/userFiles/userDto"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type IUserFilesService interface {
	CreateFile(dto *userDto.CreateFileDTO, file multipart.File, header *multipart.FileHeader, ctx *context.Context) (*UserFile, error, int)
	SelectOneFile(dto userDto.SelectOneFileDTO, tx *bun.Tx, ctx *context.Context) (*UserFile, error, int)
}

type UserFilesService struct {
	basis IUserFilesService
}

func (s *UserFilesService) CreateFile(
	dto *userDto.CreateFileDTO, file multipart.File, header *multipart.FileHeader, ctx *context.Context,
) (*UserFile, error, int) {
	database.DbConfig.Connect()

	tx, err := database.Service.BeginTx(*ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	exists, _, _ := s.SelectOneFile(&userDto.SelectOneFileDTO{
		Filename:  dto.Filename,
		UserId:    dto.Owner,
		Type:      dto.Type,
		ProjectId: dto.ProjectId,
	}, &tx, ctx)
	if exists != nil {
		return nil, fmt.Errorf("FILE_ALREADY_PRESENT_IN_DATABASE"), http.StatusConflict
	}

	userFile := EntityFromDTO(dto)
	err = tx.NewInsert().Model(userFile).Returning("*").Scan(*ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err, http.StatusInternalServerError
	}
	minioClient, err := minio.New(common.Config.Get("FILESERVER_ADDR"), &minio.Options{
		Creds: credentials.NewStaticV4(
			os.Getenv("FILESERVER_ACCESS_KEY"), os.Getenv("FILESERVER_SECRET_KEY"), "",
		),
		Secure: false,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, err, http.StatusInternalServerError
	}
	_, err = minioClient.PutObject(
		*ctx,
		common.Config.Get("FILESERVER_BUCKET")+"-"+strings.ToLower(string(userFile.Type)),
		fmt.Sprintf("%s.%s", strconv.FormatInt(userFile.Id, 10), userFile.Ext),
		file, header.Size,
		minio.PutObjectOptions{})
	if err != nil {
		_ = tx.Rollback()
		return nil, err, http.StatusInternalServerError
	}
	_ = tx.Commit()
	return userFile, nil, http.StatusOK
}

func (s *UserFilesService) SelectOneFile(
	dto *userDto.SelectOneFileDTO, tx *bun.Tx, ctx *context.Context,
) (*UserFile, error, int) {
	filename, ext := common.PrepareFilename(dto.Filename)
	entity := new(UserFile)
	var statement *bun.SelectQuery
	if tx != nil {
		statement = tx.NewSelect()
	} else {
		statement = database.Service.NewSelect()
	}
	statement = statement.Model(entity).
		Where("filename = ?", filename).
		Where("ext = ?", ext).
		Where("type = ?::platformtypeenum", dto.Type)
	if dto.ProjectId == nil {
		statement = statement.Where("project_id IS ?", dto.ProjectId)
	} else {
		statement = statement.Where("project_id = ?", *dto.ProjectId)
	}
	err := statement.
		Limit(1).
		Scan(*ctx)
	if err != nil {
		return nil, fmt.Errorf("FILE_NOT_FOUND"), http.StatusNotFound
	}
	return entity, nil, http.StatusOK
}

func (s *UserFilesService) DownloadOneFile(
	dto *userDto.SelectOneFileDTO, ctx *context.Context,
) (*[]byte, error, int) {
	userFile, err, status := s.SelectOneFile(dto, nil, ctx)
	if err != nil {
		return nil, err, status
	}

	minioClient, err := minio.New(common.Config.Get("FILESERVER_ADDR"), &minio.Options{
		Creds: credentials.NewStaticV4(
			os.Getenv("FILESERVER_ACCESS_KEY"), os.Getenv("FILESERVER_SECRET_KEY"), "",
		),
		Secure: false,
	})
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	object, err := minioClient.GetObject(
		*ctx,
		common.Config.Get("FILESERVER_BUCKET")+"-"+strings.ToLower(string(userFile.Type)),
		fmt.Sprintf("%s.%s", strconv.FormatInt(userFile.Id, 10), userFile.Ext),
		minio.GetObjectOptions{})
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	defer object.Close()
	stat, err := object.Stat()
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	bytes := make([]byte, stat.Size)
	_, err = object.Read(bytes)
	if err != nil && err != io.EOF {
		return nil, err, http.StatusInternalServerError
	}
	return &bytes, nil, http.StatusOK
}

var Service = &UserFilesService{}
