/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package infra

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// TITLE:環境変数の読込み

// STRUCT:
type Config struct {
	SourceDB PostgresConfig
	DistDB   PostgresConfig
}

type PostgresConfig struct {
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

	// PROCESS: sourceDB
	var sourceDB PostgresConfig
	if err = envconfig.Process("SOURCE_POSTGRES", &sourceDB); err != nil {
		log.Fatal(err)
	}

	// PROCESS: distDB
	var distDB PostgresConfig
	if err = envconfig.Process("DIST_POSTGRES", &distDB); err != nil {
		log.Fatal(err)
	}

	return &Config{SourceDB: sourceDB, DistDB: distDB}
}
