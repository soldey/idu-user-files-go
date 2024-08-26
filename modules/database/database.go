package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"os"
)

var Service *bun.DB

type Database struct {
	host     string
	port     string
	user     string
	password string
	dbName   string
}

func NewDatabase(host string, port string, user string, password string, dbName string) Database {
	database := Database{
		host: host, port: port, user: user, password: password, dbName: dbName,
	}
	database.Connect()
	return database
}

func (d *Database) Connect() error {
	if Service == nil {
		sqldb := sql.OpenDB(
			pgdriver.NewConnector(
				pgdriver.WithNetwork("tcp"),
				pgdriver.WithAddr(fmt.Sprintf("%s:%s", d.host, d.port)),
				pgdriver.WithUser(d.user),
				pgdriver.WithPassword(d.password),
				pgdriver.WithDatabase(d.dbName),
				pgdriver.WithApplicationName("testApp"),
				pgdriver.WithInsecure(true),
			),
		)
		Service = bun.NewDB(sqldb, pgdialect.New())
		Service.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
		if _, err := Service.ExecContext(context.Background(), "select 1"); err == nil {
			fmt.Println("ready")
		} else {
			return err
		}
	}
	return nil
}

func (d *Database) dispose() error {
	if Service != nil {
		err := Service.Close()
		if err != nil {
			fmt.Printf("error %s", err)
			return err
		}
		Service = nil
	}
	return nil
}

var DbConfig = NewDatabase(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_DATABASE"))
