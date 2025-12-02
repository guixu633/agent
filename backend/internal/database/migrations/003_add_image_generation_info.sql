-- 添加图片生成信息字段
ALTER TABLE images ADD COLUMN IF NOT EXISTS source_type VARCHAR(50) NOT NULL DEFAULT 'upload';
ALTER TABLE images ADD COLUMN IF NOT EXISTS prompt TEXT DEFAULT '';
ALTER TABLE images ADD COLUMN IF NOT EXISTS ref_images JSONB DEFAULT '[]'::jsonb;
ALTER TABLE images ADD COLUMN IF NOT EXISTS message_list JSONB DEFAULT '[]'::jsonb;

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_images_source_type ON images(source_type);

