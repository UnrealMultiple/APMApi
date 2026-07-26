// Package manifest 维护老API的插件清单缓存(Plugins.json 原样返回)与同步状态
package manifest

import (
	"encoding/json"
	"os"
	"sync"
	"sync/atomic"

	"github.com/UnrealMultiple/APMApi/pkg/constants"
)

// Manifest 插件清单条目(老API模型, 大驼峰)
type Manifest struct {
	Name         string            `json:"Name"`
	Version      string            `json:"Version"`
	Author       string            `json:"Author"`
	Description  map[string]string `json:"Description"`
	AssemblyName string            `json:"AssemblyName"`
	Path         string            `json:"Path"`
	Dependencies []string          `json:"Dependencies"`
	HotReload    bool              `json:"HotReload"`
}

var (
	mu      sync.RWMutex
	raw     []byte
	items   []json.RawMessage
	parsed  []Manifest
	syncing atomic.Bool
)

// Load 从磁盘加载插件清单
func Load() error {
	data, err := os.ReadFile(constants.PluginsJSONFile)
	if err != nil {
		return err
	}
	return Set(data)
}

// Set 更新清单缓存
func Set(data []byte) error {
	var rawItems []json.RawMessage
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return err
	}
	var ps []Manifest
	if err := json.Unmarshal(data, &ps); err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()
	raw = data
	items = rawItems
	parsed = ps
	return nil
}

// Raw 返回清单原始JSON(与上传的 Plugins.json 完全一致)
func Raw() ([]byte, bool) {
	mu.RLock()
	defer mu.RUnlock()
	return raw, raw != nil
}

// Parsed 返回解析后的清单列表
func Parsed() []Manifest {
	mu.RLock()
	defer mu.RUnlock()
	return parsed
}

// Find 按程序集名称查找清单条目, 返回原始JSON对象
func Find(assemblyName string) (json.RawMessage, bool) {
	mu.RLock()
	defer mu.RUnlock()
	for i, p := range parsed {
		if p.AssemblyName == assemblyName {
			return items[i], true
		}
	}
	return nil, false
}

// Syncing 是否正在同步插件
func Syncing() bool {
	return syncing.Load()
}

// SetSyncing 设置同步状态
func SetSyncing(b bool) {
	syncing.Store(b)
}
