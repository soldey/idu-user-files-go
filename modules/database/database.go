package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

var Database *bun.DB

type DatabaseConfig struct {
	host     string
	port     string
	user     string
	password string
	dbName   string
}

func NewDatabase(host string, port string, user string, password string, dbName string) DatabaseConfig {
	database := DatabaseConfig{
		host: host, port: port, user: user, password: password, dbName: dbName,
	}
	database.Connect()
	return database
}

func (d *DatabaseConfig) Connect() error {
	if Database == nil {
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
		Database = bun.NewDB(sqldb, pgdialect.New())
		Database.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
		if _, err := Database.ExecContext(context.Background(), "select 1"); err == nil {
			fmt.Println("ready")
		} else {
			return err
		}
	}
	return nil
}

func (d *DatabaseConfig) dispose() error {
	if Database != nil {
		err := Database.Close()
		if err != nil {
			fmt.Printf("error %s", err)
			return err
		}
		Database = nil
	}
	return nil
}

var DbConfig DatabaseConfig

//var DbConfig = NewDatabase(
//	common.Config.Get("DB_HOST"),
//	common.Config.Get("DB_PORT"),
//	common.Config.Get("DB_USER"),
//	common.Config.Get("DB_PASSWORD"),
//	common.Config.Get("DB_DATABASE"),
//)
