package db

import (
	"context"
	"fmt"
	"time"

	"database/sql"

	"github.com/prometheus/alertmanager/template"
)

// MySQLDB A database that does nothing
type MySQLDB struct {
	db *sql.DB
}

// ConnectMySQL connect to a MySQL database using the provided data source name
func ConnectMySQL(dsn string) (*MySQLDB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("Empty DSN provided, can't connect to database")
	}
	database, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL database: %s", err)
	}

	return &MySQLDB{
		db: database,
	}, nil
}

// Save implements Storer interface
func (d MySQLDB) Save(data *template.Data) error {
	return d.unitOfWork(func() error {
		r, err := d.db.Exec(`
			INSERT INTO AlertGroup (timestamp, receiver, status, externalURL)
			VALUES (now(), ?, ?, ?)`, data.Receiver, data.Status, data.ExternalURL)
		if err != nil {
			return fmt.Errorf("failed to insert into AlertGroups: %s", err)
		}

		_, err = r.LastInsertId() // alertGroupID
		if err != nil {
			return fmt.Errorf("failed to get AlertGroups inserted id: %s", err)
		}

		return nil
	})
}

func (d MySQLDB) unitOfWork(f func() error) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %s", err)
	}

	err = f()

	if err != nil {
		e := tx.Rollback()
		if e != nil {
			return fmt.Errorf("failed to rollback transaction (%s) after failing execution: %s", e, err)
		}
		return fmt.Errorf("failed execution: %s", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %s", err)
	}
	return nil
}

// Ping implements Storer interface
func (d MySQLDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return d.db.PingContext(ctx)
}

func (MySQLDB) String() string {
	return "mysql database driver"
}