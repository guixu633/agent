-- 创建 workspaces 表
CREATE TABLE IF NOT EXISTS workspaces (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_workspaces_name ON workspaces(name);
CREATE INDEX IF NOT EXISTS idx_workspaces_created_at ON workspaces(created_at);

-- 创建 images 表
CREATE TABLE IF NOT EXISTS images (
    id BIGSERIAL PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    oss_path VARCHAR(512) NOT NULL,
    oss_url VARCHAR(512) NOT NULL,
    thumbnail_path VARCHAR(512) NOT NULL DEFAULT '',
    thumbnail_url VARCHAR(512) NOT NULL DEFAULT '',
    size BIGINT NOT NULL DEFAULT 0,
    mime_type VARCHAR(100) NOT NULL DEFAULT 'image/jpeg',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- 确保同一工作区内的图片名称唯一
    UNIQUE(workspace_id, name)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_images_workspace_id ON images(workspace_id);
CREATE INDEX IF NOT EXISTS idx_images_name ON images(name);
CREATE INDEX IF NOT EXISTS idx_images_created_at ON images(created_at);
CREATE INDEX IF NOT EXISTS idx_images_oss_path ON images(oss_path);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为 workspaces 表创建更新时间触发器
DROP TRIGGER IF EXISTS update_workspaces_updated_at ON workspaces;
CREATE TRIGGER update_workspaces_updated_at
    BEFORE UPDATE ON workspaces
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 为 images 表创建更新时间触发器
DROP TRIGGER IF EXISTS update_images_updated_at ON images;
CREATE TRIGGER update_images_updated_at
    BEFORE UPDATE ON images
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

