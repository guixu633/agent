-- 添加 is_current 字段到 workspaces 表
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS is_current BOOLEAN NOT NULL DEFAULT FALSE;

-- 创建索引以便快速查找当前工作区
CREATE INDEX IF NOT EXISTS idx_workspaces_is_current ON workspaces(is_current) WHERE is_current = TRUE;

-- 如果表中已有数据，将第一个工作区设置为当前工作区（可选）
-- UPDATE workspaces SET is_current = TRUE WHERE id = (SELECT id FROM workspaces ORDER BY created_at ASC LIMIT 1);

