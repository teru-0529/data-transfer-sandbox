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
	LegacyDB DbConfig
	SourceDB DbConfig
	DistDB   DbConfig
}

type DbConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Db       string
}

// FUNCTION:
func LeadEnv() *Config {
	// PROCESS: envファイルのロード
	_, err := os.Stat(".env")
	if !os.IsNotExist(err) {
		godotenv.Load()
		log.Print("loaded environment variables from .env file.")
	}

	// PROCESS: legacyDB
	var legacyDB DbConfig
	if err = envconfig.Process("LEGACY_MARIADB", &legacyDB); err != nil {
		log.Fatal(err)
	}

	// PROCESS: sourceDB
	var sourceDB DbConfig
	if err = envconfig.Process("SOURCE_POSTGRES", &sourceDB); err != nil {
		log.Fatal(err)
	}

	// PROCESS: distDB
	var distDB DbConfig
	if err = envconfig.Process("DIST_POSTGRES", &distDB); err != nil {
		log.Fatal(err)
	}

	return &Config{LegacyDB: legacyDB, SourceDB: sourceDB, DistDB: distDB}
}
