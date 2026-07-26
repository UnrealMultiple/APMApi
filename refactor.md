# APM API 重构

1. 使用 go hertz 框架，参考项目 <https://github.com/ACaiCat/tiktok-go>
2. 添加数据库，pg
3. 实现API /api/v1/plugin/search 搜索插件
   - 支持分页查询
   - 支持按插件名称模糊查询

4. 实现API /api/v1/plugin/{id}/detail 获取插件详情
   - 支持根据插件ID查询插件详情

5. 实现API /api/v1/plugin/rate 评分插件
   - 0- 10分
   - 用IP区分用户

6. 实现API /api/v1/plugin/download_all
   - 下载所有插件

7. 实现API /api/v1/plugin/{id}/download
   - 根据插件ID下载插件

8. 实现API /api/v1/plugin/list
    - 获取所有插件列表
    - 支持分页查询

9. 保留原来所有的接口，功能不变

10. 支持下载量统计，download_all 不增加下载量

11. 脚手架idl使用thrift
12. 不需要使用 redis
13. 老API的模型保留大驼峰，新API全用 snake_case
14. 写一下 ci，编译成Linux就行
15. 配置文件用yaml，要可以配host port，还要upload的key
16. 只用 gorm , 不用 gorm-gen
17. 字段id, name, description, version, download_url, download_count, rating_count, rating_score, created_at, updated_at, deleted_at
18. 下载接口默认最新版本，可以指定版本下载
19. 要实现版本管理，插件可以有多个版本
20. 原来的接口现在在http://api.terraria.ink:11434/plugin/get_plugin_list，一些定义你可以看一下