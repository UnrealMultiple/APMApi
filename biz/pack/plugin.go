package pack

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/UnrealMultiple/APMApi/biz/model/plugin"
	"github.com/UnrealMultiple/APMApi/pkg/db"
)

// unmarshalStringSlice 解析JSON字符串数组, 失败时返回空切片
func unmarshalStringSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	var v []string
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return []string{}
	}
	return v
}

// unmarshalStringMap 解析JSON字符串字典, 失败时返回空map
func unmarshalStringMap(s string) map[string]string {
	if s == "" {
		return map[string]string{}
	}
	var v map[string]string
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return map[string]string{}
	}
	return v
}

// DownloadURL 拼接插件下载链接
func DownloadURL(pluginID int64) string {
	return fmt.Sprintf("/api/v1/plugin/%d/download", pluginID)
}

// PluginInfo 数据库插件转API模型
func PluginInfo(p *db.Plugin) *plugin.PluginInfo {
	return &plugin.PluginInfo{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Version:       p.Version,
		DownloadURL:   DownloadURL(p.ID),
		DownloadCount: p.DownloadCount,
		RatingCount:   p.RatingCount,
		RatingScore:   p.RatingScore,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
		AssemblyName:  p.AssemblyName,
		Author:        p.Author,
		Descriptions:  unmarshalStringMap(p.Descriptions),
		HotReload:     p.HotReload,
		Path:          p.Path,
	}
}

// PluginInfoList 数据库插件列表转API模型列表
func PluginInfoList(ps []db.Plugin) []*plugin.PluginInfo {
	items := make([]*plugin.PluginInfo, 0, len(ps))
	for i := range ps {
		items = append(items, PluginInfo(&ps[i]))
	}
	return items
}

// PluginVersionInfoList 数据库插件版本列表转API模型列表
func PluginVersionInfoList(vs []db.PluginVersion) []*plugin.PluginVersionInfo {
	items := make([]*plugin.PluginVersionInfo, 0, len(vs))
	for _, v := range vs {
		items = append(items, &plugin.PluginVersionInfo{
			Version:      v.Version,
			CreatedAt:    v.CreatedAt.Format(time.RFC3339),
			Dependencies: unmarshalStringSlice(v.Dependencies),
			HotReload:    v.HotReload,
			Path:         v.Path,
		})
	}
	return items
}
