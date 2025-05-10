package tx

import (
	mysql "blog/dao/mysql/sqlc"
	"context"
	"database/sql"
)

type SqlStore struct {
	*mysql.Queries // 嵌入 *db.Queries，可以直接访问 db.Queries 中定义的方法和字段，不需要间接访问
	Pool           *sql.DB
}

func (s SqlStore) ExecContext(ctx context.Context, s2 string, i ...interface{}) (sql.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (s SqlStore) PrepareContext(ctx context.Context, s2 string) (*sql.Stmt, error) {
	//TODO implement me
	panic("implement me")
}

func (s SqlStore) QueryContext(ctx context.Context, s2 string, i ...interface{}) (*sql.Rows, error) {
	//TODO implement me
	panic("implement me")
}

func (s SqlStore) QueryRowContext(ctx context.Context, s2 string, i ...interface{}) *sql.Row {
	//TODO implement me
	panic("implement me")
}
