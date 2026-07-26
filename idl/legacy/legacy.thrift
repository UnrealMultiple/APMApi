namespace go legacy

// 老API插件清单(保留大驼峰)
struct PluginManifest {
    // 插件名称
    1: required string Name;
    // 插件版本
    2: required string Version;
    // 插件作者
    3: required string Author;
    // 多语言描述
    4: required map<string, string> Description;
    // 程序集名称
    5: required string AssemblyName;
    // 插件路径
    6: required string Path;
    // 依赖列表
    7: required list<string> Dependencies;
    // 是否支持热重载
    8: required bool HotReload;
}

// 镜像信息请求
struct SupermarketXmlReq {
}

// 镜像信息响应
struct SupermarketXmlResp {
    // 致谢
    1: required string Thanks (go.tag = 'json:"超(市)级感谢"');
    // 备案号
    2: required string ICP (go.tag = 'json:"备案号"');
}

// 上传插件包请求
struct UploadPluginReq {
    // 上传密钥
    1: required string token (api.form = 'token');
}

// 上传插件包响应
struct UploadPluginResp {
    // 返回消息
    1: required string message;
}

// 获取插件清单列表请求
struct GetPluginListReq {
}

// 获取插件清单列表响应(实际返回清单数组)
struct GetPluginListResp {
    // 插件清单列表
    1: required list<PluginManifest> plugins;
}

// 获取单个插件清单请求
struct GetPluginManifestReq {
    // 程序集名称
    1: required string AssemblyName (api.query = 'assembly_name');
}

// 获取单个插件清单响应(实际返回清单对象)
struct GetPluginManifestResp {
    // 插件清单
    1: required PluginManifest plugin;
}

// 下载所有插件请求
struct GetAllPluginsReq {
}

// 下载所有插件响应(实际返回zip文件流)
struct GetAllPluginsResp {
}

// 下载单个插件请求
struct GetPluginZipReq {
    // 程序集名称
    1: required string AssemblyName (api.query = 'assembly_name');
}

// 下载单个插件响应(实际返回zip文件流)
struct GetPluginZipResp {
}

service LegacyService {
    // 镜像信息
    SupermarketXmlResp SupermarketXml(1: SupermarketXmlReq req) (api.get = '/supermarket/xml');
    // 上传插件包
    UploadPluginResp UploadPlugin(1: UploadPluginReq req) (api.post = '/plugin/upload');
    // 获取插件清单列表
    GetPluginListResp GetPluginList(1: GetPluginListReq req) (api.get = '/plugin/get_plugin_list');
    // 获取单个插件清单
    GetPluginManifestResp GetPluginManifest(1: GetPluginManifestReq req) (api.get = '/plugin/get_plugin_manifest/');
    // 下载所有插件
    GetAllPluginsResp GetAllPlugins(1: GetAllPluginsReq req) (api.get = '/plugin/get_all_plugins');
    // 下载单个插件
    GetPluginZipResp GetPluginZip(1: GetPluginZipReq req) (api.get = '/plugin/get_plugin_zip');
}
