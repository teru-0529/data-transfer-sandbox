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
	LegacyDB  DbConfig
	workDB    DbConfig
	productDB DbConfig
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

	// PROCESS: workDB
	var workDB DbConfig
	if err = envconfig.Process("WORK_POSTGRES", &workDB); err != nil {
		log.Fatal(err)
	}

	// PROCESS: productDB
	var productDB DbConfig
	if err = envconfig.Process("PRODUCT_POSTGRES", &productDB); err != nil {
		log.Fatal(err)
	}

	return &Config{LegacyDB: legacyDB, workDB: workDB, productDB: productDB}
}
