package application

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vasyahuyasa/librebread/api"
	"github.com/vasyahuyasa/librebread/infrastructure"
	"github.com/vasyahuyasa/librebread/migrations"
	"github.com/vasyahuyasa/librebread/web"
)

type Application struct {
	DatabasePath string
}

func (app *Application) Run() error {
	db, err := app.openSqlite()
	if err != nil {
		return err
	}
	defer db.Close()

	err = migrations.Migrate(db.DB)
	if err != nil {
		return err
	}

	sms := infrastructure.NewSQLSMSRepo(db)

	l := api.NewLibrebread(sms)
	h := api.Handler(l)

	srv := web.NewServer(h)

	err = srv.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (app *Application) openSqlite() (*sqlx.DB, error) {
	return sqlx.Open("sqlite3", app.DatabasePath)
}
