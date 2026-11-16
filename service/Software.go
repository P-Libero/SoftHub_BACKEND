package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"softhub/model"
	"softhub/repository"
	"softhub/request"
	"strconv"

	"gorm.io/gorm"
)

type SoftwareService struct {
	Repo *repository.SoftwareRepository
}

func NewSoftwareService(repo *repository.SoftwareRepository, storage *storage.MinIO) *SoftwareService {
	return &SoftwareService{
		Repo:    repo,
		Storage: storage,
	}
}

func (s *SoftwareService) List(ctx context.Context) ([]model.Software, error) {
	return s.Repo.List(ctx)
}

func (s *SoftwareService) CreateBatch(ctx context.Context, req []request.SoftwareReq) ([]model.Software, error) {
	softwares := make([]model.Software, 0, len(req))
	for _, i := range req {
		softwares = append(softwares, model.Software{
			Name:        i.Name,
			Description: i.Description,
			Version:     i.Version,
		})
	}

	if err := s.Repo.CreateBatch(ctx, softwares); err != nil {
		return nil, err
	}
	return softwares, nil
}

func (s *SoftwareService) GetByID(ctx context.Context, id uint64) (*model.Software, error) {
	software, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("软件不存在")
		}
		return nil, err
	}
	return software, nil
}

func (s *SoftwareService) Update(ctx context.Context, id uint64, req *request.SoftwareReq) (*model.Software, error) {
	data := map[string]any{
		"name":        req.Name,
		"description": req.Description,
		"version":     req.Version,
	}
	software, err := s.Repo.Update(ctx, id, data)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("软件没有找到")
		}
		return nil, err
	}
	return software, nil
}

func (s *SoftwareService) Delete(ctx context.Context, id uint64) (*model.Software, error) {
	software, err := s.Repo.DeleteWithVersion(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("软件不存在")
		}
		return nil, err
	}
	uploadDir := filepath.Join("uploads", "software", strconv.FormatUint(id, 10))
	_ = os.RemoveAll(uploadDir)
	return software, nil
}

func (s *SoftwareService) CreateVersion(ctx context.Context, softwareID uint64, req *request.VersionReq) (*model.SoftwareVersion, error) {
	softwareVersion, err := s.Repo.CreateVersion(ctx, softwareID, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("软件不存在")
		}
		return nil, err
	}
	return softwareVersion, err
}

func (s *SoftwareService) ListAllVersions(ctx context.Context, softwareID uint64) ([]model.SoftwareVersion, error) {
	if _, err := s.Repo.FindByID(ctx, softwareID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("软件不存在")
		}
		return nil, err
	}
	return s.Repo.ListVersionBySoftwareID(ctx, softwareID)
}

func (s *SoftwareService) GetLatestVersion(ctx context.Context, softwareID uint64) (*model.SoftwareVersion, error) {
	if _, err := s.Repo.FindByID(ctx, softwareID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("软件不存在") //可以先找有没有
		}
		return nil, err
	}
	version, err := s.Repo.FindLatestVersion(ctx, softwareID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("该软件暂无发布版本")
		}
		return nil, err
	}
	return version, nil
}

func (s *SoftwareService) UploadVersionFile(ctx context.Context, softwareID uint64, versionID uint64, file *multipart.FileHeader) (*model.SoftwareVersion, error) {

	//先找你存不存在
	softwareVersion, err := s.Repo.FindVersionBySoftwareIDAndVersionID(ctx, softwareID, versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("软件版本不存在")
		}
		return nil, err
	}

	const maxFileSize = 500 << 20
	if file.Size > maxFileSize {
		return nil, errors.New("上传文件大小不能大于500MB")
	}

	//这里可能会有优化 比如以后加入对象存储时
	//先创建目录
	uploadDir := filepath.Join("uploads", "software", strconv.FormatUint(softwareID, 10))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, errors.New("创建目录失败")
	}
	//然后拼接上传路径
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("version-%d%s", versionID, ext)
	savePath := filepath.Join(uploadDir, filename)
	//将上传的文件写入磁盘
	if err := saveUploadedFile(file, savePath); err != nil {
		return nil, errors.New("上传文件失败")
	}
	//更新位置
	// TODO: 练习阶段暂存磁盘路径，接入对象存储后改为对象 key / 下载 URL
	if err := s.Repo.UpdateVersionDownloadURL(ctx, softwareVersion, savePath); err != nil {
		return nil, errors.New("更新软件下载地址失败")
	}
	return softwareVersion, nil
}

func (s *SoftwareService) GetDownloadFilePath(ctx context.Context, softwareID uint64, versionID uint64) (string, string, error) {
	if _, err := s.Repo.FindVersionBySoftwareIDAndVersionID(ctx, softwareID, versionID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errors.New("软件版本不存在")
		}
		return "", "", err
	}
	pattern := filepath.Join("uploads", "software", strconv.FormatUint(softwareID, 10), fmt.Sprintf("version-%d.*", versionID))
	//匹配所以后缀
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", "", errors.New("该版本没有上传文件")
	}
	filePath := matches[0]
	downloadName := filepath.Base(filePath)
	return filePath, downloadName, nil

}
