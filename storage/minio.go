package storage

import (
	"context"
	"fmt"
	"io"
	"softhub/config"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIO struct {
	Client *minio.Client
	Bucket string
	Expire time.Duration
}

func NewMinIO(cfg config.MinIO) (*MinIO, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL, // false 走 http，true 走 https
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client failed: %w", err)
	}
	//发送请求 5秒探活 并创建ctx子上下文 然下文继续监听ctx.done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	//检查桶的名字是否存在
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket exists failed: %w", err)
	}
	//如果不存在 则自动创建存储桶
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket %s failed: %w", cfg.Bucket, err)
		}
	}
	return &MinIO{
		Client: client,
		Bucket: cfg.Bucket,
		Expire: 15 * time.Minute,
	}, nil

}

func (m *MinIO) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	// 第 1 个返回值是上传回执信息（含 ETag 等），不需要故用 _ 忽略
	//它是 MinIO SDK 最核心的文件上传函数 相当于 HTTP 里的 PUT 请求
	//桶里已经有一个同名的 objectName 文件，PutObject 会直接覆盖原文件。
	_, err := m.Client.PutObject(ctx, m.Bucket, objectName, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("upload file failed: %w", err)
	}
	return nil
}

// GeneratePresignedURL 为私有桶中的对象生成带有效期的临时下载链接
func (m *MinIO) GeneratePresignedURL(ctx context.Context, objectName string) (string, error) {
	// PresignedGetObject 在本地用 SecretKey 计算 HMAC-SHA256 签名，生成有效期为 m.Expire 的临时 URL
	// 最后一个参数是额外查询参数（url.Values）
	url, err := m.Client.PresignedGetObject(ctx, m.Bucket, objectName, m.Expire, nil)
	if err != nil {
		return "", fmt.Errorf("generate presigned url failed: %w", err)
	}
	// 将 *url.URL 结构体指针转为完整的 URL 字符串返回
	//这里toString会自动拼接结构体字段组成url还有？后的凭证
	return url.String(), nil
}

func (m *MinIO) DeleteFile(ctx context.Context, objectName string) error {
	//RemoveObjectOptions{} 为默认空选项
	if err := m.Client.RemoveObject(ctx, m.Bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete object failed: %w", err)
	}
	return nil
}
