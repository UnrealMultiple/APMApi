namespace go plugin

// 插件信息
struct PluginInfo {
    // 插件ID
    1: required i64 id;
    // 插件名称
    2: required string name;
    // 插件描述
    3: required string description;
    // 最新版本号
    4: required string version;
    // 下载链接
    5: required string download_url;
    // 下载次数
    6: required i64 download_count;
    // 评分人数
    7: required i64 rating_count;
    // 平均评分(0-10)
    8: required double rating_score;
    // 创建时间
    9: required string created_at;
    // 更新时间
    10: required string updated_at;
    // 程序集名称
    11: required string assembly_name;
}

// 插件版本信息
struct PluginVersionInfo {
    // 版本号
    1: required string version;
    // 发布时间
    2: required string created_at;
}

// 插件详情
struct PluginDetail {
    // 插件信息
    1: required PluginInfo plugin;
    // 历史版本列表
    2: required list<PluginVersionInfo> versions;
}

// 插件分页数据
struct PluginListData {
    // 总数
    1: required i64 total;
    // 页码
    2: required i64 page;
    // 每页数量
    3: required i64 page_size;
    // 插件列表
    4: required list<PluginInfo> items;
}

// 搜索插件请求
struct SearchReq {
    // 按插件名称模糊查询
    1: optional string keyword (api.query = 'keyword');
    // 页码, 默认1
    2: optional i64 page (api.query = 'page');
    // 每页数量, 默认20
    3: optional i64 page_size (api.query = 'page_size');
}

// 搜索插件响应
struct SearchResp {
    // 状态码
    1: required i32 code;
    // 返回消息
    2: required string msg;
    // 分页数据
    3: optional PluginListData data;
}

// 插件列表请求
struct ListReq {
    // 页码, 默认1
    1: optional i64 page (api.query = 'page');
    // 每页数量, 默认20
    2: optional i64 page_size (api.query = 'page_size');
}

// 插件列表响应
struct ListResp {
    // 状态码
    1: required i32 code;
    // 返回消息
    2: required string msg;
    // 分页数据
    3: optional PluginListData data;
}

// 插件详情请求
struct DetailReq {
    // 插件ID
    1: required i64 id (api.path = 'id');
}

// 插件详情响应
struct DetailResp {
    // 状态码
    1: required i32 code;
    // 返回消息
    2: required string msg;
    // 插件详情
    3: optional PluginDetail data;
}

// 评分插件请求
struct RateReq {
    // 插件ID
    1: required i64 plugin_id (api.body = 'plugin_id');
    // 评分(0-10)
    2: required i32 score (api.body = 'score', api.vd = "$ >= 0 && $ <= 10");
}

// 评分数据
struct RateData {
    // 评分人数
    1: required i64 rating_count;
    // 平均评分
    2: required double rating_score;
}

// 评分插件响应
struct RateResp {
    // 状态码
    1: required i32 code;
    // 返回消息
    2: required string msg;
    // 评分数据
    3: optional RateData data;
}

// 下载插件请求
struct DownloadReq {
    // 插件ID
    1: required i64 id (api.path = 'id');
    // 版本号, 默认最新版本
    2: optional string version (api.query = 'version');
}

// 下载插件响应(实际返回zip文件流)
struct DownloadResp {
    // 状态码
    1: required i32 code;
    // 返回消息
    2: required string msg;
}

// 下载所有插件请求
struct DownloadAllReq {
}

// 下载所有插件响应(实际返回zip文件流)
struct DownloadAllResp {
    // 状态码
    1: required i32 code;
    // 返回消息
    2: required string msg;
}

service PluginService {
    // 搜索插件
    SearchResp Search(1: SearchReq req) (api.get = '/api/v1/plugin/search');
    // 获取插件列表
    ListResp List(1: ListReq req) (api.get = '/api/v1/plugin/list');
    // 获取插件详情
    DetailResp Detail(1: DetailReq req) (api.get = '/api/v1/plugin/:id/detail');
    // 评分插件
    RateResp Rate(1: RateReq req) (api.post = '/api/v1/plugin/rate');
    // 下载插件
    DownloadResp Download(1: DownloadReq req) (api.get = '/api/v1/plugin/:id/download');
    // 下载所有插件
    DownloadAllResp DownloadAll(1: DownloadAllReq req) (api.get = '/api/v1/plugin/download_all');
}
