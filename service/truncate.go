/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package service

import (
	"context"
	"fmt"

	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
)

// TITLE: トランケート

// STRUCT: テーブル名
type TableNameOnly struct {
	Tablename string `boil:"tablename"`
}

// FUNCTION: cleanDBのテーブルを全てtruncate
func TruncateCleanDbAll(conns infra.DbConnection) {
	boil.DebugMode = true

	var ctx context.Context = context.Background()
	var tables []TableNameOnly
	queries.Raw("SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = 'clean'").Bind(ctx, conns.WorkDB, &tables)
	for _, table := range tables {
		queries.Raw(fmt.Sprintf("truncate clean.%s CASCADE;", table.Tablename)).ExecContext(ctx, conns.WorkDB)
	}

	boil.DebugMode = false
}
