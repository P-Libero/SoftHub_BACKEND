package repository

import (
	"context"
	"errors"
	"fmt"
	"softhub/model"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (u *UserRepository) CheckUsernameExist(ctx context.Context, username string) error {
	var count int64
	err := u.DB.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return fmt.Errorf("检查用户名唯一性失败: %w", err) //保留原始错误类型
	}
	if count > 0 {
		return ErrUsernameExists
	}
	return nil
}

func (u *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	if err := u.DB.WithContext(ctx).Create(user).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		//当 MySQL 数据库执行 SQL 报错时，MySQL 服务端真正发回给 Go 的不仅仅是一句英文，而是一个包含精确错误编号的数据包 它会被装进MySQLError这个专门的结构体里
		// MySQL 全世界统一的官方错误编号（Number）永远不会变：
		//1062：永远代表 唯一索引冲突（Duplicate entry，数据重复）

		//As判断 err 里面是不是包着某一种特定类型的结构体；如果是，就把它提取并赋值给第二个参数，好让你接下来去读它里面的字段
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUsernameExists
		}
		return err
	}
	return nil
}

func (u *UserRepository) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := u.DB.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepository) FindUserByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	if err := u.DB.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
