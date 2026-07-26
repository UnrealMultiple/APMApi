# APMApi

TShock 插件市场 API, 基于 [CloudWeGo Hertz](https://github.com/cloudwego/hertz) + Thrift IDL + GORM + PostgreSQL。

## 快速开始

```bash
# 1. 准备配置
cp config/config.yaml.example config/config.yaml   # 修改 host/port、上传key、pg 连接

# 2. 运行
go run .

# 3. 构建 Linux 产物
make build-linux
```

## 代码生成

IDL 位于 `idl/`, 使用 [hz](https://www.cloudwego.io/docs/hertz/tutorials/toolkit/) + thrift 脚手架:

```bash
make hz-gen-api
```

## API

### 新API (`/api/v1`, snake_case, 响应 code/msg 平铺)

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v1/plugin/search?keyword=&page=&page_size=` | 搜索插件(名称模糊查询, 分页) |
| GET | `/api/v1/plugin/list?page=&page_size=` | 插件列表(分页) |
| GET | `/api/v1/plugin/:id/detail` | 插件详情(含历史版本) |
| POST | `/api/v1/plugin/rate` | 评分插件(0-10分, 按IP区分用户) |
| GET | `/api/v1/plugin/:id/download?version=` | 下载插件(默认最新版, 可指定版本, 下载量+1) |
| GET | `/api/v1/plugin/download_all` | 下载所有插件(不增加下载量) |

### 老API (保留, 功能不变, 模型大驼峰)

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/supermarket/xml` | 镜像信息 |
| POST | `/plugin/upload` | 上传插件包(form: token + file) |
| GET | `/plugin/get_plugin_list` | 插件清单列表 |
| GET | `/plugin/get_plugin_manifest/?assembly_name=` | 单个插件清单 |
| GET | `/plugin/get_all_plugins` | 下载所有插件 |
| GET | `/plugin/get_plugin_zip?assembly_name=` | 下载单个插件 |

## 目录结构

```
idl/            Thrift IDL
biz/handler     HTTP 处理器
biz/service     业务逻辑
biz/pack        响应封装
biz/mw          中间件
config/         配置(viper + yaml)
pkg/db          GORM + PostgreSQL
pkg/manifest    老API插件清单缓存
```
