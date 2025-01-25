/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package infra

// TITLE:環境変数の読込み

import (
	"log"
	"os"

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
	LegacyFile string `envconfig:"LEGACY_LOAD_FILE" required:"true"`
	AppVersion string `envconfig:"APP_VERSION" default:"v0.0.1"`
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
func LeadConfig() (Config, DbConnection, func()) {
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

	// PROCESS: データベース(Sqlboiler)コネクションの取得
	conns, cleanUp := initDB(&config)

	return config, conns, cleanUp
}
