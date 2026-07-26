package constants

const (
	// UploadedPluginsDir 插件包解压目录
	UploadedPluginsDir = "uploaded_plugins"
	// PackedPluginsDir 单插件打包目录(按 程序集名/版本号.zip 存放)
	PackedPluginsDir = "packed_plugins"
	// PluginsJSONFile 插件清单文件
	PluginsJSONFile = UploadedPluginsDir + "/Plugins.json"
	// PluginsZipFile 全量插件包文件
	PluginsZipFile = UploadedPluginsDir + "/Plugins.zip"

	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100

	MinScore = 0
	MaxScore = 10
)
