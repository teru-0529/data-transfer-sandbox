/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package infra

// TITLE:環境変数の読込み

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// STRUCT:
const TEMP_PATH string = "/tmp/dump.sql"

var TEMP_GZ_PATH string = fmt.Sprintf("%s.gz", TEMP_PATH)

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
	User          string `envconfig:"USER" required:"true"`
	Password      string `envconfig:"PASSWORD" required:"true"`
	Host          string `envconfig:"HOST" default:"localhost"`
	Port          int    `envconfig:"PORT" required:"true"`
	Database      string `envconfig:"DB" required:"true"`
	ContainerName string `envconfig:"CONTAINER" required:"true"`
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

// FUNCTION: ファイルをコンテナ上のDBにLoadする
func (config DbConfig) Load(loadfilePath string) error {
	s := time.Now()

	// PROCESS: コンテナ内にコピー
	// docker cp {dumpfile.sql.gz} {work-db}:{/tmp/dumpfile.sql.gz}
	copyArgs := []string{"cp", loadfilePath, fmt.Sprintf("%s:%s", config.ContainerName, TEMP_GZ_PATH)}
	if err := dockerExec(copyArgs); err != nil {
		return fmt.Errorf("failed to copy load file: %v", err)
	}

	// PROCESS: cleanDBにデータロード
	// docker exec -e PGPASSWORD={password} -i {work-db} bash -c gzip -d -c {/tmp/dumpfile.sql.gz} | psql -U {postgres} -d {workDB}
	command := fmt.Sprintf("gzip -d -c %s | psql -U %s -d %s", TEMP_GZ_PATH, config.User, config.Database)
	loadArgs := []string{
		"exec",
		"-e", fmt.Sprintf("PGPASSWORD=%s", config.Password),
		"-i", config.ContainerName,
		"bash", "-c", command,
	}
	if err := dockerExec(loadArgs); err != nil {
		return fmt.Errorf("failed to db load: %v", err)
	}

	duration := time.Since(s).Seconds()
	log.Printf("load completed [%s] … %3.2fs\n", filepath.Base(loadfilePath), duration)
	return nil
}

// FUNCTION: DBデータをダンプする
func (config DbConfig) Dump(dumpfilePath string, extArgs []string) error {
	s := time.Now()

	// PROCESS: フォルダが存在しない場合作成する
	dir := filepath.Dir(dumpfilePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return fmt.Errorf("cannot create directory: %s", err.Error())
		}
	}

	// PROCESS: cleanDBをダンプ
	// docker exec -e PGPASSWORD={password} -i {work-db} bash -c pg_dump -U {postgres} -d {workDB} {--data-only --schema=clean} > {/tmp/dump.sql} && gzip {/tmp/dump.sql}
	command := fmt.Sprintf("pg_dump -U %s -d %s %s > %s && gzip -f %s", config.User, config.Database, strings.Join(extArgs, " "), TEMP_PATH, TEMP_PATH)
	dumpArgs := []string{
		"exec",
		"-e", fmt.Sprintf("PGPASSWORD=%s", config.Password),
		"-i", config.ContainerName,
		"bash", "-c", command,
	}
	if err := dockerExec(dumpArgs); err != nil {
		return fmt.Errorf("failed to db dump: %v", err)
	}

	// PROCESS: ローカルにコピー
	// docker cp {work-db}:{/tmp/dumpfile.sql.gz} {dumpfile.sql.gz}
	copyArgs := []string{"cp", fmt.Sprintf("%s:%s", config.ContainerName, TEMP_GZ_PATH), dumpfilePath}
	if err := dockerExec(copyArgs); err != nil {
		return fmt.Errorf("failed to copy dump file: %v", err)
	}

	duration := time.Since(s).Seconds()
	log.Printf("pg_dump completed [%s] … %3.2fs\n", filepath.Base(dumpfilePath), duration)
	return nil
}

// FUNCTION: Dockerコマンド実行
func dockerExec(args []string) error {
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// PROCESS: コマンドを実行

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// FUNCTION: データ変換出力先のディレクトリ:`dist/toolversion/appVersion(legacyDataKey)`
func (config Config) TransferDir() string {
	dirName := fmt.Sprintf("%s(%s)", config.Base.AppVersion, config.Base.LegacyDataKey)
	return path.Join("dist", path.Join(config.Base.ToolVersion, dirName))
}

// FUNCTION: クレンジング出力先のディレクトリ:`work/toolversion/legacyDataKey`
func (config Config) CleansingDir() string {
	return path.Join("work", path.Join(config.Base.ToolVersion, config.Base.LegacyDataKey))
}

// FUNCTION: ユニックスタイムからの秒数に変換し、フォーマット
func ElapsedStr(now time.Time) string {
	var tZero = time.Unix(0, 0).UTC()
	return tZero.Add(time.Duration(time.Since(now))).Format("15:04:05.000")
}
