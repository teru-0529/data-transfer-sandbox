/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package infra

// TITLE:環境変数の読込み

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// STRUCT:
type Config struct {
	Base      BaseConfig `envconfig:""`
	LegacyDB  DbConfig   `envconfig:"LEGACY_MARIADB"`
	WorkDB    DbConfig   `envconfig:"WORK_POSTGRES"`
	ProductDB DbConfig   `envconfig:"PRODUCT_POSTGRES"`
}

// 基本設定
type BaseConfig struct {
	LegacyDataKey string `envconfig:"LEGACY_DATA_KEY" required:"true"`
	AppVersion    string `envconfig:"APP_VERSION" default:"v0.0.1"`
	ToolVersion   string
}

// データベース接続設定
type DbConfig struct {
	User     string `envconfig:"USER" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
	Host     string `envconfig:"HOST" default:"localhost"`
	Port     int    `envconfig:"PORT" required:"true"`
	Database string `envconfig:"DB" required:"true"`
}

// FUNCTION:
func LeadConfig(version string) (Config, DbConnection, func()) {
	// PROCESS: envファイルのロード
	_, err := os.Stat(".env")
	if !os.IsNotExist(err) {
		godotenv.Load()
		log.Print("loaded environment variables from .env file.")
	}

	// PROCESS: オブジェクトに変換
	var config Config
	if err = envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	config.Base.ToolVersion = version

	// PROCESS: データベース(Sqlboiler)コネクションの取得
	conns, cleanUp := initDB(&config)

	return config, conns, cleanUp
}

// FUNCTION: 出力先のディレクトリ名標準
func (config Config) DirName() string {
	return fmt.Sprintf("%s(%s)", config.Base.ToolVersion, config.Base.LegacyDataKey)
}

// FUNCTION: ユニックスタイムからの秒数に変換し、フォーマット
func ElapsedStr(now time.Time) string {
	var tZero = time.Unix(0, 0).UTC()
	return tZero.Add(time.Duration(time.Since(now))).Format("15:04:05.000")
}
