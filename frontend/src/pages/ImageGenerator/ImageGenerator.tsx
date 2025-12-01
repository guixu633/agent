import { useState } from 'react';
import { imageService } from '@/services/image/imageService';
import type { GeneratedImage } from '@/types/image';
import './ImageGenerator.css';

export function ImageGenerator() {
  const [prompt, setPrompt] = useState('');
  const [files, setFiles] = useState<File[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<{
    images: GeneratedImage[];
    description?: string;
  } | null>(null);

  // 处理文件选择
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = Array.from(e.target.files || []);
    setFiles(selectedFiles);
    setError(null);
  };

  // 移除文件
  const removeFile = (index: number) => {
    setFiles(files.filter((_, i) => i !== index));
  };

  // 提交生成请求
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!prompt.trim()) {
      setError('请输入提示词');
      return;
    }

    setLoading(true);
    setError(null);
    setResult(null);

    try {
      // 转换文件为 Base64
      const imageBase64List = await imageService.filesToBase64(files);

      // 调用 API
      const response = await imageService.generateImage({
        prompt: prompt.trim(),
        images: imageBase64List,
      });

      setResult(response);
    } catch (err) {
      setError(err instanceof Error ? err.message : '生成失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="image-generator">
      <div className="container">
        <h1>AI 图片生成器</h1>
        <p className="subtitle">使用 Gemini 3 Pro Image Preview 模型生成和编辑图片</p>

        <form onSubmit={handleSubmit} className="generator-form">
          {/* 提示词输入 */}
          <div className="form-group">
            <label htmlFor="prompt">提示词</label>
            <textarea
              id="prompt"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              placeholder="请描述你想要生成的图片，例如：帮我把图片修改为羊毛毡的可爱风格，短手短脚的那种可爱玩偶的感觉"
              rows={4}
              disabled={loading}
            />
          </div>

          {/* 图片上传 */}
          <div className="form-group">
            <label htmlFor="images">参考图片（可选）</label>
            <input
              type="file"
              id="images"
              accept="image/*"
              multiple
              onChange={handleFileChange}
              disabled={loading}
            />
            
            {files.length > 0 && (
              <div className="file-list">
                {files.map((file, index) => (
                  <div key={index} className="file-item">
                    <img
                      src={URL.createObjectURL(file)}
                      alt={file.name}
                      className="file-preview"
                    />
                    <span className="file-name">{file.name}</span>
                    <button
                      type="button"
                      onClick={() => removeFile(index)}
                      className="remove-btn"
                      disabled={loading}
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* 错误提示 */}
          {error && <div className="error-message">{error}</div>}

          {/* 提交按钮 */}
          <button
            type="submit"
            disabled={loading || !prompt.trim()}
            className="submit-btn"
          >
            {loading ? '生成中...' : '生成图片'}
          </button>
        </form>

        {/* 生成结果 */}
        {result && (
          <div className="result-section">
            <h2>生成结果</h2>
            
            {result.description && (
              <div className="description">
                <h3>描述</h3>
                <p>{result.description}</p>
              </div>
            )}

            {result.images.length > 0 && (
              <div className="generated-images">
                <h3>生成的图片</h3>
                <div className="image-grid">
                  {result.images.map((img, index) => (
                    <div key={index} className="generated-image-item">
                      <img
                        src={`data:${img.mimeType};base64,${img.data}`}
                        alt={`生成的图片 ${index + 1}`}
                        className="generated-image"
                      />
                      <a
                        href={`data:${img.mimeType};base64,${img.data}`}
                        download={`generated-${Date.now()}-${index}.png`}
                        className="download-btn"
                      >
                        下载图片
                      </a>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}


