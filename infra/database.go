/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package infra

// TITLE:DB設定

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// STRUCT:
type DbConnection struct {
	SourceDB *sql.DB
	DistDB   *sql.DB
}

// FUNCTION: DB setting
func InitDB() (DbConnection, func()) {

	// PROCESS: envファイルのロード
	config := LeadEnv()

	// PROCESS: Connection作成
	cons := DbConnection{
		SourceDB: createConnection(config.SourceDB),
		DistDB:   createConnection(config.DistDB),
	}
	return cons, func() {
		cons.SourceDB.Close()
		cons.DistDB.Close()
	}
}

// FUNCTION: conection
func createConnection(config PostgresConfig) *sql.DB {

	dns := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Db,
	)

	// PROCESS:database open
	con, err := sql.Open("postgres", dns)
	if err != nil {
		log.Fatal(err)
	}

	// PROCESS:connection pool settings
	con.SetMaxIdleConns(10)
	con.SetMaxOpenConns(10)
	con.SetConnMaxLifetime(300 * time.Second)

	// PROCESS:connection test
	if err = con.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Printf("db connection prepared [%s]\n", dns)
	return con
}
