package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RunMigrations 执行数据库迁移
func RunMigrations() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 迁移文件列表（按顺序执行）
	migrationFiles := []string{
		"001_create_tables.sql",
		"002_add_is_current_to_workspaces.sql",
		"003_add_image_generation_info.sql",
	}

	// 尝试多个可能的路径前缀
	possiblePrefixes := []string{
		filepath.Join("internal", "database", "migrations"),
		filepath.Join("backend", "internal", "database", "migrations"),
		"./internal/database/migrations",
	}

	for _, migrationFile := range migrationFiles {
		var data []byte
		var err error
		var found bool

		for _, prefix := range possiblePrefixes {
			migrationPath := filepath.Join(prefix, migrationFile)
			data, err = os.ReadFile(migrationPath)
			if err == nil {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("读取迁移文件失败: %s", migrationFile)
		}

		// 分割 SQL 语句（按分号分割，但需要处理触发器中的分号）
		sqlStatements := splitSQLStatements(string(data))

		// 执行每个 SQL 语句
		for _, stmt := range sqlStatements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			_, err := DB.Exec(stmt)
			if err != nil {
				return fmt.Errorf("执行迁移失败 (文件: %s): %w\nSQL: %s", migrationFile, err, stmt)
			}
		}
	}

	return nil
}

// splitSQLStatements 分割 SQL 语句
// 简单实现：按分号分割，但保留在 $$ 块内的分号
func splitSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inDollarQuote := false
	var dollarTag string

	for i := 0; i < len(sql); i++ {
		char := sql[i]

		// 检测 $$ 开始
		if i < len(sql)-1 && sql[i] == '$' && sql[i+1] == '$' {
			// 提取 $tag$ 格式
			if !inDollarQuote {
				inDollarQuote = true
				dollarTag = "$$"
				i++ // 跳过第二个 $
				current.WriteString("$$")
				continue
			} else if dollarTag == "$$" {
				inDollarQuote = false
				dollarTag = ""
				current.WriteString("$$")
				i++ // 跳过第二个 $
				continue
			}
		}

		current.WriteByte(char)

		// 如果不在 $$ 块内，遇到分号就分割
		if !inDollarQuote && char == ';' {
			stmt := current.String()
			if strings.TrimSpace(stmt) != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	// 添加最后一个语句
	if current.Len() > 0 {
		stmt := current.String()
		if strings.TrimSpace(stmt) != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}
