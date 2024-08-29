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
	"runtime"
	"slices"
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

	tx, err := database.Database.BeginTx(*ctx, &sql.TxOptions{})
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
	backgroundCtx := context.Background()
	go minioClient.PutObject(
		backgroundCtx,
		common.Config.Get("FILESERVER_BUCKET")+"-"+strings.ToLower(string(userFile.Type)),
		fmt.Sprintf("%s.%s", strconv.FormatInt(userFile.Id, 10), userFile.Ext),
		file, header.Size,
		minio.PutObjectOptions{NumThreads: uint(runtime.NumCPU()) * 2})
	bytes := make([]byte, header.Size)
	_, err = file.Read(bytes)
	if err == nil {
		go database.Redis.SaveBytes(&backgroundCtx, fmt.Sprintf("user_files:%d.%s", userFile.Id, userFile.Ext), bytes, 300)
	} else {
		fmt.Println(err.Error())
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
		statement = database.Database.NewSelect()
	}
	statement = statement.Model(entity).
		Where("filename = ?", filename).
		Where("ext = ?", ext).
		Where("type = ?::platformtypeenum", dto.Type).
		Where("is_deleted = ?", false)
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
	if userFile.Public == false && userFile.Owner != dto.UserId {
		return nil, fmt.Errorf("ACCESS_DENIED"), http.StatusForbidden
	}
	key := fmt.Sprintf("user_files:%d.%s", userFile.Id, userFile.Ext)
	redisKeys := database.Redis.GetStringList(ctx)
	fmt.Printf("%s %+v\n", key, redisKeys)
	if slices.Contains(
		database.Redis.GetStringList(ctx), key,
	) {
		bytes, err := database.Redis.GetBytes(ctx, key)
		if err != nil {
			return nil, err, http.StatusInternalServerError
		}
		database.Redis.SetTTL(ctx, key, 300)
		return &bytes, nil, http.StatusOK
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
	database.Redis.SaveBytes(ctx, fmt.Sprintf("user_files:%d.%s", userFile.Id, userFile.Ext), bytes, 300)
	return &bytes, nil, http.StatusOK
}

func (s *UserFilesService) GetUserFilesList(
	dto *userDto.SelectManyFilesDTO, ctx *context.Context,
) (*map[string][]string, error, int) {
	entities := make([]UserFile, 0)
	statement := database.Database.NewSelect().Model(&entities).
		Column("filename", "ext", "public").
		Where("(public = ? or owner = ?)", true, dto.UserId).
		Where("type = ?::platformtypeenum", dto.Type).
		Where("is_deleted = ?", false)
	if dto.ProjectId != nil {
		statement = statement.Where("project_id = ?", dto.ProjectId)
	}
	err := statement.
		Group("public", "filename", "ext", "updated_at").
		Order("public DESC", "updated_at DESC").
		Scan(*ctx)
	if err != nil {
		return nil, err, http.StatusNotFound
	}
	resultMap := map[string][]string{}
	for _, v := range entities {
		if v.Public == true {
			if _, ok := resultMap["public"]; !ok {
				resultMap["public"] = make([]string, 0)
			}
			resultMap["public"] = append(resultMap["public"], v.Filename+"."+v.Ext)
		} else {
			if _, ok := resultMap["private"]; !ok {
				resultMap["private"] = make([]string, 0)
			}
			resultMap["private"] = append(resultMap["private"], v.Filename+"."+v.Ext)
		}
	}
	return &resultMap, nil, http.StatusOK
}

func (s *UserFilesService) PatchUserFile(
	dto *userDto.SelectOneFileDTO, entity *userDto.PatchFileDTO, ctx *context.Context,
) (*UserFile, error, int) {
	exists := make([]UserFile, 0)
	statement := database.Database.NewSelect().Model(&exists).
		Where("is_deleted = ?", false)
	if entity.Filename != nil {
		filename, ext := common.PrepareFilename(*entity.Filename)
		entity.Filename = &filename
		statement = statement.Where("filename = ?", filename).Where("ext = ?", ext)
	}
	if entity.Type != nil {
		statement = statement.Where("type = ?::platformtypeenum", entity.Type)
	}
	err := statement.Scan(*ctx)
	fmt.Printf("%+v\n", exists)
	if err != nil {
		return nil, fmt.Errorf("SOMETHING_IS_WRONG"), http.StatusInternalServerError
	}
	if len(exists) != 0 {
		return nil, fmt.Errorf("NEW_FILE_PRESENT_IN_DATABASE"), http.StatusConflict
	}
	exists = nil

	userFile, err, status := s.SelectOneFile(dto, nil, ctx)
	if err != nil {
		return nil, err, status
	}
	if userFile.Public == false && userFile.Owner != dto.UserId {
		return nil, fmt.Errorf("ACCESS_DENIED"), http.StatusForbidden
	}

	userFile.FillEntityFromPatchDTO(entity)
	_, err = database.Database.NewUpdate().Model(userFile).WherePK().Returning("*").Exec(*ctx)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return userFile, nil, http.StatusOK
}

func (s *UserFilesService) DeleteUserFile(dto *userDto.SelectOneFileDTO, ctx *context.Context) (*UserFile, error, int) {
	userFile, err, status := s.SelectOneFile(dto, nil, ctx)
	if err != nil {
		return nil, err, status
	}
	if userFile.Public == false && userFile.Owner != dto.UserId {
		return nil, fmt.Errorf("ACCESS_DENIED"), http.StatusForbidden
	}
	_, err = database.Database.NewUpdate().Model(userFile).
		Set("is_deleted = ?", true).
		WherePK().
		Returning("*").
		Exec(*ctx)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return userFile, nil, http.StatusOK
}

var Service *UserFilesService
