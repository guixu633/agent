import { useState, useRef, useEffect } from 'react';
import { Edit, Trash2, RotateCcw, Settings, Send, Globe } from 'lucide-react';
import { imageService } from '@/services/image/imageService';
import { workspaceService } from '@/services/workspace/workspaceService';
import type { ImageGenerateResponse, ImageInfo } from '@/types/image';
import type { Workspace } from '@/types/workspace';
import { formatDateTimeToBeijing } from '@/utils/date';
import './ImageGenerator.css';

// 辅助函数：从路径获取文件名
const getFileNameFromPath = (path: string) => {
  const parts = path.split('/');
  return parts[parts.length - 1];
};

export function ImageGenerator() {
  const [prompt, setPrompt] = useState('');
  const [selectedImages, setSelectedImages] = useState<ImageInfo[]>([]); // 从列表中选择的图片
  const [uploading, setUploading] = useState(false); // 上传状态
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<ImageGenerateResponse | null>(null);
  const [generateCount, setGenerateCount] = useState(1); // 生成数量，默认1张，最多3张
  const [enableWebSearch, setEnableWebSearch] = useState(false); // 是否启用联网搜索，默认不开启
  const [results, setResults] = useState<Array<{
    data: ImageGenerateResponse | null;
    status: 'pending' | 'generating' | 'success' | 'error';
    elapsedTime: number | null;
    error?: string;
  }>>([]);
  const [expandedTexts, setExpandedTexts] = useState<Set<number>>(new Set()); // 展开的文本消息结果索引，默认全部折叠
  
  // 回溯的图片信息（用于展示对话历史）
  const [restoredImageInfo, setRestoredImageInfo] = useState<ImageInfo | null>(null);
  
  // Workspace 相关状态
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [currentWorkspace, setCurrentWorkspace] = useState<string>('');
  const [showWorkspaceModal, setShowWorkspaceModal] = useState(false);
  const [showWorkspaceDropdown, setShowWorkspaceDropdown] = useState(false);
  const [newWorkspaceName, setNewWorkspaceName] = useState('');
  const [workspaceLoading, setWorkspaceLoading] = useState(false);
  const workspaceDropdownRefInSettings = useRef<HTMLDivElement>(null);
  const isLoadingRef = useRef(false); // 防止重复加载
  
  // 设置菜单相关状态
  const [showSettingsMenu, setShowSettingsMenu] = useState(false);
  const settingsMenuRef = useRef<HTMLDivElement>(null);
  
  // 生成数量下拉列表相关状态
  const [showCountDropdown, setShowCountDropdown] = useState(false);
  const countDropdownRef = useRef<HTMLDivElement>(null);
  
  // 图片列表相关状态
  const [workspaceImages, setWorkspaceImages] = useState<ImageInfo[]>([]);
  const [imagesLoading, setImagesLoading] = useState(false);
  const imageListRef = useRef<HTMLDivElement>(null);
  
  // 可拖动分隔线相关状态
  const [sidebarWidth, setSidebarWidth] = useState(650); // 左侧边栏宽度
  const [isResizing, setIsResizing] = useState(false);
  const resizeStartXRef = useRef<number>(0);
  const resizeStartWidthRef = useRef<number>(0);
  
  // 重命名相关状态
  const [renamingImage, setRenamingImage] = useState<ImageInfo | null>(null);
  const [newImageName, setNewImageName] = useState('');
  const [imageExtension, setImageExtension] = useState(''); // 保存文件扩展名
  
  // 图片预览相关状态
  const [previewImage, setPreviewImage] = useState<ImageInfo | null>(null);
  const previewElementRef = useRef<HTMLDivElement | null>(null);
  
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

  // 拖动分隔线处理
  const handleResizeStart = (e: React.MouseEvent) => {
    e.preventDefault();
    setIsResizing(true);
    resizeStartXRef.current = e.clientX;
    resizeStartWidthRef.current = sidebarWidth;
  };

  useEffect(() => {
    const handleResizeMove = (e: MouseEvent) => {
      if (!isResizing) return;
      
      const deltaX = e.clientX - resizeStartXRef.current;
      const newWidth = resizeStartWidthRef.current + deltaX;
      
      // 限制最小和最大宽度
      const minWidth = 300;
      const maxWidth = window.innerWidth - 600; // 右侧至少保留600px
      
      setSidebarWidth(Math.max(minWidth, Math.min(maxWidth, newWidth)));
    };

    const handleResizeEnd = () => {
      setIsResizing(false);
    };

    if (isResizing) {
      document.addEventListener('mousemove', handleResizeMove);
      document.addEventListener('mouseup', handleResizeEnd);
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
    }

    return () => {
      document.removeEventListener('mousemove', handleResizeMove);
      document.removeEventListener('mouseup', handleResizeEnd);
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
    };
  }, [isResizing]);

  // 点击外部关闭下拉菜单
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        workspaceDropdownRefInSettings.current &&
        !workspaceDropdownRefInSettings.current.contains(event.target as Node)
      ) {
        setShowWorkspaceDropdown(false);
      }
      if (
        settingsMenuRef.current &&
        !settingsMenuRef.current.contains(event.target as Node)
      ) {
        setShowSettingsMenu(false);
      }
      if (
        countDropdownRef.current &&
        !countDropdownRef.current.contains(event.target as Node)
      ) {
        setShowCountDropdown(false);
      }
    };

    if (showWorkspaceDropdown || showSettingsMenu || showCountDropdown) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showWorkspaceDropdown, showSettingsMenu, showCountDropdown]);

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
  const handleDeleteImage = async (imageOrPath: ImageInfo | string, e?: React.MouseEvent) => {
    if (e) e.stopPropagation(); // 阻止触发选择事件
    
    const path = typeof imageOrPath === 'string' ? imageOrPath : imageOrPath.path;
    
    try {
      await imageService.deleteImage({ path });
      // 从列表中移除（使用 diff 更新，避免重新加载导致折叠）
      setWorkspaceImages((prev) => prev.filter((img) => img.path !== path));
      // 如果图片被选中，也从选中列表中移除
      setSelectedImages((prev) => prev.filter((img) => img.path !== path));
      
      // 更新生成结果列表 (从 results 中移除包含该图片的项)
      setResults((prev) => prev.filter(item => {
         const hasDeletedImage = item.data?.parts.some(p => p.type === 'image' && p.image?.path === path);
         return !hasDeletedImage; 
      }));
      
      // 更新 result (如果存在)
      if (result) {
          const hasDeletedImage = result.parts.some(p => p.type === 'image' && p.image?.path === path);
          if (hasDeletedImage) {
              setResult(null);
          }
      }
      
      // 不再调用 loadWorkspaceImages()，避免列表折叠
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除图片失败');
      // 如果删除失败，重新加载列表以确保数据同步
      await loadWorkspaceImages();
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

      // 更新列表中的图片信息（使用 diff 更新，避免重新加载导致折叠）
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
      
      // 更新 results 中的图片信息
      setResults((prev) => 
        prev.map((item) => {
           if (!item.data) return item;
           
           const hasRenamedImage = item.data.parts.some(p => p.type === 'image' && p.image?.path === renamingImage.path);
           
           if (hasRenamedImage) {
               const newParts = item.data.parts.map(p => {
                 if (p.type === 'image' && p.image?.path === renamingImage.path) {
                   return {
                     ...p,
                     image: {
                       ...p.image!,
                       path: response.image.path,
                       url: response.image.url,
                     }
                   };
                 }
                 return p;
               });
               return {
                 ...item,
                 data: { ...item.data, parts: newParts }
               };
           }
           return item;
        })
      );
      
      // 更新 result (如果存在)
      if (result) {
           const hasRenamedImage = result.parts.some(p => p.type === 'image' && p.image?.path === renamingImage.path);
           if (hasRenamedImage) {
               const newParts = result.parts.map(p => {
                 if (p.type === 'image' && p.image?.path === renamingImage.path) {
                   return {
                     ...p,
                     image: {
                       ...p.image!,
                       path: response.image.path,
                       url: response.image.url,
                     }
                   };
                 }
                 return p;
               });
               setResult({ ...result, parts: newParts });
           }
      }

      handleCancelRename();
      // 不再调用 loadWorkspaceImages()，避免列表折叠
    } catch (err) {
      setError(err instanceof Error ? err.message : '重命名图片失败');
    }
  };

  // 批量上传文件
  const uploadFiles = async (files: File[]) => {
    if (files.length === 0) return;

    setError(null);
    setUploading(true);
    
    try {
      // 上传所有文件
      const uploadResults = await Promise.all(
        files.map((file) => imageService.uploadImage(file, currentWorkspace))
      );
      
      // 先添加临时信息到列表（避免列表折叠）
      const tempImages: ImageInfo[] = uploadResults.map((result, index) => ({
        path: result.path,
        url: result.url,
        thumbnail_url: result.url, // 暂时使用原图作为缩略图
        name: files[index].name,
        size: files[index].size,
        updated: new Date().toISOString(),
        source_type: 'upload' as const,
      }));
      setWorkspaceImages((prev) => [...tempImages, ...prev]);
      
      // 然后在后台静默刷新列表以获取完整信息（包括缩略图等）
      // 使用 setTimeout 避免阻塞 UI
      setTimeout(async () => {
        try {
          await loadWorkspaceImages();
        } catch (err) {
          console.error('后台刷新图片列表失败:', err);
          // 刷新失败时保持临时信息
        }
      }, 300);
    } catch (err) {
      setError(err instanceof Error ? err.message : '部分图片上传失败');
      // 上传失败时刷新列表以确保数据同步
      await loadWorkspaceImages();
    } finally {
      setUploading(false);
    }
  };

  // 处理文件选择并立即上传
  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const newFiles = Array.from(e.target.files);
      await uploadFiles(newFiles);
      // 清空 input 值，允许重复选择同一个文件
      e.target.value = '';
    }
  };

  // 监听粘贴事件
  useEffect(() => {
    const handlePaste = async (e: ClipboardEvent) => {
      // ... existing code ...
      if (uploading) return;

      const items = e.clipboardData?.items;
      if (!items) return;

      const files: File[] = [];
      for (let i = 0; i < items.length; i++) {
        const item = items[i];
        if (item.kind === 'file' && item.type.startsWith('image/')) {
          const file = item.getAsFile();
          if (file) {
            // 如果是从剪贴板粘贴的图片，可能没有文件名或者文件名是通用的 "image.png"
            // 我们可以给它生成一个带时间戳的文件名，避免冲突（虽然后端可能会处理，但前端处理更好）
            // 这里我们暂时保留原文件名，如果需要唯一性，后端服务应该处理或前端生成
            // 实际上 getAsFile() 返回的 File 对象通常有默认名 "image.png"
            // 如果一次粘贴多张，可能都叫 image.png，这可能导致问题
            // 我们可以重命名一下
            if (file.name === 'image.png' || !file.name) {
                const timestamp = new Date().getTime();
                const ext = file.type.split('/')[1] || 'png';
                const newName = `pasted_image_${timestamp}_${i}.${ext}`;
                // File 对象的 name 属性是只读的，需要重新创建 File
                const newFile = new File([file], newName, { type: file.type });
                files.push(newFile);
            } else {
                files.push(file);
            }
          }
        }
      }

      if (files.length > 0) {
        e.preventDefault();
        await uploadFiles(files);
      }
    };

    window.addEventListener('paste', handlePaste);
    return () => {
      window.removeEventListener('paste', handlePaste);
    };
  }, [uploading, currentWorkspace]); // 依赖项

  // 监听快捷键：切换生成数量
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Cmd (Mac) 或 Ctrl (Win) + 数字键
      if (e.metaKey || e.ctrlKey) {
        switch (e.key) {
          case '1':
            e.preventDefault();
            setGenerateCount(1);
            break;
          case '2':
            e.preventDefault();
            setGenerateCount(2);
            break;
          case '3':
            e.preventDefault();
            setGenerateCount(3);
            break;
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, []);

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
    setRestoredImageInfo(null); // 清空回溯状态
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
            enable_web_search: enableWebSearch,
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

  // 画廊重命名辅助函数
  const handleGalleryRename = (path: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const existingInfo = workspaceImages.find(img => img.path === path);
    const imageInfo = existingInfo || {
        path: path,
        name: getFileNameFromPath(path),
        url: '',
        thumbnail_url: '',
        size: 0,
        updated: new Date().toISOString(),
        source_type: 'upload' as const
    };
    handleStartRename(imageInfo, e);
  };

  // 回溯生成参数
  const handleRestoreContext = async (image: ImageInfo, e: React.MouseEvent) => {
    e.stopPropagation();
    
    console.log('回溯图片信息:', {
      path: image.path,
      prompt: image.prompt,
      ref_images: image.ref_images,
      message_list: image.message_list,
      source_type: image.source_type
    });
    
    // 填充 prompt
    if (image.prompt) {
        setPrompt(image.prompt);
    } else {
        setError('该图片没有保存生成时的提示词信息');
        return;
    }
    
    // 填充引用图片
    if (image.ref_images && image.ref_images.length > 0) {
        // 找到对应的图片对象
        const refImages = workspaceImages.filter(img => image.ref_images?.includes(img.path));
        setSelectedImages(refImages);
        
        // 如果有些引用图片不在当前列表中，提示用户
        if (refImages.length < image.ref_images.length) {
            const missingCount = image.ref_images.length - refImages.length;
            setError(`已回溯参数，但有 ${missingCount} 张引用图片不在当前工作区列表中`);
        } else {
            // 清空之前的错误信息
            setError(null);
        }
    } else {
        setSelectedImages([]);
        setError(null);
    }
    
    // 如果有 message_list，保存图片信息以便在生成结果区域展示
    if (image.message_list && image.message_list.length > 0) {
        setRestoredImageInfo(image);
    } else {
        setRestoredImageInfo(null);
    }
    
    // 滚动到输入框位置，方便用户查看
    setTimeout(() => {
      const promptElement = document.getElementById('prompt');
      if (promptElement) {
        promptElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
        promptElement.focus();
      }
    }, 100);
  };

  return (
    <div className="image-generator">
      {/* 全局设置按钮 - 固定在页面右上角 */}
      <div className="settings-button-wrapper" ref={settingsMenuRef}>
        <button
          type="button"
          onClick={() => setShowSettingsMenu(!showSettingsMenu)}
          className="settings-button"
          title="设置"
        >
          <Settings size={20} />
        </button>
        
        {showSettingsMenu && (
          <div className="settings-menu">
            <div className="settings-menu-header">设置</div>
            <div className="settings-menu-item">
              <div className="settings-menu-label">工作区切换</div>
              <div className="workspace-selector-wrapper-inline" ref={workspaceDropdownRefInSettings}>
                <div className="workspace-selector">
                  <button
                    type="button"
                    onClick={() => {
                      setShowWorkspaceDropdown(!showWorkspaceDropdown);
                    }}
                    className="workspace-trigger-inline"
                    disabled={workspaceLoading}
                  >
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
                              await workspaceService.setCurrentWorkspace({ name: ws.name });
                              setCurrentWorkspace(ws.name);
                              setShowWorkspaceDropdown(false);
                              setShowSettingsMenu(false);
                              setSelectedImages([]);
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
                          setShowSettingsMenu(false);
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
          </div>
        )}
      </div>
      
      <div className="image-generator-layout">
        {/* 左侧边栏 */}
        <div 
          className="image-list-sidebar" 
          ref={imageListRef}
          style={{ width: `${sidebarWidth}px` }}
        >
          {/* 图片列表 */}
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
                  {/* 对图片列表进行排序：选中的图片置顶，其余按原顺序（时间倒序） */}
                  {[...workspaceImages]
                    .sort((a, b) => {
                      const aSelected = isImageSelected(a);
                      const bSelected = isImageSelected(b);
                      if (aSelected && !bSelected) return -1;
                      if (!aSelected && bSelected) return 1;
                      return 0;
                    })
                    .map((image, index) => (
                    <div
                      key={image.path || index}
                      ref={(el) => {
                        if (previewImage?.path === image.path) {
                          previewElementRef.current = el;
                        }
                      }}
                      className={`image-list-item ${isImageSelected(image) ? 'selected' : ''} ${image.source_type === 'generate' ? 'generated' : ''}`}
                      onClick={() => handleSelectImage(image)}
                      onMouseEnter={(e) => {
                        setPreviewImage(image);
                        previewElementRef.current = e.currentTarget;
                      }}
                      onMouseLeave={() => setPreviewImage(null)}
                    >
                      <img
                        src={image.thumbnail_url || image.url}
                        alt={image.name}
                        className="image-list-thumbnail"
                        title={image.name}
                      />
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
                            <Edit size={12} />
                          </button>
                          <button
                            type="button"
                            onClick={(e) => handleDeleteImage(image, e)}
                            className="image-action-btn image-delete-action"
                            title="删除"
                          >
                            <Trash2 size={12} />
                          </button>
                          {image.source_type === 'generate' && (
                            <button
                                type="button"
                                onClick={(e) => handleRestoreContext(image, e)}
                                className="image-action-btn image-restore-action"
                                title="回溯参数"
                            >
                                <RotateCcw size={12} />
                            </button>
                          )}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
        </div>

        {/* 可拖动分隔线 */}
        <div 
          className="resizer"
          onMouseDown={handleResizeStart}
          style={{ cursor: 'col-resize' }}
        />

        {/* 主内容区 */}
        <div className="container">

        <form onSubmit={handleSubmit} className="generator-form">
          {/* 提示词输入 */}
          <div className="form-group">
            <label htmlFor="prompt">提示词</label>
            <div className="prompt-input-wrapper">
              <textarea
                id="prompt"
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                onKeyDown={(e) => {
                  // Cmd+Enter (Mac) 或 Ctrl+Enter (Windows/Linux) 快速提交
                  if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                    e.preventDefault();
                    if (!loading && prompt.trim()) {
                      // 触发表单提交
                      const form = e.currentTarget.closest('form') as HTMLFormElement;
                      if (form) {
                        form.requestSubmit();
                      }
                    }
                  }
                }}
                placeholder="请描述你想要生成的图片，例如：帮我把图片修改为羊毛毡的可爱风格，短手短脚的那种可爱玩偶的感觉"
                rows={4}
                disabled={loading}
                className="prompt-textarea"
              />
              {/* 生成数量下拉列表 */}
              <div className="generate-count-dropdown-wrapper" ref={countDropdownRef}>
                <button
                  type="button"
                  onClick={() => setShowCountDropdown(!showCountDropdown)}
                  disabled={loading}
                  className="generate-count-trigger"
                  title="选择生成数量"
                >
                  <span className="generate-count-value">{generateCount}</span>
                  <span className={`generate-count-arrow ${showCountDropdown ? 'open' : ''}`}>
                    ▼
                  </span>
                </button>
                {showCountDropdown && (
                  <div className="generate-count-dropdown">
                    {[1, 2, 3].map((count) => (
                      <button
                        key={count}
                        type="button"
                        onClick={() => {
                          setGenerateCount(count);
                          setShowCountDropdown(false);
                        }}
                        className={`generate-count-option ${generateCount === count ? 'active' : ''}`}
                      >
                        <span>{count} 张</span>
                        {generateCount === count && (
                          <span className="generate-count-check">✓</span>
                        )}
                      </button>
                    ))}
                  </div>
                )}
              </div>
              {/* 启用联网搜索开关 */}
              <button
                type="button"
                onClick={() => !loading && setEnableWebSearch(!enableWebSearch)}
                disabled={loading}
                className={`web-search-toggle ${enableWebSearch ? 'active' : ''} ${loading ? 'disabled' : ''}`}
                title={enableWebSearch ? "已启用联网搜索" : "启用联网搜索"}
              >
                <Globe size={18} />
              </button>
              <button
                type="submit"
                disabled={loading || !prompt.trim()}
                className="prompt-submit-btn"
                title={(() => {
                  if (loading) {
                    const completedCount = results.filter(r => r.status === 'success' || r.status === 'error').length;
                    const currentCount = completedCount + results.filter(r => r.status === 'generating').length;
                    if (generateCount > 1) {
                      return `生成中 ${currentCount}/${generateCount} (${formatTime(elapsedTime)})...`;
                    }
                    return `生成中 (${formatTime(elapsedTime)})...`;
                  }
                  return `生成图片${generateCount > 1 ? ` (${generateCount} 张)` : ''} (Ctrl+Enter 或 Cmd+Enter)`;
                })()}
              >
                {(() => {
                  if (loading) {
                    return <span className="submit-btn-spinner">⏳</span>;
                  }
                  return <Send size={18} className="submit-btn-icon" />;
                })()}
              </button>
            </div>
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
        </form>

        {/* 生成结果 */}
        {(results.length > 0 || result || restoredImageInfo) && (
          <div className="result-section">
            <div className="result-header">
              <h2>{restoredImageInfo ? '回溯的对话历史' : '生成结果'}</h2>
            </div>
            
            {/* 顶部图片画廊 */}
            <div className="generated-images-gallery">
              {results.map((resultItem, index) => {
                const imagePart = resultItem.data?.parts.find(p => p.type === 'image');
                
                return (
                  <div key={`gallery-${index}`} className="gallery-image-card">
                    <div className="gallery-image-wrapper">
                      {resultItem.status === 'success' && imagePart && imagePart.image ? (
                        <>
                          <img 
                            src={imagePart.image.url || (imagePart.image.data ? `data:${imagePart.image.mimeType};base64,${imagePart.image.data}` : '')}
                            alt={`生成的图片 ${index + 1}`}
                            className="gallery-image"
                          />
                          {renamingImage?.path === imagePart.image.path ? (
                             <div className="gallery-rename-input-wrapper" onClick={e => e.stopPropagation()}>
                                <div style={{display: 'flex', gap: '0.25rem', marginBottom: '0.5rem'}}>
                                   <input
                                     type="text"
                                     value={newImageName}
                                     onChange={(e) => setNewImageName(e.target.value)}
                                     onKeyDown={(e) => {
                                       if (e.key === 'Enter') handleConfirmRename();
                                       else if (e.key === 'Escape') handleCancelRename();
                                     }}
                                     className="image-rename-input"
                                     autoFocus
                                     style={{width: '100%'}}
                                     onClick={e => e.stopPropagation()}
                                   />
                                   {imageExtension && <span className="image-rename-extension">{imageExtension}</span>}
                                </div>
                                <div className="image-rename-actions">
                                    <button type="button" onClick={(e) => { e.stopPropagation(); handleConfirmRename(); }} className="image-rename-btn image-rename-confirm">✓</button>
                                    <button type="button" onClick={(e) => { e.stopPropagation(); handleCancelRename(); }} className="image-rename-btn image-rename-cancel">×</button>
                                </div>
                             </div>
                          ) : (
                              <div className="gallery-image-overlay">
                                <div className="gallery-actions" style={{display: 'flex', gap: '0.5rem'}}>
                                    <button
                                      type="button"
                                      onClick={(e) => handleGalleryRename(imagePart.image!.path!, e)}
                                      className="gallery-action-btn"
                                      title="重命名"
                                    >
                                      <Edit size={16} />
                                    </button>
                                    <button
                                      type="button"
                                      onClick={(e) => handleDeleteImage(imagePart.image!.path!, e)}
                                      className="gallery-action-btn delete-btn"
                                      title="删除"
                                    >
                                      <Trash2 size={16} />
                                    </button>
                                    <a
                                      href={imagePart.image.url || (imagePart.image.data ? `data:${imagePart.image.mimeType};base64,${imagePart.image.data}` : '#')}
                                      download={imagePart.image.url ? undefined : `generated-${Date.now()}-${index}.png`}
                                      target={imagePart.image.url ? '_blank' : undefined}
                                      rel="noopener noreferrer"
                                      className="gallery-action-btn"
                                      title="下载/查看原图"
                                    >
                                      ↓
                                    </a>
                                </div>
                              </div>
                          )}
                        </>
                      ) : (
                        <div className="gallery-image-status">
                          {resultItem.status === 'pending' && (
                            <span className="gallery-status-text">等待生成...</span>
                          )}
                          {resultItem.status === 'generating' && (
                             <div className="status-generating">生成中...</div>
                          )}
                          {resultItem.status === 'error' && (
                             <div className="gallery-error-text">生成失败</div>
                          )}
                        </div>
                      )}
                    </div>
                    <div className="gallery-card-footer">
                       <div className="gallery-card-title">第 {index + 1} 张</div>
                       {resultItem.elapsedTime !== null && (
                         <div className="gallery-card-time">耗时: {formatTime(resultItem.elapsedTime)}</div>
                       )}
                    </div>
                  </div>
                );
              })}

              {/* 兼容旧的单个结果展示 */}
              {results.length === 0 && result && result.parts.find(p => p.type === 'image') && (
                <div className="gallery-image-card">
                   {(() => {
                     const part = result.parts.find(p => p.type === 'image');
                     if (!part || !part.image) return null;
                     return (
                       <>
                        <div className="gallery-image-wrapper">
                          <img 
                            src={part.image.url || (part.image.data ? `data:${part.image.mimeType};base64,${part.image.data}` : '')}
                            alt="生成的图片"
                            className="gallery-image"
                          />
                          {renamingImage?.path === part.image.path ? (
                             <div className="gallery-rename-input-wrapper" onClick={e => e.stopPropagation()}>
                                <div style={{display: 'flex', gap: '0.25rem', marginBottom: '0.5rem'}}>
                                   <input
                                     type="text"
                                     value={newImageName}
                                     onChange={(e) => setNewImageName(e.target.value)}
                                     onKeyDown={(e) => {
                                       if (e.key === 'Enter') handleConfirmRename();
                                       else if (e.key === 'Escape') handleCancelRename();
                                     }}
                                     className="image-rename-input"
                                     autoFocus
                                     style={{width: '100%'}}
                                     onClick={e => e.stopPropagation()}
                                   />
                                   {imageExtension && <span className="image-rename-extension">{imageExtension}</span>}
                                </div>
                                <div className="image-rename-actions">
                                    <button type="button" onClick={(e) => { e.stopPropagation(); handleConfirmRename(); }} className="image-rename-btn image-rename-confirm">✓</button>
                                    <button type="button" onClick={(e) => { e.stopPropagation(); handleCancelRename(); }} className="image-rename-btn image-rename-cancel">×</button>
                                </div>
                             </div>
                          ) : (
                              <div className="gallery-image-overlay">
                                <div className="gallery-actions" style={{display: 'flex', gap: '0.5rem'}}>
                                    <button
                                      type="button"
                                      onClick={(e) => handleGalleryRename(part.image!.path!, e)}
                                      className="gallery-action-btn"
                                      title="重命名"
                                    >
                                      <Edit size={16} />
                                    </button>
                                    <button
                                      type="button"
                                      onClick={(e) => handleDeleteImage(part.image!.path!, e)}
                                      className="gallery-action-btn delete-btn"
                                      title="删除"
                                    >
                                      <Trash2 size={16} />
                                    </button>
                                    <a
                                      href={part.image.url || (part.image.data ? `data:${part.image.mimeType};base64,${part.image.data}` : '#')}
                                      download={part.image.url ? undefined : `generated-${Date.now()}.png`}
                                      target={part.image.url ? '_blank' : undefined}
                                      rel="noopener noreferrer"
                                      className="gallery-action-btn"
                                      title="下载/查看原图"
                                    >
                                      ↓
                                    </a>
                                </div>
                              </div>
                          )}
                        </div>
                        <div className="gallery-card-footer">
                           <div className="gallery-card-title">生成结果</div>
                        </div>
                       </>
                     );
                   })()}
                </div>
              )}
            </div>
            
            {/* 下方：消息列表 (包含图片和文本，默认折叠) */}
            {results.length > 0 && (
              <div className="generated-results-list">
                {results.map((resultItem, resultIndex) => {
                  const isExpanded = expandedTexts.has(resultIndex);
                  const textParts = resultItem.data?.parts.filter(p => p.type === 'text') || [];
                  const imageParts = resultItem.data?.parts.filter(p => p.type === 'image') || [];
                  
                  return (
                    <div key={resultIndex} className="generated-result-item">
                      <div 
                        className="result-item-header" 
                        onClick={() => {
                          setExpandedTexts((prev) => {
                            const newSet = new Set(prev);
                            if (isExpanded) {
                              newSet.delete(resultIndex);
                            } else {
                              newSet.add(resultIndex);
                            }
                            return newSet;
                          });
                        }}
                        style={{ cursor: 'pointer', userSelect: 'none' }}
                      >
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                          <span 
                            className="toggle-icon" 
                            style={{ 
                              display: 'inline-block',
                              transform: isExpanded ? 'rotate(90deg)' : 'none', 
                              transition: 'transform 0.2s',
                              fontSize: '0.75rem',
                              color: '#666'
                            }}
                          >
                            ▶
                          </span>
                          <h3>第 {resultIndex + 1} 张 - 生成详情</h3>
                        </div>
                        <div className="result-item-status">
                          {resultItem.status === 'pending' && <span className="status-pending">等待中</span>}
                          {resultItem.status === 'generating' && <span className="status-generating">生成中...</span>}
                          {resultItem.status === 'success' && <span className="status-success">成功</span>}
                          {resultItem.status === 'error' && <span className="status-error">失败</span>}
                        </div>
                      </div>
                      
                      {isExpanded && (
                        <div className="generated-content" style={{ marginTop: '1rem', animation: 'fadeIn 0.3s ease' }}>
                          {resultItem.status === 'error' && resultItem.error && (
                            <div className="result-error-message">{resultItem.error}</div>
                          )}
                          
                          {/* 列表中的图片部分 - 恢复显示 */}
                          {imageParts.map((part, imageIndex) => (
                            part.image && (
                              <div key={`image-${imageIndex}`} className="result-part result-part-image">
                                <div className="generated-image-item">
                                  <img
                                    src={part.image.url || (part.image.data ? `data:${part.image.mimeType};base64,${part.image.data}` : '')}
                                    alt={`生成的图片 ${resultIndex + 1}-${imageIndex + 1}`}
                                    className="generated-image"
                                  />
                                  <a
                                    href={part.image.url || (part.image.data ? `data:${part.image.mimeType};base64,${part.image.data}` : '#')}
                                    download={part.image.url ? undefined : `generated-${Date.now()}-${resultIndex}-${imageIndex}.png`}
                                    target={part.image.url ? '_blank' : undefined}
                                    rel={part.image.url ? 'noopener noreferrer' : undefined}
                                    className="download-btn"
                                  >
                                    {part.image.url ? '查看原图' : '下载图片'}
                                  </a>
                                </div>
                              </div>
                            )
                          ))}

                          {/* 展示对话历史：优先显示 message_list，如果没有则显示 textParts */}
                          {(() => {
                            // 从第一张图片的 path 查找对应的图片信息
                            const firstImagePath = imageParts[0]?.image?.path;
                            const imageInfo = firstImagePath ? workspaceImages.find(img => img.path === firstImagePath) : null;
                            
                            // 如果有保存的 message_list，优先显示它（避免重复）
                            if (imageInfo?.message_list && imageInfo.message_list.length > 0) {
                              return (
                                <div className="text-messages-content">
                                  {imageInfo.message_list.map((msg, msgIndex) => (
                                    <div key={msgIndex} className="text-message" style={{ 
                                      borderLeftColor: msg.role === 'user' ? '#667eea' : '#10b981',
                                      backgroundColor: msg.role === 'user' ? '#f8f9fa' : '#f0fdf4'
                                    }}>
                                      {msg.role === 'user' && (
                                        <div style={{ 
                                          fontSize: '0.75rem', 
                                          color: '#666', 
                                          marginBottom: '0.5rem',
                                          fontWeight: 600,
                                          textTransform: 'uppercase'
                                        }}>
                                          👤 用户
                                        </div>
                                      )}
                                      <div style={{ lineHeight: 1.6, color: '#333' }}>
                                        {msg.content}
                                      </div>
                                    </div>
                                  ))}
                                </div>
                              );
                            }
                            
                            // 如果没有 message_list，显示 textParts（兼容旧数据）
                            if (textParts.length > 0) {
                              return (
                                <div className="text-messages-content">
                                  {textParts.map((part, textIndex) => (
                                    <div key={textIndex} className="text-message">
                                      {part.text}
                                    </div>
                                  ))}
                                </div>
                              );
                            }
                            
                            return null;
                          })()}
                          
                          {/* 如果成功且没有任何内容 */}
                          {resultItem.status === 'success' && textParts.length === 0 && imageParts.length === 0 && (
                              <div style={{ fontSize: '0.875rem', color: '#999', fontStyle: 'italic' }}>
                                  无详细内容。
                              </div>
                          )}
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
            
            {/* 兼容旧的单个结果展示 */}
            {results.length === 0 && result && (
              <div className="generated-result-item">
                  {(() => {
                    const isExpanded = expandedTexts.has(-1);
                    const textParts = result.parts.filter(p => p.type === 'text');
                    const imageParts = result.parts.filter(p => p.type === 'image');

                    return (
                      <>
                        <div 
                          className="result-item-header"
                          onClick={() => {
                            setExpandedTexts((prev) => {
                              const newSet = new Set(prev);
                              if (isExpanded) newSet.delete(-1);
                              else newSet.add(-1);
                              return newSet;
                            });
                          }}
                          style={{ cursor: 'pointer', userSelect: 'none', borderBottom: isExpanded ? '2px solid #e0e0e0' : 'none', paddingBottom: isExpanded ? '1rem' : '0', marginBottom: isExpanded ? '1rem' : '0' }}
                        >
                           <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                              <span 
                                className="toggle-icon" 
                                style={{ 
                                  display: 'inline-block',
                                  transform: isExpanded ? 'rotate(90deg)' : 'none', 
                                  transition: 'transform 0.2s',
                                  fontSize: '0.75rem',
                                  color: '#666'
                                }}
                              >
                                ▶
                              </span>
                              <h3>生成详情</h3>
                           </div>
                        </div>

                        {isExpanded && (
                          <div className="generated-content" style={{ animation: 'fadeIn 0.3s ease' }}>
                             {/* 图片部分 */}
                             {imageParts.map((part, imageIndex) => (
                                part.image && (
                                  <div key={`image-${imageIndex}`} className="result-part result-part-image">
                                    <div className="generated-image-item">
                                      <img
                                        src={part.image.url || (part.image.data ? `data:${part.image.mimeType};base64,${part.image.data}` : '')}
                                        alt={`生成的图片 ${imageIndex + 1}`}
                                        className="generated-image"
                                      />
                                      <a
                                        href={part.image.url || (part.image.data ? `data:${part.image.mimeType};base64,${part.image.data}` : '#')}
                                        download={part.image.url ? undefined : `generated-${Date.now()}-${imageIndex}.png`}
                                        target={part.image.url ? '_blank' : undefined}
                                        rel={part.image.url ? 'noopener noreferrer' : undefined}
                                        className="download-btn"
                                      >
                                        {part.image.url ? '查看原图' : '下载图片'}
                                      </a>
                                    </div>
                                  </div>
                                )
                              ))}

                              {/* 文本部分 */}
                              {textParts.length > 0 && (
                                <div className="text-messages-content">
                                  {textParts.map((part, i) => <div key={i} className="text-message">{part.text}</div>)}
                                </div>
                              )}
                              
                              {textParts.length === 0 && imageParts.length === 0 && (
                                <div style={{color:'#999', fontStyle: 'italic'}}>无详细内容。</div>
                              )}
                          </div>
                        )}
                      </>
                    );
                  })()}
              </div>
            )}
            
            {/* 回溯时的对话历史展示 */}
            {restoredImageInfo && restoredImageInfo.message_list && restoredImageInfo.message_list.length > 0 && (
              <div className="generated-results-list">
                <div className="generated-result-item">
                  <div className="result-item-header" style={{ cursor: 'default' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <h3>回溯的对话历史</h3>
                    </div>
                    <div className="result-item-status">
                      <span className="status-success">已回溯</span>
                    </div>
                  </div>
                  <div className="generated-content" style={{ marginTop: '1rem', animation: 'fadeIn 0.3s ease' }}>
                    {/* 展示最终生成的图片 */}
                    <div className="result-part result-part-image">
                      <div className="generated-image-item">
                        <img
                          src={restoredImageInfo.url}
                          alt={restoredImageInfo.name}
                          className="generated-image"
                        />
                        <a
                          href={restoredImageInfo.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="download-btn"
                        >
                          查看原图
                        </a>
                      </div>
                    </div>
                    
                    {/* 展示对话历史 */}
                    <div className="text-messages-content">
                      {restoredImageInfo.message_list.map((msg, msgIndex) => (
                        <div key={msgIndex} className="text-message" style={{ 
                          borderLeftColor: msg.role === 'user' ? '#667eea' : '#10b981',
                          backgroundColor: msg.role === 'user' ? '#f8f9fa' : '#f0fdf4'
                        }}>
                          {msg.role === 'user' && (
                            <div style={{ 
                              fontSize: '0.75rem', 
                              color: '#666', 
                              marginBottom: '0.5rem',
                              fontWeight: 600,
                              textTransform: 'uppercase'
                            }}>
                              👤 用户
                            </div>
                          )}
                          <div style={{ lineHeight: 1.6, color: '#333' }}>
                            {msg.content}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
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
      {previewImage && previewElementRef.current && (() => {
        // 获取缩略图元素的位置
        const thumbnailRect = previewElementRef.current.getBoundingClientRect();
        const offsetX = 20;
        const previewMaxWidth = 800; // 预览图片最大宽度
        const viewportWidth = window.innerWidth;
        const viewportHeight = window.innerHeight;
        const padding = 20; // 视口边距
        
        // 预估最大高度，用于计算位置。实际高度由 CSS max-height 限制
        const estimatedMaxHeight = Math.min(800, viewportHeight - padding * 2); 
        
        // 计算可用空间
        const spaceRight = viewportWidth - thumbnailRect.right - padding;
        const spaceLeft = thumbnailRect.left - padding;
        
        let left: number;
        let top: number;
        
        // 优先尝试显示在右侧
        if (spaceRight >= previewMaxWidth) {
          left = thumbnailRect.right + offsetX;
        } 
        // 如果右侧空间不足，尝试显示在左侧
        else if (spaceLeft >= previewMaxWidth) {
          left = thumbnailRect.left - previewMaxWidth - offsetX;
        }
        // 如果左右都不够，选择空间更大的一侧
        else {
          if (spaceRight >= spaceLeft) {
            left = thumbnailRect.right + offsetX;
          } else {
            left = thumbnailRect.left - previewMaxWidth - offsetX;
          }
        }
        
        // 垂直位置计算
        // 策略：默认顶部与缩略图对齐
        top = thumbnailRect.top;
        
        // 检查底部是否溢出
        // 使用估计高度来检查
        if (top + estimatedMaxHeight > viewportHeight - padding) {
          // 如果溢出，尝试向上移动，使底部对齐视口底部（留出 padding）
          top = viewportHeight - estimatedMaxHeight - padding;
        }
        
        // 再次检查顶部是否溢出（防止上移过多）
        if (top < padding) {
          top = padding;
        }
        
        // 最终边界检查，确保完全在视口内
        // 左边界检查
        if (left < padding) {
          left = padding;
        }
        // 右边界检查（注意：这里假设宽度是 previewMaxWidth，实际可能更小，但作为边界检查是安全的）
        if (left + previewMaxWidth > viewportWidth - padding) {
          // 如果右侧溢出，且是显示在右侧的情况，可能需要调整
          // 但通常由上面的 spaceRight 判断处理了
          // 这里主要处理 left 计算后的溢出
           if (left > viewportWidth - padding - 300) { // 如果实在太靠右
               left = viewportWidth - previewMaxWidth - padding;
           }
        }

        return (
          <div 
            className="image-preview-overlay" 
            key={previewImage.path}
            style={{
              left: `${left}px`,
              top: `${top}px`,
              // 限制最大高度，确保不超出视口
              maxHeight: `${viewportHeight - padding * 2}px`,
              display: 'flex',
              flexDirection: 'column'
            }}
          >
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
        );
      })()}

    </div>
  );
}
