package main

import "github.com/fatih/color"

func doMigrate(arg2, arg3 string) error {
	color.Yellow("Running doMigrate")

	// run the migration command
	switch arg2 {
	case "up":
		err := cel.RunPopMigrations()
		if err != nil {
			exitGracefully(err)
			return err
		}

	case "down":
		if arg3 == "all" {
			err := cel.PopMigrateDown(-1)
			if err != nil {
				return err
			}
		} else {
			err := cel.PopMigrateDown(1)
			if err != nil {
				return err
			}
		}

	case "reset":
		err := cel.PopMigrateReset()
		if err != nil {
			return err
		}
	default:
		showHelp()
	}

	return nil
}
