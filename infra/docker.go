/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package infra

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"
)

// TITLE:Dockerアクセス

// STRUCT:
type DbContainer struct {
	Name   string
	Config DbConfig
}

// STRUCT:
const TEMP_PATH string = "/tmp/dump.sql"

var TEMP_GZ_PATH string = fmt.Sprintf("%s.gz", TEMP_PATH)

// FUNCTION: コンテナ作成
func NewContainer(name string, config DbConfig) DbContainer {
	return DbContainer{Name: name, Config: config}
}

// FUNCTION: ダンプファイルをLoadする
func (ct DbContainer) LoadDb(loadfile string) error {
	s := time.Now()

	// PROCESS: ファイルが存在しない場合エラー
	loadfilePath := path.Join("dist", path.Join("cleansing", loadfile))
	if f, err := os.Stat(loadfilePath); os.IsNotExist(err) || f.IsDir() {
		return fmt.Errorf("not exist dumpfile[%s]: %s", loadfile, err.Error())
	}

	// PROCESS: コンテナ内にコピー
	// docker cp {dumpfile.sql.gz} {work-db}:/tmp/dumpfile.sql.gz
	copyArgs := []string{"cp", loadfilePath, fmt.Sprintf("%s:%s", ct.Name, TEMP_GZ_PATH)}
	if err := dockerExec(copyArgs); err != nil {
		return fmt.Errorf("failed to copy load file: %v", err)
	}

	// PROCESS: cleanDBにデータロード
	// docker exec -e PGPASSWORD={password} -i {work-db} bash -c "gzip -d -c /tmp/dumpfile.sql.gz | psql -U {postgres} -d {workDB}"
	command := fmt.Sprintf("gzip -d -c %s | psql -U %s -d %s", TEMP_GZ_PATH, ct.Config.User, ct.Config.Database)
	loadArgs := []string{
		"exec",
		"-e", fmt.Sprintf("PGPASSWORD=%s", ct.Config.Password),
		"-i", ct.Name,
		"bash", "-c", command,
	}
	if err := dockerExec(loadArgs); err != nil {
		return fmt.Errorf("failed to db load: %v", err)
	}

	duration := time.Since(s).Seconds()
	log.Printf("load completed [%s] … %3.2fs\n", filepath.Base(loadfile), duration)
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
