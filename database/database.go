package database

import "database/sql"

const (
	dbConnectionStr = "root:@tcp(localhost:3306)/majestic_million"
	dbMaxIdleConns  = 4
	dbMaxConns      = 100
	totalWorker     = 100
	csvFile         = "majestic_million.csv"
)

func OpenDBConnection() (*sql.DB, error) {

	db, err := sql.Open("mysql", dbConnectionStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(dbMaxConns)
	db.SetMaxIdleConns(dbMaxIdleConns)

	return db, nil
}
