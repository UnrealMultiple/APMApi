package legacy

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/UnrealMultiple/APMApi/pkg/constants"
	"github.com/UnrealMultiple/APMApi/pkg/db"
	"github.com/UnrealMultiple/APMApi/pkg/manifest"
)

type LegacyService struct {
	ctx context.Context
}

func NewLegacyService(ctx context.Context) *LegacyService {
	return &LegacyService{ctx: ctx}
}

// VersionZipPath 单插件某版本的zip路径
func VersionZipPath(assemblyName, version string) string {
	return filepath.Join(constants.PackedPluginsDir, assemblyName, version+".zip")
}

// Upload 处理插件包上传: 解压、打包单插件zip、刷新清单缓存、同步数据库
func (s *LegacyService) Upload(content []byte) error {
	manifest.SetSyncing(true)
	defer manifest.SetSyncing(false)

	zipReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return err
	}

	// 重建解压目录(打包目录保留, 历史版本文件不删除)
	if err := os.RemoveAll(constants.UploadedPluginsDir); err != nil {
		return err
	}
	if err := os.MkdirAll(constants.UploadedPluginsDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(constants.PackedPluginsDir, 0o755); err != nil {
		return err
	}

	if err := extractZip(zipReader, constants.UploadedPluginsDir); err != nil {
		return err
	}

	if err := os.WriteFile(constants.PluginsZipFile, content, 0o644); err != nil {
		return err
	}

	data, err := os.ReadFile(constants.PluginsJSONFile)
	if err != nil {
		return err
	}
	if err := manifest.Set(data); err != nil {
		return err
	}

	for _, m := range manifest.Parsed() {
		if err := packPlugin(m.AssemblyName, m.Version); err != nil {
			return err
		}
	}

	return syncDB(manifest.Parsed())
}

// extractZip 解压zip到指定目录(带zip slip防护)
func extractZip(zr *zip.Reader, destDir string) error {
	for _, f := range zr.File {
		target := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("非法的zip路径: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := writeZipFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

func writeZipFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

// packPlugin 将单个插件的所有文件打包成 packed_plugins/程序集名/版本号.zip
func packPlugin(assemblyName, version string) error {
	files, err := filepath.Glob(filepath.Join(constants.UploadedPluginsDir, "Plugins", assemblyName+".*"))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(constants.PackedPluginsDir, assemblyName), 0o755); err != nil {
		return err
	}

	out, err := os.Create(VersionZipPath(assemblyName, version))
	if err != nil {
		return err
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		w, err := zw.Create(filepath.Base(file))
		if err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
	}
	return zw.Close()
}

// pickDescription 优先取zh-CN描述, 其次en-US, 否则任取一个
func pickDescription(desc map[string]string) string {
	if d, ok := desc["zh-CN"]; ok {
		return d
	}
	if d, ok := desc["en-US"]; ok {
		return d
	}
	for _, d := range desc {
		return d
	}
	return ""
}

// syncDB 将清单同步到数据库: 更新插件与版本, 下架清单中不存在的插件
func syncDB(ms []manifest.Manifest) error {
	assemblyNames := make([]string, 0, len(ms))

	for _, m := range ms {
		assemblyNames = append(assemblyNames, m.AssemblyName)

		var p db.Plugin
		err := db.DB.Unscoped().Where("assembly_name = ?", m.AssemblyName).First(&p).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			p = db.Plugin{
				Name:         m.Name,
				AssemblyName: m.AssemblyName,
				Description:  pickDescription(m.Description),
				Version:      m.Version,
			}
			if err := db.DB.Create(&p).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			// 重新上传的插件恢复上架
			if err := db.DB.Unscoped().Model(&p).Updates(map[string]any{
				"name":        m.Name,
				"description": pickDescription(m.Description),
				"version":     m.Version,
				"deleted_at":  nil,
			}).Error; err != nil {
				return err
			}
		}

		if err := db.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "plugin_id"}, {Name: "version"}},
			DoUpdates: clause.AssignmentColumns([]string{"file_path", "updated_at"}),
		}).Create(&db.PluginVersion{
			PluginID: p.ID,
			Version:  m.Version,
			FilePath: VersionZipPath(m.AssemblyName, m.Version),
		}).Error; err != nil {
			return err
		}
	}

	// 下架清单中不存在的插件
	q := db.DB
	if len(assemblyNames) > 0 {
		q = q.Where("assembly_name NOT IN ?", assemblyNames)
	} else {
		q = q.Session(&gorm.Session{AllowGlobalUpdate: true})
	}
	if err := q.Delete(&db.Plugin{}).Error; err != nil {
		return err
	}

	hlog.Infof("插件清单同步完成, 共%d个插件", len(ms))
	return nil
}

// ResolveLegacyZip 老API: 按程序集名称解析最新版本zip路径
func (s *LegacyService) ResolveLegacyZip(assemblyName string) (string, bool) {
	for _, m := range manifest.Parsed() {
		if m.AssemblyName == assemblyName {
			path := VersionZipPath(m.AssemblyName, m.Version)
			if _, err := os.Stat(path); err == nil {
				return path, true
			}
			return "", false
		}
	}
	return "", false
}

// IncDownloadCount 按程序集名称增加下载量(老API下载单插件时统计)
func (s *LegacyService) IncDownloadCount(assemblyName string) {
	if db.DB == nil {
		return
	}
	if err := db.DB.Model(&db.Plugin{}).Where("assembly_name = ?", assemblyName).
		UpdateColumn("download_count", gorm.Expr("download_count + 1")).Error; err != nil {
		hlog.Errorf("更新下载量失败: %v", err)
	}
}
