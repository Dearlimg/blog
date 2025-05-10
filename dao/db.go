package dao

import (
	"blog/dao/mysql"
)

type database struct {
	DB mysql.DB
}

var Database = new(database)
