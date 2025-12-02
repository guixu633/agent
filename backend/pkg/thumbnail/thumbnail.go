package thumbnail

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"strings"

	"golang.org/x/image/draw"
)

const (
	MaxThumbnailSize = 100      // 缩略图最大尺寸（长宽都不超过100）
	ThumbnailSuffix  = "_thumb" // 缩略图文件名后缀
)

// GenerateThumbnail 生成缩略图
// imageData: 原始图片数据
// mimeType: 图片 MIME 类型
// 返回: 缩略图数据
func GenerateThumbnail(imageData []byte, mimeType string) ([]byte, error) {
	// 解码图片
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %w", err)
	}

	// 获取原始图片尺寸
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 计算缩放比例（按长边缩放，确保长宽都不超过100）
	var scale float64
	if width > height {
		scale = float64(MaxThumbnailSize) / float64(width)
	} else {
		scale = float64(MaxThumbnailSize) / float64(height)
	}

	// 计算新尺寸
	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	// 创建新图片（使用高质量的双线性缩放）
	thumbnail := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// 使用双线性缩放算法（提供更好的质量）
	draw.BiLinear.Scale(thumbnail, thumbnail.Bounds(), img, img.Bounds(), draw.Src, nil)

	// 编码为图片格式
	var buf bytes.Buffer
	if strings.Contains(mimeType, "png") {
		err = png.Encode(&buf, thumbnail)
	} else {
		// 默认使用 JPEG
		err = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 85})
	}
	if err != nil {
		return nil, fmt.Errorf("编码缩略图失败: %w", err)
	}

	return buf.Bytes(), nil
}

// GetThumbnailFilename 获取缩略图文件名
// filename: 原始文件名
// 返回: 缩略图文件名
func GetThumbnailFilename(filename string) string {
	// 分离文件名和扩展名
	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 {
		// 没有扩展名，直接添加后缀
		return filename + ThumbnailSuffix
	}

	name := filename[:lastDot]
	ext := filename[lastDot:]
	return name + ThumbnailSuffix + ext
}

// GetOriginalFilename 从缩略图文件名获取原始文件名
// thumbnailFilename: 缩略图文件名
// 返回: 原始文件名，如果不是缩略图则返回空字符串
func GetOriginalFilename(thumbnailFilename string) string {
	if !strings.Contains(thumbnailFilename, ThumbnailSuffix) {
		return ""
	}
	return strings.Replace(thumbnailFilename, ThumbnailSuffix, "", 1)
}

// IsThumbnail 判断文件名是否为缩略图
func IsThumbnail(filename string) bool {
	return strings.Contains(filename, ThumbnailSuffix)
}
