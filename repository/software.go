package repository

import (
	"context"
	"softhub/model"
	"softhub/request"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SoftwareRepository struct {
	DB *gorm.DB
}

func NewSoftwareRepository(db *gorm.DB) *SoftwareRepository {
	return &SoftwareRepository{
		DB: db,
	}
}

// 查询所有软件列

func (r *SoftwareRepository) List(ctx context.Context) ([]model.Software, error) {
	var softwares []model.Software
	err := r.DB.WithContext(ctx).Order("created_at desc").Find(&softwares).Error
	return softwares, err
}

//

func (r *SoftwareRepository) CreateBatch(ctx context.Context, software []model.Software) error {
	return r.DB.WithContext(ctx).Create(&software).Error
}

func (r *SoftwareRepository) FindByID(ctx context.Context, id uint64) (*model.Software, error) {
	var software model.Software
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&software).Error
	if err != nil {
		return nil, err
	}
	return &software, nil
}

func (r *SoftwareRepository) Update(ctx context.Context, id uint64, data map[string]any) (*model.Software, error) {
	var software model.Software
	if err := r.DB.WithContext(ctx).Where("id = ?", id).First(&software).Error; err != nil {
		return nil, err
	}
	if err := r.DB.WithContext(ctx).Model(&software).Updates(data).Error; err != nil {
		return nil, err
	}
	if err := r.DB.WithContext(ctx).First(&software, id).Error; err != nil {
		return nil, err
	}
	return &software, nil
}
func (r *SoftwareRepository) DeleteWithVersion(ctx context.Context, id uint64) (*model.Software, error) {
	var software model.Software
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).First(&software, id).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("software_id = ?", id).Delete(&model.SoftwareVersion{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Delete(&software).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &software, nil
}

// CreateVersionWithLock 事务创建新版本：锁住软件行 -> 计算最大 version_id + 1 -> 插入版本

func (r *SoftwareRepository) CreateVersion(ctx context.Context, softwareID uint64, Req *request.VersionReq) (*model.SoftwareVersion, error) {
	var softwareVersion model.SoftwareVersion
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var software model.Software
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&software, softwareID).Error; err != nil {
			return err
		}
		var maxVersionID uint
		if err := tx.Model(&softwareVersion).Where("software_id = ?", softwareID).Select("COALESCE(MAX(version_id),0)").Scan(&maxVersionID).Error; err != nil {
			return err
		}
		softwareVersion = model.SoftwareVersion{
			SoftwareID:  software.ID,
			VersionID:   maxVersionID + 1,
			Version:     Req.Version,
			Description: Req.Description,
		}
		if err := tx.Create(&softwareVersion).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &softwareVersion, err
}

func (r *SoftwareRepository) ListVersionBySoftwareID(ctx context.Context, softwareID uint64) ([]model.SoftwareVersion, error) {
	var softwareVersion []model.SoftwareVersion
	if err := r.DB.WithContext(ctx).Where("software_id = ?", softwareID).Order("version_id DESC").Find(&softwareVersion).Error; err != nil {
		return nil, err
	}
	return softwareVersion, nil
}

func (r *SoftwareRepository) FindLatestVersion(ctx context.Context, softwareID uint64) (*model.SoftwareVersion, error) {
	var softwareVersion model.SoftwareVersion
	if err := r.DB.WithContext(ctx).Where("software_id = ?", softwareID).Order("version_id DESC").First(&softwareVersion).Error; err != nil {
		return nil, err
	}
	return &softwareVersion, nil
}

func (r *SoftwareRepository) FindVersionBySoftwareIDAndVersionID(ctx context.Context, softwareID uint64, versionID uint64) (*model.SoftwareVersion, error) {
	var softwareVersion model.SoftwareVersion
	if err := r.DB.WithContext(ctx).Where("software_id = ? AND version_id = ?", softwareID, versionID).First(&softwareVersion).Error; err != nil {
		return nil, err
	}
	return &softwareVersion, nil
}

func (r *SoftwareRepository) UpdateVersionDownloadURL(ctx context.Context, version *model.SoftwareVersion, downloadURL string) error {
	return r.DB.WithContext(ctx).Model(version).Update("download_url", downloadURL).Error
}
