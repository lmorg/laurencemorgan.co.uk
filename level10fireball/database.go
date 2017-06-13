// database
package main

import (
	"database/sql"
	"fmt"
	//"github.com/ziutek/mymysql/godrv"
	_ "github.com/go-sql-driver/mysql"
)

const DATE_FMT_DATABASE string = "2006-01-02 15:04:05"

////////////////////////////////////////////////////////////////////////////////

func dbOpen() (failed bool, db *sql.DB) {
	var err error

	//db, err := sql.Open("mymysql", fmt.Sprintf("%s/%s/%s", DB_DATABASE, DB_USERNAME, DB_PASSWORD))

	if DB_CON_TYPE == "tcp" {

		// TCP/IP
		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4,utf8",
			DB_USERNAME, DB_PASSWORD, DB_HOSTNAME, DB_PORT_NUM, DB_DATABASE))
	} else {

		// UNIX socket
		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@unix(%s)/%s?charset=utf8mb4,utf8",
			DB_USERNAME, DB_PASSWORD, DB_UNIX_SOC, DB_DATABASE))
	}

	db.SetMaxIdleConns(1000)

	if err != nil {
		if err.Error() == "Error 1040: Too many connections" {
			failed = true
		} else {
			failOnErr(err, "dbOpen")
		}
	}

	return
}

func dbClose(db *sql.DB) {
	db.Close()
}

////////////////////////////////////////////////////////////////////////////////

func dbQueryRow(session *Session, str_sql string, args ...interface{}) (row *sql.Row) {
	row = session.db.QueryRow(str_sql, args...)
	return
}

func dbSelectRow(session *Session, surpress_error_msg bool, row *sql.Row, args ...interface{}) (err error) {
	if row != nil {
		err = row.Scan(args...)
		isErr(session, err, surpress_error_msg, "Database", "dbSelectRow")
		return
	}
	return raiseErr(session, "row *sql.Row == nil", surpress_error_msg, "Database", "dbSelectRow")
}

////////////////////////////////////////////////////////////////////////////////

func dbQueryRows(session *Session, surpress_error_msg bool, errp *error, str_sql string, args ...interface{}) (rows *sql.Rows) {
	rows, *errp = session.db.Query(str_sql, args...)
	isErr(session, *errp, surpress_error_msg, "Database", "dbQueryRows")
	return
}

func dbSelectRows(session *Session, surpress_error_msg bool, errp *error, rows *sql.Rows, args ...interface{}) (data_returned bool) {
	if rows != nil {
		data_returned = rows.Next()
		if data_returned {
			*errp = rows.Scan(args...)
			//errp = &err
			if *errp != nil {
				isErr(session, *errp, surpress_error_msg, "Database", "dbSelectRows")
				return false
			}
		}

		return
	}
	raiseErr(session, "rows *sql.Rows == nil", surpress_error_msg, "Database", "dbSelectRows")
	return false
}

////////////////////////////////////////////////////////////////////////////////

func dbInsertRow(session *Session, surpress_error_msg bool, auto_close_transaction bool, str_sql string, args ...interface{}) (transaction *sql.Tx, result sql.Result, err error) {

	// create a transaction
	transaction, err = session.db.Begin()
	if err != nil {
		isErr(session, err, surpress_error_msg, "Database", "dbInsertRow")
		return
	}

	result, err = dbInsertTransaction(session, transaction, surpress_error_msg, auto_close_transaction, str_sql, args...)
	return
}

func dbInsertTransaction(session *Session, transaction *sql.Tx, surpress_error_msg bool, auto_close_transaction bool, str_sql string, args ...interface{}) (result sql.Result, err error) {
	// create a statement
	statement, err := transaction.Prepare(str_sql)
	if err == nil {
		// run insert SQL
		result, err = statement.Exec(args...)
	}

	if err != nil {
		isErr(session, err, surpress_error_msg, "Database", "dbInsertTransation")
		return
	}

	// attempt to close the trasaction if auto close enabled
	if auto_close_transaction {
		err = transaction.Commit()
		isErr(session, err, surpress_error_msg, "Database", "dbInsertTransation")
	}
	return
}
