package celeritas

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/pop"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (c *Celeritas) popConnect() (*pop.Connection, error) {
	tx, err := pop.Connect("development")
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreatePopMigration creates up/down pop migrations
func (c *Celeritas) CreatePopMigration(up, down []byte, migrationName, migrationType string) error {
	var migrationPath = c.RootPath + "/migrations"
	err := pop.MigrationCreate(migrationPath, migrationName, migrationType, up, down)
	if err != nil {
		return err
	}
	return nil
}

func (c *Celeritas) RunPopMigrations(tx *pop.Connection) error {
	var migrationPath = c.RootPath + "/migrations"

	fm, err := pop.NewFileMigrator(migrationPath, tx)
	if err != nil {
		return err
	}

	// run the migrations
	err = fm.Up()
	if err != nil {
		return err
	}
	return nil
}

func (c *Celeritas) PopMigrateDown(tx *pop.Connection, steps ...int) error {
	var migrationPath = c.RootPath + "/migrations"

	step := 1
	if len(steps) > 0 {
		step = steps[0]
	}
	fm, err := pop.NewFileMigrator(migrationPath, tx)
	if err != nil {
		return err
	}

	// run the migrations
	err = fm.Down(step)
	if err != nil {
		return err
	}
	return nil
}

func (c *Celeritas) PopMigrateReset(tx *pop.Connection) error {
	var migrationPath = c.RootPath + "/migrations"

	fm, err := pop.NewFileMigrator(migrationPath, tx)
	if err != nil {
		return err
	}

	// run the migrations
	err = fm.Reset()
	if err != nil {
		return err
	}
	return nil
}
