package oss

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// Config OSS 配置
type Config struct {
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	ImagePrefix     string `json:"image_prefix"` // 图片存储前缀，如 "image/"
}

// Client OSS 客户端
type Client struct {
	bucket   *oss.Bucket
	config   *Config
	baseURL  string // OSS 基础 URL，用于生成访问链接
}

// NewClient 创建 OSS 客户端
// configPath: 配置文件路径，可以是统一配置文件路径（config.json）或独立的 OSS 配置文件路径
// 如果 configPath 为空，默认使用 configs/config.json
func NewClient(configPath string) (*Client, error) {
	if configPath == "" {
		configPath = "configs/config.json"
	}

	// 尝试从统一配置文件加载
	ossConfig, err := loadFromUnifiedConfig(configPath)
	if err == nil {
		return createClientFromConfig(&ossConfig)
	}

	// 如果统一配置加载失败，尝试作为独立 OSS 配置文件加载（向后兼容）
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取 OSS 配置文件失败: %w", err)
	}

	var ossConfig2 Config
	if err := json.Unmarshal(data, &ossConfig2); err != nil {
		return nil, fmt.Errorf("解析 OSS 配置失败: %w", err)
	}

	return createClientFromConfig(&ossConfig2)
}

// loadFromUnifiedConfig 从统一配置文件加载 OSS 配置
func loadFromUnifiedConfig(configPath string) (Config, error) {
	if configPath == "" {
		configPath = "configs/config.json"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	var unifiedConfig struct {
		OSS struct {
			Endpoint        string `json:"endpoint"`
			Bucket          string `json:"bucket"`
			AccessKeyID     string `json:"access_key_id"`
			AccessKeySecret string `json:"access_key_secret"`
			ImagePrefix     string `json:"image_prefix"`
		} `json:"oss"`
	}

	if err := json.Unmarshal(data, &unifiedConfig); err != nil {
		return Config{}, err
	}

	return Config{
		Endpoint:        unifiedConfig.OSS.Endpoint,
		Bucket:          unifiedConfig.OSS.Bucket,
		AccessKeyID:     unifiedConfig.OSS.AccessKeyID,
		AccessKeySecret: unifiedConfig.OSS.AccessKeySecret,
		ImagePrefix:     unifiedConfig.OSS.ImagePrefix,
	}, nil
}

// createClientFromConfig 从配置结构创建客户端
func createClientFromConfig(config *Config) (*Client, error) {

	// 创建 OSS 客户端
	client, err := oss.New(config.Endpoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建 OSS 客户端失败: %w", err)
	}

	// 获取 Bucket
	bucket, err := client.Bucket(config.Bucket)
	if err != nil {
		return nil, fmt.Errorf("获取 Bucket 失败: %w", err)
	}

	// 构建基础 URL（用于生成访问链接）
	// OSS URL 格式: https://bucket-name.oss-region.aliyuncs.com
	endpoint := config.Endpoint
	// 移除协议前缀（如果有）
	if len(endpoint) > 7 && endpoint[:7] == "http://" {
		endpoint = endpoint[7:]
	} else if len(endpoint) > 8 && endpoint[:8] == "https://" {
		endpoint = endpoint[8:]
	}
	// 构建完整的 OSS URL
	baseURL := fmt.Sprintf("https://%s.%s", config.Bucket, endpoint)

	return &Client{
		bucket:  bucket,
		config:  config,
		baseURL: baseURL,
	}, nil
}

// UploadImage 上传图片到 OSS
// file: 要上传的文件
// filename: 文件名（可选，如果不提供则使用原始文件名）
// workspace: 工作区名称
// 返回: OSS 中的路径（相对于 bucket 根目录）
func (c *Client) UploadImage(file io.Reader, filename string, workspace string) (string, error) {
	// 如果没有提供文件名，生成一个基于时间戳的文件名
	if filename == "" {
		filename = fmt.Sprintf("%d%s", time.Now().UnixNano(), ".jpg")
	}

	// 构建路径: image/workspace/filename
	prefix := strings.TrimSuffix(c.config.ImagePrefix, "/")
	if prefix == "" {
		prefix = "image"
	}
	objectKey := fmt.Sprintf("%s/%s/%s", prefix, workspace, filename)
	// 确保路径使用正斜杠
	objectKey = strings.ReplaceAll(objectKey, "\\", "/")

	// 上传文件
	err := c.bucket.PutObject(objectKey, file)
	if err != nil {
		return "", fmt.Errorf("上传文件到 OSS 失败: %w", err)
	}

	return objectKey, nil
}

// UploadFile 上传文件到 OSS 指定路径
// path: OSS 中的完整路径
func (c *Client) UploadFile(reader io.Reader, path string) error {
	err := c.bucket.PutObject(path, reader)
	if err != nil {
		return fmt.Errorf("上传文件到 OSS 失败: %w", err)
	}
	return nil
}

// ListWorkspaces 列出所有工作区（查看 /image 下的目录）
func (c *Client) ListWorkspaces() ([]string, error) {
	prefix := strings.TrimSuffix(c.config.ImagePrefix, "/")
	if prefix == "" {
		prefix = "image"
	}
	searchPrefix := prefix + "/"

	// 列出所有对象，使用分隔符 "/" 来获取目录
	lsRes, err := c.bucket.ListObjects(oss.Prefix(searchPrefix), oss.Delimiter("/"))
	if err != nil {
		return nil, fmt.Errorf("列出工作区失败: %w", err)
	}

	workspaces := make([]string, 0)
	for _, commonPrefix := range lsRes.CommonPrefixes {
		// 提取工作区名称（去掉 image/ 前缀和末尾的 /）
		workspaceName := strings.TrimPrefix(commonPrefix, searchPrefix)
		workspaceName = strings.TrimSuffix(workspaceName, "/")
		if workspaceName != "" {
			workspaces = append(workspaces, workspaceName)
		}
	}

	return workspaces, nil
}

// CreateWorkspace 创建工作区（创建一个目录）
func (c *Client) CreateWorkspace(name string) error {
	prefix := strings.TrimSuffix(c.config.ImagePrefix, "/")
	if prefix == "" {
		prefix = "image"
	}
	// 创建一个空对象作为目录标记
	objectKey := fmt.Sprintf("%s/%s/.keep", prefix, name)
	objectKey = strings.ReplaceAll(objectKey, "\\", "/")

	// 创建一个空内容
	err := c.bucket.PutObject(objectKey, strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("创建工作区失败: %w", err)
	}

	return nil
}

// DeleteWorkspace 删除工作区（删除目录下的所有文件）
func (c *Client) DeleteWorkspace(name string) error {
	prefix := strings.TrimSuffix(c.config.ImagePrefix, "/")
	if prefix == "" {
		prefix = "image"
	}
	workspacePrefix := fmt.Sprintf("%s/%s/", prefix, name)
	workspacePrefix = strings.ReplaceAll(workspacePrefix, "\\", "/")

	// 列出工作区下的所有对象
	lsRes, err := c.bucket.ListObjects(oss.Prefix(workspacePrefix))
	if err != nil {
		return fmt.Errorf("列出工作区文件失败: %w", err)
	}

	// 删除所有对象
	objects := make([]string, 0, len(lsRes.Objects))
	for _, object := range lsRes.Objects {
		objects = append(objects, object.Key)
	}

	if len(objects) > 0 {
		_, err = c.bucket.DeleteObjects(objects)
		if err != nil {
			return fmt.Errorf("删除工作区文件失败: %w", err)
		}
	}

	return nil
}

// GetImageURL 获取图片的访问 URL
// path: OSS 中的路径（相对于 bucket 根目录）
// 返回: 完整的访问 URL
func (c *Client) GetImageURL(path string) string {
	return fmt.Sprintf("%s/%s", c.baseURL, path)
}

// DownloadImage 从 OSS 下载图片
// path: OSS 中的路径（相对于 bucket 根目录）
// 返回: 图片数据
func (c *Client) DownloadImage(path string) ([]byte, error) {
	body, err := c.bucket.GetObject(path)
	if err != nil {
		return nil, fmt.Errorf("从 OSS 下载文件失败: %w", err)
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}

	return data, nil
}

// GetImageURLOrDownload 获取图片 URL 或下载图片数据
// 如果模型支持 URL，返回 URL；否则下载图片数据
// path: OSS 中的路径
// useURL: 是否使用 URL（如果模型支持）
// 返回: URL（如果 useURL=true）或图片数据（如果 useURL=false）
func (c *Client) GetImageURLOrDownload(path string, useURL bool) (string, []byte, error) {
	if useURL {
		url := c.GetImageURL(path)
		return url, nil, nil
	}

	data, err := c.DownloadImage(path)
	if err != nil {
		return "", nil, err
	}

	return "", data, nil
}

// ImageInfo 图片信息
type ImageInfo struct {
	Path    string `json:"path"`    // OSS 中的路径
	URL     string `json:"url"`     // 图片访问 URL
	Name    string `json:"name"`    // 文件名
	Size    int64  `json:"size"`    // 文件大小（字节）
	Updated string `json:"updated"` // 最后更新时间（ISO 8601 格式）
}

// ListWorkspaceImages 列出工作区的所有图片
// workspace: 工作区名称
// 返回: 图片信息列表
func (c *Client) ListWorkspaceImages(workspace string) ([]ImageInfo, error) {
	prefix := strings.TrimSuffix(c.config.ImagePrefix, "/")
	if prefix == "" {
		prefix = "image"
	}
	workspacePrefix := fmt.Sprintf("%s/%s/", prefix, workspace)
	workspacePrefix = strings.ReplaceAll(workspacePrefix, "\\", "/")

	// 列出工作区下的所有对象
	lsRes, err := c.bucket.ListObjects(oss.Prefix(workspacePrefix))
	if err != nil {
		return nil, fmt.Errorf("列出工作区图片失败: %w", err)
	}

	images := make([]ImageInfo, 0)
	for _, object := range lsRes.Objects {
		// 跳过目录标记文件（如 .keep）
		if strings.HasSuffix(object.Key, "/.keep") {
			continue
		}

		// 提取文件名
		fileName := strings.TrimPrefix(object.Key, workspacePrefix)
		if fileName == "" {
			continue
		}

		// 构建图片信息
		imageInfo := ImageInfo{
			Path:    object.Key,
			URL:     c.GetImageURL(object.Key),
			Name:    fileName,
			Size:    object.Size,
			Updated: object.LastModified.Format(time.RFC3339),
		}

		images = append(images, imageInfo)
	}

	return images, nil
}

// DeleteImage 删除图片
// path: OSS 中的路径（相对于 bucket 根目录）
func (c *Client) DeleteImage(path string) error {
	err := c.bucket.DeleteObject(path)
	if err != nil {
		return fmt.Errorf("删除图片失败: %w", err)
	}
	return nil
}

// RenameImage 重命名图片（通过复制到新路径然后删除旧文件实现）
// oldPath: 旧路径（相对于 bucket 根目录）
// newName: 新文件名
// workspace: 工作区名称
// 返回: 新路径
func (c *Client) RenameImage(oldPath string, newName string, workspace string) (string, error) {
	prefix := strings.TrimSuffix(c.config.ImagePrefix, "/")
	if prefix == "" {
		prefix = "image"
	}
	
	// 构建新路径
	newPath := fmt.Sprintf("%s/%s/%s", prefix, workspace, newName)
	newPath = strings.ReplaceAll(newPath, "\\", "/")
	
	// 复制文件到新路径
	_, err := c.bucket.CopyObject(oldPath, newPath)
	if err != nil {
		return "", fmt.Errorf("复制图片失败: %w", err)
	}
	
	// 删除旧文件
	err = c.bucket.DeleteObject(oldPath)
	if err != nil {
		// 如果删除失败，尝试删除新文件以回滚
		c.bucket.DeleteObject(newPath)
		return "", fmt.Errorf("删除旧图片失败: %w", err)
	}
	
	return newPath, nil
}

