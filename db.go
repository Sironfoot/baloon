package baloon

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// DB represents a series of database scripts to run against a database given its Connection.
// It uses database/sql behind the scenes so your database driver will need to support it.
type DB struct {
	Connection DBConn
	Script     Script
	Scripts    []Script
}

// Run will run the database setup
func (dbSetup DB) Run(appRoot string) error {
	db, err := sql.Open(dbSetup.Connection.Driver, dbSetup.Connection.String)
	if err != nil {
		return fmt.Errorf("Error connecting to database: %s", err.Error())
	}
	defer db.Close()

	var scripts []Script

	if dbSetup.Script.Type == ScriptTypeLiteral || dbSetup.Script.Type == ScriptTypePath {
		scripts = append(scripts, dbSetup.Script)
	}
	scripts = append(scripts, dbSetup.Scripts...)

	for _, script := range scripts {
		if script.Type == ScriptTypeLiteral {
			_, err = db.Exec(script.Command)
			if err != nil {
				return fmt.Errorf("Error running script \"%s\": %s", truncate(script.Command, 40, "..."), err.Error())
			}
		} else if script.Type == ScriptTypePath {
			globPath := filepath.Join(appRoot, script.Command)
			files, err := filepath.Glob(globPath)
			if err != nil {
				return fmt.Errorf("Error getting files from path \"%s\": %s", script.Command, err.Error())
			}

			for _, file := range files {
				data, err := ioutil.ReadFile(file)
				_, err = db.Exec(string(data))
				if err != nil {
					return fmt.Errorf("Error executing script \"%s\": %s", file, err.Error())
				}
			}
		}
	}

	return nil
}
