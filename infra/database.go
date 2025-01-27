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

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// STRUCT:
type DbConnection struct {
	LegacyDB  *sql.DB
	WorkDB    *sql.DB
	ProductDB *sql.DB
}

// FUNCTION: DB setting
func initDB(config *Config) (DbConnection, func()) {

	// PROCESS: Connection作成
	cons := DbConnection{
		LegacyDB:  createCon(genMysqlDns(config.LegacyDB)),
		WorkDB:    createCon(genPsqlDns(config.WorkDB)),
		ProductDB: createCon(genPsqlDns(config.ProductDB)),
	}
	return cons, func() {
		cons.LegacyDB.Close()
		cons.WorkDB.Close()
		cons.ProductDB.Close()
	}
}

// FUNCTION: connection
func createCon(dbtype string, dns string) *sql.DB {

	// PROCESS:database open
	con, err := sql.Open(dbtype, dns)
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

	log.Printf("db(%s) connection prepared [%s]\n", dbtype, dns)
	return con
}

// FUNCTION: psqlDNS
func genPsqlDns(config DbConfig) (string, string) {

	return "postgres", fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)
}

// FUNCTION: mysqlDNS
func genMysqlDns(config DbConfig) (string, string) {

	return "mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&loc=Asia%%2FTokyo",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)
}
