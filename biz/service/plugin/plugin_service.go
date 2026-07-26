package plugin

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/UnrealMultiple/APMApi/biz/model/plugin"
	"github.com/UnrealMultiple/APMApi/biz/pack"
	"github.com/UnrealMultiple/APMApi/pkg/constants"
	"github.com/UnrealMultiple/APMApi/pkg/db"
	"github.com/UnrealMultiple/APMApi/pkg/errno"
	"github.com/UnrealMultiple/APMApi/pkg/manifest"
)

type PluginService struct {
	ctx context.Context
}

func NewPluginService(ctx context.Context) *PluginService {
	return &PluginService{ctx: ctx}
}

func normalizePage(page, pageSize *int64) (int64, int64) {
	p, size := int64(constants.DefaultPage), int64(constants.DefaultPageSize)
	if page != nil && *page > 0 {
		p = *page
	}
	if pageSize != nil && *pageSize > 0 {
		size = *pageSize
	}
	if size > constants.MaxPageSize {
		size = constants.MaxPageSize
	}
	return p, size
}

// escapeLike 转义 LIKE 模式中的特殊字符
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

func (s *PluginService) query(keyword string, page, pageSize *int64) (*plugin.PluginListData, error) {
	p, size := normalizePage(page, pageSize)

	q := db.DB.Model(&db.Plugin{})
	if keyword != "" {
		q = q.Where("name ILIKE ?", "%"+escapeLike(keyword)+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []db.Plugin
	if err := q.Order("id").Offset(int((p - 1) * size)).Limit(int(size)).Find(&rows).Error; err != nil {
		return nil, err
	}

	return &plugin.PluginListData{
		Total:    total,
		Page:     p,
		PageSize: size,
		Items:    pack.PluginInfoList(rows),
	}, nil
}

// Search 搜索插件, 支持按名称模糊查询与分页
func (s *PluginService) Search(req *plugin.SearchReq) (*plugin.PluginListData, error) {
	return s.query(req.GetKeyword(), req.Page, req.PageSize)
}

// List 获取插件列表, 支持分页
func (s *PluginService) List(req *plugin.ListReq) (*plugin.PluginListData, error) {
	return s.query("", req.Page, req.PageSize)
}

// Detail 获取插件详情(含历史版本)
func (s *PluginService) Detail(req *plugin.DetailReq) (*plugin.PluginDetail, error) {
	var p db.Plugin
	if err := db.DB.First(&p, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.PluginNotExistErr
		}
		return nil, err
	}

	var versions []db.PluginVersion
	if err := db.DB.Where("plugin_id = ?", p.ID).Order("id DESC").Find(&versions).Error; err != nil {
		return nil, err
	}

	return &plugin.PluginDetail{
		Plugin:   pack.PluginInfo(&p),
		Versions: pack.PluginVersionInfoList(versions),
	}, nil
}

// Rate 评分插件(0-10分, 用IP区分用户, 重复评分覆盖)
func (s *PluginService) Rate(req *plugin.RateReq, ip string) (*plugin.RateData, error) {
	if req.Score < constants.MinScore || req.Score > constants.MaxScore {
		return nil, errno.ScoreInvalidErr
	}

	var p db.Plugin
	if err := db.DB.First(&p, req.PluginID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.PluginNotExistErr
		}
		return nil, err
	}

	data := &plugin.RateData{}
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "plugin_id"}, {Name: "ip"}},
			DoUpdates: clause.AssignmentColumns([]string{"score", "updated_at"}),
		}).Create(&db.PluginRating{PluginID: p.ID, IP: ip, Score: req.Score}).Error; err != nil {
			return err
		}

		var stat struct {
			Count int64
			Avg   float64
		}
		if err := tx.Model(&db.PluginRating{}).
			Select("COUNT(*) AS count, COALESCE(AVG(score), 0) AS avg").
			Where("plugin_id = ?", p.ID).
			Scan(&stat).Error; err != nil {
			return err
		}
		stat.Avg = math.Round(stat.Avg*100) / 100

		if err := tx.Model(&db.Plugin{}).Where("id = ?", p.ID).
			UpdateColumns(map[string]any{
				"rating_count": stat.Count,
				"rating_score": stat.Avg,
			}).Error; err != nil {
			return err
		}

		data.RatingCount = stat.Count
		data.RatingScore = stat.Avg
		return nil
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Download 解析插件下载文件(默认最新版本, 可指定版本), 并增加下载量
func (s *PluginService) Download(req *plugin.DownloadReq) (filePath, fileName string, err error) {
	if manifest.Syncing() {
		return "", "", errno.PluginSyncingErr
	}

	var p db.Plugin
	if err := db.DB.First(&p, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errno.PluginNotExistErr
		}
		return "", "", err
	}

	version := p.Version
	if v := req.GetVersion(); v != "" {
		version = v
	}

	var pv db.PluginVersion
	if err := db.DB.Where("plugin_id = ? AND version = ?", p.ID, version).First(&pv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errno.VersionNotExistErr
		}
		return "", "", err
	}

	if _, err := os.Stat(pv.FilePath); err != nil {
		return "", "", errno.VersionNotExistErr
	}

	if err := db.DB.Model(&db.Plugin{}).Where("id = ?", p.ID).
		UpdateColumn("download_count", gorm.Expr("download_count + 1")).Error; err != nil {
		return "", "", err
	}

	return pv.FilePath, fmt.Sprintf("%s-%s.zip", p.AssemblyName, version), nil
}

// DownloadAll 解析全量插件包文件(不增加下载量)
func (s *PluginService) DownloadAll() (string, error) {
	if manifest.Syncing() {
		return "", errno.PluginSyncingErr
	}
	if _, err := os.Stat(constants.PluginsZipFile); err != nil {
		return "", errno.PluginPackNotUploadedErr
	}
	return constants.PluginsZipFile, nil
}
