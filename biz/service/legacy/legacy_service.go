package legacy

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
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

	ms := manifest.Parsed()
	byAssembly := make(map[string]manifest.Manifest, len(ms))
	for _, m := range ms {
		byAssembly[m.AssemblyName] = m
	}
	for _, m := range ms {
		if err := packPlugin(m, byAssembly); err != nil {
			return err
		}
	}

	return syncDB(ms)
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

// dependencyClosure 返回插件自身及其依赖树上的所有程序集名(BFS, 防循环依赖)
func dependencyClosure(root manifest.Manifest, byAssembly map[string]manifest.Manifest) []string {
	visited := make(map[string]bool)
	names := make([]string, 0, 1+len(root.Dependencies))
	queue := []string{root.AssemblyName}

	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if visited[name] {
			continue
		}
		visited[name] = true
		names = append(names, name)

		// 依赖若也是清单中的插件, 继续展开其依赖; 否则是纯库文件, 无下级依赖
		if m, ok := byAssembly[name]; ok {
			queue = append(queue, m.Dependencies...)
		}
	}
	return names
}

// packPlugin 将插件及其整个依赖树的文件打包成 packed_plugins/程序集名/版本号.zip
func packPlugin(m manifest.Manifest, byAssembly map[string]manifest.Manifest) error {
	var files []string
	for _, name := range dependencyClosure(m, byAssembly) {
		fs, err := filepath.Glob(filepath.Join(constants.UploadedPluginsDir, "Plugins", name+".*"))
		if err != nil {
			return err
		}
		files = append(files, fs...)
	}

	if err := os.MkdirAll(filepath.Join(constants.PackedPluginsDir, m.AssemblyName), 0o755); err != nil {
		return err
	}

	out, err := os.Create(VersionZipPath(m.AssemblyName, m.Version))
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

		// 序列化 Descriptions 和 Dependencies 为 JSON
		descriptionsJSON, _ := json.Marshal(m.Description)
		dependenciesJSON, _ := json.Marshal(m.Dependencies)

		var p db.Plugin
		err := db.DB.Unscoped().Where("assembly_name = ?", m.AssemblyName).First(&p).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			p = db.Plugin{
				Name:         m.Name,
				AssemblyName: m.AssemblyName,
				Description:  pickDescription(m.Description),
				Descriptions: string(descriptionsJSON),
				Author:       m.Author,
				HotReload:    m.HotReload,
				Path:         m.Path,
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
				"name":         m.Name,
				"description":  pickDescription(m.Description),
				"descriptions": string(descriptionsJSON),
				"author":       m.Author,
				"hot_reload":   m.HotReload,
				"path":         m.Path,
				"version":      m.Version,
				"deleted_at":   nil,
			}).Error; err != nil {
				return err
			}
		}

		if err := db.DB.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "plugin_id"}, {Name: "version"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"file_path", "dependencies", "hot_reload", "path", "updated_at",
			}),
		}).Create(&db.PluginVersion{
			PluginID:     p.ID,
			Version:      m.Version,
			FilePath:     VersionZipPath(m.AssemblyName, m.Version),
			Dependencies: string(dependenciesJSON),
			HotReload:    m.HotReload,
			Path:         m.Path,
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

// ResolveLegacyZip 老API: 按程序集名称从数据库解析最新版本zip路径
func (s *LegacyService) ResolveLegacyZip(assemblyName string) (string, bool) {
	var p db.Plugin
	if err := db.DB.Where("assembly_name = ?", assemblyName).First(&p).Error; err != nil {
		return "", false
	}
	var pv db.PluginVersion
	if err := db.DB.Where("plugin_id = ? AND version = ?", p.ID, p.Version).First(&pv).Error; err != nil {
		return "", false
	}
	if _, err := os.Stat(pv.FilePath); err != nil {
		return "", false
	}
	return pv.FilePath, true
}

// dbPluginToManifest 将数据库插件+版本转换为老API清单格式
func dbPluginToManifest(p db.Plugin, pv db.PluginVersion) manifest.Manifest {
	var descriptions map[string]string
	if p.Descriptions != "" {
		_ = json.Unmarshal([]byte(p.Descriptions), &descriptions)
	}
	if descriptions == nil {
		descriptions = map[string]string{}
	}

	var deps []string
	if pv.Dependencies != "" {
		_ = json.Unmarshal([]byte(pv.Dependencies), &deps)
	}
	if deps == nil {
		deps = []string{}
	}

	return manifest.Manifest{
		Name:         p.Name,
		Version:      p.Version,
		Author:       p.Author,
		Description:  descriptions,
		AssemblyName: p.AssemblyName,
		Path:         pv.Path,
		Dependencies: deps,
		HotReload:    pv.HotReload,
	}
}

// GetAllPluginManifests 从数据库读取所有插件清单(供老API GetPluginList使用)
func (s *LegacyService) GetAllPluginManifests() ([]manifest.Manifest, error) {
	var plugins []db.Plugin
	if err := db.DB.Find(&plugins).Error; err != nil {
		return nil, err
	}
	if len(plugins) == 0 {
		return []manifest.Manifest{}, nil
	}

	// 批量取所有最新版本记录
	var allVersions []db.PluginVersion
	if err := db.DB.Find(&allVersions).Error; err != nil {
		return nil, err
	}
	// key: "pluginID:version"
	pvMap := make(map[string]db.PluginVersion, len(allVersions))
	for _, pv := range allVersions {
		pvMap[fmt.Sprintf("%d:%s", pv.PluginID, pv.Version)] = pv
	}

	result := make([]manifest.Manifest, 0, len(plugins))
	for _, p := range plugins {
		pv := pvMap[fmt.Sprintf("%d:%s", p.ID, p.Version)]
		result = append(result, dbPluginToManifest(p, pv))
	}
	return result, nil
}

// GetPluginManifestByAssembly 从数据库读取单个插件清单(供老API GetPluginManifest使用)
func (s *LegacyService) GetPluginManifestByAssembly(assemblyName string) (manifest.Manifest, bool) {
	var p db.Plugin
	if err := db.DB.Where("assembly_name = ?", assemblyName).First(&p).Error; err != nil {
		return manifest.Manifest{}, false
	}
	var pv db.PluginVersion
	if err := db.DB.Where("plugin_id = ? AND version = ?", p.ID, p.Version).First(&pv).Error; err != nil {
		return manifest.Manifest{}, false
	}
	return dbPluginToManifest(p, pv), true
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
