package celeritas

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gobuffalo/pop"
	"log"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/go-sql-driver/mysql"
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
	tx, err := c.popConnect()
	if err != nil {
		return err
	}
	defer tx.Close()

	var migrationPath = c.RootPath + "/migrations"
	err = pop.MigrationCreate(migrationPath, migrationName, migrationType, up, down)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (c *Celeritas) RunPopMigrations() error {
	var migrationPath = c.RootPath + "/migrations"
	tx, err := c.popConnect()
	if err != nil {
		color.Red("Error: %v\n", err)
		return err
	}
	defer tx.Close()

	fm, err := pop.NewFileMigrator(migrationPath, tx)
	if err != nil {
		color.Red("Error: %v\n", err)
		return err
	}

	// run the migrations
	err = fm.Up()
	if err != nil {
		color.Red("Error: %v\n", err)
		return err
	}
	return nil
}

func (c *Celeritas) PopMigrateDown(steps ...int) error {
	var migrationPath = c.RootPath + "/migrations"
	tx, err := c.popConnect()
	if err != nil {
		return err
	}
	defer tx.Close()

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

func (c *Celeritas) PopMigrateReset() error {
	var migrationPath = c.RootPath + "/migrations"
	tx, err := c.popConnect()
	if err != nil {
		return err
	}
	defer tx.Close()

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

func (c *Celeritas) PopSteps(steps int) error {
	return c.PopMigrateDown(steps)
}

func (c *Celeritas) MigrateUp(dsn string) error {
	m, err := migrate.New("file://"+c.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		log.Println("Error running migration:", err)
		return err
	}
	return nil
}

func (c *Celeritas) MigrateDownAll(dsn string) error {
	m, err := migrate.New("file://"+c.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Down(); err != nil {
		return err
	}

	return nil
}

func (c *Celeritas) Steps(n int, dsn string) error {
	m, err := migrate.New("file://"+c.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(n); err != nil {
		return err
	}

	return nil
}

func (c *Celeritas) MigrateForce(dsn string) error {
	m, err := migrate.New("file://"+c.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Force(-1); err != nil {
		return err
	}

	return nil
}
