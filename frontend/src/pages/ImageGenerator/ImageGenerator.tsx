import { useState, useRef, useEffect } from 'react';
import { Edit, Trash2 } from 'lucide-react';
import { imageService } from '@/services/image/imageService';
import { workspaceService } from '@/services/workspace/workspaceService';
import type { ImageGenerateResponse, ImageInfo } from '@/types/image';
import type { Workspace } from '@/types/workspace';
import { formatDateOnlyToBeijing, formatDateTimeToBeijing } from '@/utils/date';
import './ImageGenerator.css';

export function ImageGenerator() {
  const [prompt, setPrompt] = useState('');
  const [selectedImages, setSelectedImages] = useState<ImageInfo[]>([]); // 从列表中选择的图片
  const [uploading, setUploading] = useState(false); // 上传状态
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<ImageGenerateResponse | null>(null);
  const [generateCount, setGenerateCount] = useState(1); // 生成数量，默认1张，最多3张
  const [results, setResults] = useState<Array<{
    data: ImageGenerateResponse | null;
    status: 'pending' | 'generating' | 'success' | 'error';
    elapsedTime: number | null;
    error?: string;
  }>>([]);
  
  // Workspace 相关状态
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [currentWorkspace, setCurrentWorkspace] = useState<string>('');
  const [showWorkspaceModal, setShowWorkspaceModal] = useState(false);
  const [showWorkspaceDropdown, setShowWorkspaceDropdown] = useState(false);
  const [newWorkspaceName, setNewWorkspaceName] = useState('');
  const [workspaceLoading, setWorkspaceLoading] = useState(false);
  const workspaceDropdownRef = useRef<HTMLDivElement>(null);
  const isLoadingRef = useRef(false); // 防止重复加载
  
  // 图片列表相关状态
  const [workspaceImages, setWorkspaceImages] = useState<ImageInfo[]>([]);
  const [showImageList, setShowImageList] = useState(true); // 默认展开
  const [imagesLoading, setImagesLoading] = useState(false);
  const imageListRef = useRef<HTMLDivElement>(null);
  
  // 重命名相关状态
  const [renamingImage, setRenamingImage] = useState<ImageInfo | null>(null);
  const [newImageName, setNewImageName] = useState('');
  const [imageExtension, setImageExtension] = useState(''); // 保存文件扩展名
  
  // 图片预览相关状态
  const [previewImage, setPreviewImage] = useState<ImageInfo | null>(null);
  
  // 计时器状态
  const [elapsedTime, setElapsedTime] = useState(0);
  const timerRef = useRef<number | null>(null);
  const startTimeRef = useRef<number | null>(null);

  // 清除计时器
  const clearTimer = () => {
    if (timerRef.current) {
      window.clearInterval(timerRef.current);
      timerRef.current = null;
    }
    startTimeRef.current = null;
  };

  // 组件卸载时清理
  useEffect(() => {
    return () => clearTimer();
  }, []);

  // 点击外部关闭下拉菜单
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        workspaceDropdownRef.current &&
        !workspaceDropdownRef.current.contains(event.target as Node)
      ) {
        setShowWorkspaceDropdown(false);
      }
      if (
        imageListRef.current &&
        !imageListRef.current.contains(event.target as Node)
      ) {
        setShowImageList(false);
      }
    };

    if (showWorkspaceDropdown || showImageList) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showWorkspaceDropdown, showImageList]);

  // 初始化：加载工作区列表和当前工作区
  useEffect(() => {
    if (isLoadingRef.current) return;
    isLoadingRef.current = true;
    
    const initializeWorkspaces = async () => {
      try {
        // 先加载工作区列表（包含 is_current 字段）
        const workspacesResponse = await workspaceService.listWorkspaces();
        setWorkspaces(workspacesResponse.workspaces);
        
        // 从列表中查找当前工作区（is_current: true）
        const currentWs = workspacesResponse.workspaces.find(w => w.is_current);
        if (currentWs) {
          setCurrentWorkspace(currentWs.name);
        } else if (workspacesResponse.workspaces.length > 0) {
          // 如果没有当前工作区，使用第一个
          setCurrentWorkspace(workspacesResponse.workspaces[0].name);
        }
      } catch (err) {
        console.error('初始化工作区失败:', err);
      } finally {
        isLoadingRef.current = false;
      }
    };
    
    initializeWorkspaces();
  }, []);

  // 当工作区改变时，加载图片列表并清空选中的图片
  useEffect(() => {
    if (currentWorkspace && !isLoadingRef.current) {
      setSelectedImages([]); // 切换工作区时清空选中的图片
      loadWorkspaceImages();
    }
  }, [currentWorkspace]);


  // 创建工作区
  const handleCreateWorkspace = async () => {
    if (!newWorkspaceName.trim()) {
      setError('工作区名称不能为空');
      return;
    }

    setWorkspaceLoading(true);
    try {
      const workspaceName = newWorkspaceName.trim();
      await workspaceService.createWorkspace({ name: workspaceName });
      // 创建后自动设置为当前工作区
      await workspaceService.setCurrentWorkspace({ name: workspaceName });
      setNewWorkspaceName('');
      setShowWorkspaceModal(false);
      // 重新加载工作区列表并更新当前工作区
      const response = await workspaceService.listWorkspaces();
      setWorkspaces(response.workspaces);
      setCurrentWorkspace(workspaceName); // 这会触发图片加载
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建工作区失败');
    } finally {
      setWorkspaceLoading(false);
    }
  };

  // 删除工作区
  const handleDeleteWorkspace = async (name: string) => {
    if (!confirm(`确定要删除工作区 "${name}" 吗？这将删除该工作区下的所有文件。`)) {
      return;
    }

    setWorkspaceLoading(true);
    try {
      await workspaceService.deleteWorkspace({ name });
      // 重新加载工作区列表
      const response = await workspaceService.listWorkspaces();
      setWorkspaces(response.workspaces);
      
      // 如果删除的是当前工作区，切换到第一个或当前工作区
      if (currentWorkspace === name) {
        const currentWs = response.workspaces.find(w => w.is_current);
        if (currentWs) {
          setCurrentWorkspace(currentWs.name);
        } else if (response.workspaces.length > 0) {
          setCurrentWorkspace(response.workspaces[0].name);
        } else {
          setCurrentWorkspace('');
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除工作区失败');
    } finally {
      setWorkspaceLoading(false);
    }
  };

  // 加载工作区图片列表
  const loadWorkspaceImages = async () => {
    if (!currentWorkspace) return;
    
    setImagesLoading(true);
    try {
      const response = await imageService.listWorkspaceImages(currentWorkspace);
      setWorkspaceImages(response.images);
    } catch (err) {
      console.error('加载图片列表失败:', err);
      setWorkspaceImages([]);
    } finally {
      setImagesLoading(false);
    }
  };

  // 选择/取消选择图片
  const handleSelectImage = (image: ImageInfo) => {
    setSelectedImages((prev) => {
      const isSelected = prev.some((img) => img.path === image.path);
      if (isSelected) {
        // 如果已选中，则取消选择
        return prev.filter((img) => img.path !== image.path);
      } else {
        // 如果未选中，则添加到选择列表
        return [...prev, image];
      }
    });
  };

  // 检查图片是否被选中
  const isImageSelected = (image: ImageInfo) => {
    return selectedImages.some((img) => img.path === image.path);
  };

  // 移除选中的图片
  const removeSelectedImage = (imagePath: string) => {
    setSelectedImages((prev) => prev.filter((img) => img.path !== imagePath));
  };

  // 删除图片
  const handleDeleteImage = async (image: ImageInfo, e: React.MouseEvent) => {
    e.stopPropagation(); // 阻止触发选择事件
    
    if (!confirm(`确定要删除图片 "${image.name}" 吗？`)) {
      return;
    }

    try {
      await imageService.deleteImage({ path: image.path });
      // 从列表中移除
      setWorkspaceImages((prev) => prev.filter((img) => img.path !== image.path));
      // 如果图片被选中，也从选中列表中移除
      setSelectedImages((prev) => prev.filter((img) => img.path !== image.path));
      // 刷新列表
      await loadWorkspaceImages();
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除图片失败');
    }
  };

  // 开始重命名
  const handleStartRename = (image: ImageInfo, e: React.MouseEvent) => {
    e.stopPropagation(); // 阻止触发选择事件
    
    // 提取文件名和扩展名
    const lastDotIndex = image.name.lastIndexOf('.');
    let fileName = image.name;
    let ext = '';
    
    if (lastDotIndex > 0 && lastDotIndex < image.name.length - 1) {
      // 有扩展名
      fileName = image.name.substring(0, lastDotIndex);
      ext = image.name.substring(lastDotIndex);
    }
    
    setRenamingImage(image);
    setNewImageName(fileName);
    setImageExtension(ext);
  };

  // 取消重命名
  const handleCancelRename = () => {
    setRenamingImage(null);
    setNewImageName('');
    setImageExtension('');
  };

  // 确认重命名
  const handleConfirmRename = async () => {
    if (!renamingImage || !newImageName.trim()) {
      return;
    }

    // 构建完整的新文件名（文件名 + 扩展名）
    const fullNewName = newImageName.trim() + imageExtension;

    // 如果新名称和旧名称相同，直接取消
    if (fullNewName === renamingImage.name) {
      handleCancelRename();
      return;
    }

    try {
      const response = await imageService.renameImage({
        path: renamingImage.path,
        new_name: fullNewName,
        workspace: currentWorkspace,
      });

      // 更新列表中的图片信息
      setWorkspaceImages((prev) =>
        prev.map((img) =>
          img.path === renamingImage.path ? response.image : img
        )
      );

      // 如果图片被选中，更新选中列表中的信息
      setSelectedImages((prev) =>
        prev.map((img) =>
          img.path === renamingImage.path ? response.image : img
        )
      );

      handleCancelRename();
      // 刷新列表以确保数据同步
      await loadWorkspaceImages();
    } catch (err) {
      setError(err instanceof Error ? err.message : '重命名图片失败');
    }
  };

  // 处理文件选择并立即上传
  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const newFiles = Array.from(e.target.files);
      setError(null);
      setUploading(true);
      
      try {
        // 上传所有文件
        await Promise.all(
          newFiles.map((file) => imageService.uploadImage(file, currentWorkspace))
        );
        
        // 上传成功后刷新图片列表
        await loadWorkspaceImages();
      } catch (err) {
        setError(err instanceof Error ? err.message : '部分图片上传失败');
      } finally {
        setUploading(false);
        // 清空 input 值，允许重复选择同一个文件
        e.target.value = '';
      }
    }
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
    setElapsedTime(0);
    
    // 初始化结果数组
    const initialResults = Array.from({ length: generateCount }, () => ({
      data: null,
      status: 'pending' as const,
      elapsedTime: null,
    }));
    setResults(initialResults);

    startTimeRef.current = Date.now();

    // 启动计时器
    timerRef.current = window.setInterval(() => {
      if (startTimeRef.current !== null) {
        const elapsed = Math.floor((Date.now() - startTimeRef.current) / 1000);
        setElapsedTime(elapsed);
      }
    }, 1000);

    try {
      // 获取选中图片的路径列表
      const imagePaths = selectedImages.map((img) => img.path);

      // 循环调用接口生成多张图片
      const generatePromises = initialResults.map(async (_, index) => {
        const itemStartTime = Date.now();
        
        // 更新状态为生成中
        setResults((prev) => {
          const newResults = [...prev];
          newResults[index] = { ...newResults[index], status: 'generating' };
          return newResults;
        });

        try {
          const response = await imageService.generateImage({
            prompt: prompt.trim(),
            images: imagePaths.length > 0 ? imagePaths : undefined,
            workspace: currentWorkspace,
          });

          const itemElapsedTime = Math.floor((Date.now() - itemStartTime) / 1000);

          // 更新成功状态
          setResults((prev) => {
            const newResults = [...prev];
            newResults[index] = {
              data: response,
              status: 'success',
              elapsedTime: itemElapsedTime,
            };
            return newResults;
          });

          return response;
        } catch (err) {
          const itemElapsedTime = Math.floor((Date.now() - itemStartTime) / 1000);
          const errorMessage = err instanceof Error ? err.message : '生成失败，请重试';
          
          // 更新失败状态
          setResults((prev) => {
            const newResults = [...prev];
            newResults[index] = {
              data: null,
              status: 'error',
              elapsedTime: itemElapsedTime,
              error: errorMessage,
            };
            return newResults;
          });

          throw err;
        }
      });

      // 等待所有生成完成（使用 allSettled 确保所有请求都完成，即使有失败的）
      const settledResults = await Promise.allSettled(generatePromises);
      
      // 检查是否有成功的生成
      const hasSuccess = settledResults.some(result => result.status === 'fulfilled');
      
      if (hasSuccess) {
        // 生成成功后刷新图片列表
        loadWorkspaceImages();
      }
      
      // 如果有失败的，显示错误信息
      const failedCount = settledResults.filter(result => result.status === 'rejected').length;
      if (failedCount > 0) {
        setError(`有 ${failedCount} 张图片生成失败`);
      }
    } catch (err) {
      // 整体错误处理
      const errorMessage = err instanceof Error ? err.message : '生成过程出错';
      setError(errorMessage);
    } finally {
      setLoading(false);
      clearTimer();
    }
  };

  // 格式化时间
  const formatTime = (seconds: number) => {
    return `${seconds}s`;
  };

  return (
    <div className="image-generator">
      <div className="image-generator-layout">
        {/* 左侧边栏 */}
        <div className="image-list-sidebar" ref={imageListRef}>
          {/* 工作区选择器 */}
          <div className="sidebar-workspace-section">
            <div className="workspace-selector-wrapper" ref={workspaceDropdownRef}>
              <div className="workspace-selector">
                <button
                  type="button"
                  onClick={() => setShowWorkspaceDropdown(!showWorkspaceDropdown)}
                  className="workspace-trigger"
                  disabled={workspaceLoading}
                >
                  <span className="workspace-label">工作区</span>
                  <span className="workspace-current">{currentWorkspace}</span>
                  <span className={`workspace-arrow ${showWorkspaceDropdown ? 'open' : ''}`}>
                    ▼
                  </span>
                </button>
                
                {showWorkspaceDropdown && (
                  <div className="workspace-dropdown">
                    {workspaces.map((ws) => (
                      <button
                        key={ws.name}
                        type="button"
                        onClick={async () => {
                          try {
                            // 调用 API 切换工作区
                            await workspaceService.setCurrentWorkspace({ name: ws.name });
                            setCurrentWorkspace(ws.name);
                            setShowWorkspaceDropdown(false);
                            setSelectedImages([]); // 切换工作区时清空选中的图片
                            // 重新加载工作区列表以更新 is_current 状态（不触发图片加载，因为 currentWorkspace 已经设置）
                            const response = await workspaceService.listWorkspaces();
                            setWorkspaces(response.workspaces);
                          } catch (err) {
                            console.error('切换工作区失败:', err);
                            setError('切换工作区失败: ' + (err instanceof Error ? err.message : '未知错误'));
                          }
                        }}
                        className={`workspace-option ${
                          ws.name === currentWorkspace || ws.is_current ? 'active' : ''
                        }`}
                      >
                        <span>{ws.name}</span>
                        {(ws.name === currentWorkspace || ws.is_current) && (
                          <span className="workspace-check">✓</span>
                        )}
                      </button>
                    ))}
                    <div className="workspace-divider"></div>
                    <button
                      type="button"
                      onClick={() => {
                        setShowWorkspaceDropdown(false);
                        setShowWorkspaceModal(true);
                      }}
                      className="workspace-option workspace-option-action"
                    >
                      <span>+ 新建工作区</span>
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* 图片列表 */}
          <div className={`image-list-header ${showImageList ? 'expanded' : ''}`}>
            <button
              type="button"
              onClick={() => setShowImageList(!showImageList)}
              className="image-list-toggle"
            >
              <span>工作区图片</span>
              <span className={`image-list-arrow ${showImageList ? 'open' : ''}`}>
                ▼
              </span>
            </button>
          </div>
          
          {showImageList && (
            <div className="image-list-content">
              {/* 上传图片区域 */}
              <div className={`image-list-upload ${(loading || uploading) ? 'disabled' : ''}`}>
                <label htmlFor="image-upload-input" className="image-upload-label">
                  {uploading ? (
                    <span className="upload-status-text">上传中...</span>
                  ) : (
                    <>
                      <span className="upload-icon">📤</span>
                      <span>上传图片</span>
                    </>
                  )}
                </label>
                <input
                  type="file"
                  id="image-upload-input"
                  accept="image/*"
                  multiple
                  onChange={handleFileChange}
                  disabled={loading || uploading}
                  className="image-upload-input"
                />
              </div>

              {/* 图片列表 */}
              {imagesLoading ? (
                <div className="image-list-loading">加载中...</div>
              ) : workspaceImages.length === 0 ? (
                <div className="image-list-empty">暂无图片，点击上方上传</div>
              ) : (
                <div className="image-list-items">
                  {workspaceImages.map((image, index) => (
                    <div
                      key={index}
                      className={`image-list-item ${isImageSelected(image) ? 'selected' : ''}`}
                      onClick={() => handleSelectImage(image)}
                      onMouseEnter={() => setPreviewImage(image)}
                      onMouseLeave={() => setPreviewImage(null)}
                    >
                      <img
                        src={image.thumbnail_url || image.url}
                        alt={image.name}
                        className="image-list-thumbnail"
                      />
                      <div className="image-list-item-info">
                        {renamingImage?.path === image.path ? (
                          <div className="image-list-item-rename">
                            <div className="image-rename-input-wrapper">
                              <input
                                type="text"
                                value={newImageName}
                                onChange={(e) => setNewImageName(e.target.value)}
                                onKeyDown={(e) => {
                                  if (e.key === 'Enter') {
                                    handleConfirmRename();
                                  } else if (e.key === 'Escape') {
                                    handleCancelRename();
                                  }
                                }}
                                onClick={(e) => e.stopPropagation()}
                                className="image-rename-input"
                                autoFocus
                              />
                              {imageExtension && (
                                <span className="image-rename-extension">{imageExtension}</span>
                              )}
                            </div>
                            <div className="image-rename-actions">
                              <button
                                type="button"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleConfirmRename();
                                }}
                                className="image-rename-btn image-rename-confirm"
                                title="确认"
                              >
                                ✓
                              </button>
                              <button
                                type="button"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleCancelRename();
                                }}
                                className="image-rename-btn image-rename-cancel"
                                title="取消"
                              >
                                ×
                              </button>
                            </div>
                          </div>
                        ) : (
                          <>
                            <div className="image-list-item-name" title={image.name}>
                              {image.name}
                            </div>
                            <div className="image-list-item-meta">
                              {formatDateOnlyToBeijing(image.updated)}
                            </div>
                          </>
                        )}
                      </div>
                      {isImageSelected(image) && (
                        <div className="image-list-item-check">✓</div>
                      )}
                      {renamingImage?.path !== image.path && (
                        <div className="image-list-item-actions">
                          <button
                            type="button"
                            onClick={(e) => handleStartRename(image, e)}
                            className="image-action-btn image-rename-action"
                            title="重命名"
                          >
                            <Edit size={14} />
                          </button>
                          <button
                            type="button"
                            onClick={(e) => handleDeleteImage(image, e)}
                            className="image-action-btn image-delete-action"
                            title="删除"
                          >
                            <Trash2 size={14} />
                          </button>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>

        {/* 主内容区 */}
        <div className="container">
          <div className="main-header">
            <h1>AI 图片生成器</h1>
            <p className="subtitle">使用 Gemini 3 Pro Image Preview 模型生成和编辑图片</p>
          </div>

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

          {/* 生成数量选择 */}
          <div className="form-group">
            <label htmlFor="generate-count">生成数量</label>
            <div className="generate-count-selector">
              {[1, 2, 3].map((count) => (
                <button
                  key={count}
                  type="button"
                  onClick={() => setGenerateCount(count)}
                  disabled={loading}
                  className={`count-option ${generateCount === count ? 'active' : ''}`}
                >
                  {count} 张
                </button>
              ))}
            </div>
            <p className="form-hint">选择要生成的图片数量，最多可一次性生成 3 张</p>
          </div>

          {/* 已选中的图片 */}
          {selectedImages.length > 0 && (
            <div className="form-group">
              <label>已选中的参考图片（{selectedImages.length}）</label>
              <div className="selected-images-list">
                {selectedImages.map((image) => (
                  <div key={image.path} className="selected-image-item">
                    <img
                      src={image.url}
                      alt={image.name}
                      className="selected-image-preview"
                    />
                    <div className="selected-image-info">
                      <div className="selected-image-name" title={image.name}>
                        {image.name}
                      </div>
                    </div>
                    <button
                      type="button"
                      onClick={() => removeSelectedImage(image.path)}
                      className="remove-selected-btn"
                      disabled={loading}
                      title="移除"
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* 错误提示 */}
          {error && <div className="error-message">{error}</div>}

          {/* 提交按钮 */}
          <button
            type="submit"
            disabled={loading || !prompt.trim()}
            className="submit-btn"
          >
            {(() => {
              if (loading) {
                const completedCount = results.filter(r => r.status === 'success' || r.status === 'error').length;
                const currentCount = completedCount + results.filter(r => r.status === 'generating').length;
                if (generateCount > 1) {
                  return `生成中 ${currentCount}/${generateCount} (${formatTime(elapsedTime)})...`;
                }
                return `生成中 (${formatTime(elapsedTime)})...`;
              }
              return `生成图片${generateCount > 1 ? ` (${generateCount} 张)` : ''}`;
            })()}
          </button>
        </form>

        {/* 生成结果 */}
        {(results.length > 0 || result) && (
          <div className="result-section">
            <div className="result-header">
              <h2>生成结果</h2>
            </div>
            
            {/* 多张生成结果 */}
            {results.length > 0 && (
              <div className="generated-results-list">
                {results.map((resultItem, resultIndex) => (
                  <div key={resultIndex} className="generated-result-item">
                    <div className="result-item-header">
                      <h3>第 {resultIndex + 1} 张</h3>
                      <div className="result-item-status">
                        {resultItem.status === 'pending' && (
                          <span className="status-pending">等待中</span>
                        )}
                        {resultItem.status === 'generating' && (
                          <span className="status-generating">生成中...</span>
                        )}
                        {resultItem.status === 'success' && resultItem.elapsedTime !== null && (
                          <span className="status-success">
                            完成 ({formatTime(resultItem.elapsedTime)})
                          </span>
                        )}
                        {resultItem.status === 'error' && (
                          <span className="status-error">
                            失败 {resultItem.elapsedTime !== null && `(${formatTime(resultItem.elapsedTime)})`}
                          </span>
                        )}
                      </div>
                    </div>
                    
                    {resultItem.status === 'error' && resultItem.error && (
                      <div className="result-error-message">{resultItem.error}</div>
                    )}
                    
                    {resultItem.data && (
                      <div className="generated-content">
                        {resultItem.data.parts.map((part, partIndex) => (
                          <div key={partIndex} className={`result-part result-part-${part.type}`}>
                            {part.type === 'text' && (
                              <div className="description">
                                <p>{part.text}</p>
                              </div>
                            )}
                            
                            {part.type === 'image' && part.image && (
                              <div className="generated-image-item">
                                <img
                                  src={part.image.url || (part.image.data ? `data:${part.image.mimeType};base64,${part.image.data}` : '')}
                                  alt={`生成的图片 ${resultIndex + 1}-${partIndex + 1}`}
                                  className="generated-image"
                                />
                                <a
                                  href={part.image.url || (part.image.data ? `data:${part.image.mimeType};base64,${part.image.data}` : '#')}
                                  download={part.image.url ? undefined : `generated-${Date.now()}-${resultIndex}-${partIndex}.png`}
                                  target={part.image.url ? '_blank' : undefined}
                                  rel={part.image.url ? 'noopener noreferrer' : undefined}
                                  className="download-btn"
                                >
                                  {part.image.url ? '查看原图' : '下载图片'}
                                </a>
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
            
            {/* 兼容旧的单个结果展示 */}
            {results.length === 0 && result && (
              <div className="generated-content">
                {result.parts.map((part, index) => (
                  <div key={index} className={`result-part result-part-${part.type}`}>
                    {part.type === 'text' && (
                      <div className="description">
                        <p>{part.text}</p>
                      </div>
                    )}
                    
                    {part.type === 'image' && part.image && (
                      <div className="generated-image-item">
                        <img
                          src={part.image.url || (part.image.data ? `data:${part.image.mimeType};base64,${part.image.data}` : '')}
                          alt={`生成的图片 ${index + 1}`}
                          className="generated-image"
                        />
                        <a
                          href={part.image.url || (part.image.data ? `data:${part.image.mimeType};base64,${part.image.data}` : '#')}
                          download={part.image.url ? undefined : `generated-${Date.now()}-${index}.png`}
                          target={part.image.url ? '_blank' : undefined}
                          rel={part.image.url ? 'noopener noreferrer' : undefined}
                          className="download-btn"
                        >
                          {part.image.url ? '查看原图' : '下载图片'}
                        </a>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
        </div>
      </div>

      {/* Workspace 管理模态框 */}
      {showWorkspaceModal && (
        <div className="modal-overlay" onClick={() => setShowWorkspaceModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h2>管理工作区</h2>
            
            {/* 创建工作区 */}
            <div className="modal-section">
              <h3>创建工作区</h3>
              <div className="modal-input-group">
                <input
                  type="text"
                  value={newWorkspaceName}
                  onChange={(e) => setNewWorkspaceName(e.target.value)}
                  placeholder="输入工作区名称"
                  className="modal-input"
                  onKeyPress={(e) => {
                    if (e.key === 'Enter') {
                      handleCreateWorkspace();
                    }
                  }}
                />
                <button
                  type="button"
                  onClick={handleCreateWorkspace}
                  disabled={workspaceLoading || !newWorkspaceName.trim()}
                  className="modal-btn"
                >
                  创建
                </button>
              </div>
            </div>

            {/* 工作区列表 */}
            <div className="modal-section">
              <h3>工作区列表</h3>
              <div className="workspace-list">
                {workspaces.length === 0 ? (
                  <div className="workspace-empty">暂无工作区</div>
                ) : (
                  workspaces.map((ws) => (
                    <div key={ws.name} className="workspace-item">
                      <span className="workspace-name">{ws.name}</span>
                      <button
                        type="button"
                        onClick={() => handleDeleteWorkspace(ws.name)}
                        disabled={workspaceLoading || workspaces.length === 1}
                        className="delete-workspace-btn"
                      >
                        删除
                      </button>
                    </div>
                  ))
                )}
              </div>
            </div>

            <button
              type="button"
              onClick={() => setShowWorkspaceModal(false)}
              className="modal-close-btn"
            >
              关闭
            </button>
          </div>
        </div>
      )}

      {/* 悬浮图片预览 */}
      {previewImage && (
        <div className="image-preview-overlay" key={previewImage.path}>
          <div className="image-preview-section">
            <div className="image-preview-header">
              <h3>{previewImage.name}</h3>
              <div className="image-preview-meta">
                <span>{formatDateTimeToBeijing(previewImage.updated)}</span>
                {previewImage.size > 0 && (
                  <span className="image-preview-size">
                    {(previewImage.size / 1024).toFixed(2)} KB
                  </span>
                )}
              </div>
            </div>
            <div className="image-preview-content">
              <img
                src={previewImage.url}
                alt={previewImage.name}
                className="image-preview-large"
                loading="eager"
                onLoad={(e) => {
                  // 确保图片加载后布局稳定
                  e.currentTarget.style.opacity = '1';
                }}
                style={{ opacity: 0, transition: 'opacity 0.2s ease' }}
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
