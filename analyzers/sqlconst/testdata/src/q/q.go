package q

import (
	"context"
	"database/sql"
	"fmt"
)

const byID = "SELECT title FROM notes WHERE id = ?"

func Good(ctx context.Context, db *sql.DB, id string) (*sql.Rows, error) {
	return db.QueryContext(ctx, byID, id)
}

func Literal(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM notes WHERE id = ?", id)
	return err
}

func Concat(ctx context.Context, db *sql.DB, title string) (*sql.Rows, error) {
	return db.QueryContext(ctx, "SELECT id FROM notes WHERE title = '"+title+"'") // want `S2: this SQL query is built at run time`
}

func Format(tx *sql.Tx, id string) *sql.Row {
	return tx.QueryRow(fmt.Sprintf("SELECT title FROM notes WHERE id = '%s'", id)) // want `S2: this SQL query is built at run time`
}

func Prepared(ctx context.Context, c *sql.Conn, table string) (*sql.Stmt, error) {
	return c.PrepareContext(ctx, "SELECT * FROM "+table) // want `S2: this SQL query is built at run time`
}
