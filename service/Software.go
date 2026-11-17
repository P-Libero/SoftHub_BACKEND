package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"softhub/model"
	"softhub/repository"
	"softhub/request"
	"softhub/storage"

	"gorm.io/gorm"
)

type SoftwareService struct {
	Repo    *repository.SoftwareRepository
	Storage *storage.MinIO
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

// 删除软件版本 记录 以及 MinIO中对应的安装包对象

func (s *SoftwareService) Delete(ctx context.Context, id uint64) (*model.Software, error) {

	version, _ := s.Repo.ListVersionBySoftwareID(ctx, id)
	//进行数据库的删除任务
	software, err := s.Repo.DeleteWithVersion(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("软件不存在")
		}
		return nil, err
	}

	//进行对象存储的删除
	for _, v := range version {
		if v.DownloadURL != "" {
			_ = s.Storage.DeleteFile(ctx, v.DownloadURL)
		}
	}

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

	src, err := file.Open()
	if err != nil {
		return nil, errors.New("打开上传文件失败")
	}
	defer src.Close()                  //防止句柄泄露
	ext := filepath.Ext(file.Filename) //提取后缀
	objectName := fmt.Sprintf("software/%d/version-%d%s", softwareID, versionID, ext)
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		//如果找不到格式就默认二进制
		contentType = "application/octet-stream"
	}
	//先上传到MinIO
	if err := s.Storage.UploadFile(ctx, objectName, src, file.Size, contentType); err != nil {
		return nil, errors.New("上传文件失败")
	}

	//将对象键名（objectName）持久化写入数据库的 download_url 字段
	if err := s.Repo.UpdateVersionDownloadURL(ctx, softwareVersion, objectName); err != nil {
		return nil, errors.New("更新软件下载地址失败")
	}
	return softwareVersion, nil
}

func (s *SoftwareService) GetDownloadURL(ctx context.Context, softwareID uint64, versionID uint64) (string, error) {
	softwareVersion, err := s.Repo.FindVersionBySoftwareIDAndVersionID(ctx, softwareID, versionID) //先找存不存在
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("软件版本不存在")
		}
		return "", err
	}
	if softwareVersion.DownloadURL == "" {
		return "", errors.New("该版本暂未上传文件")
	}

	downloadURL, err := s.Storage.GeneratePresignedURL(ctx, softwareVersion.DownloadURL)
	if err != nil {
		return "", errors.New("获取下载链接失败")
	}
	return downloadURL, nil
}
