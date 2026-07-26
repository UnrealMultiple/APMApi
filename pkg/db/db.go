package db

import (
	"fmt"
	_ "time/tzdata"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/UnrealMultiple/APMApi/config"
)

var DB *gorm.DB

func Init() {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		config.AppConfig.Postgres.Host,
		config.AppConfig.Postgres.User,
		config.AppConfig.Postgres.Password,
		config.AppConfig.Postgres.DBName,
		config.AppConfig.Postgres.Port,
	)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		hlog.Fatal(err)
	}

	if err = DB.AutoMigrate(&Plugin{}, &PluginVersion{}, &PluginRating{}); err != nil {
		hlog.Fatal(err)
	}
}
