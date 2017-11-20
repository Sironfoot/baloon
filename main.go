// Package baloon is a setup and teardown test fixture library for end-to-end testing of HTTP APIs written in Go.
package baloon

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// These consts represent the 2 types of database scripts we can use, either literal or file path
const (
	// ScriptTypeLiteral specifies literal database command text
	ScriptTypeLiteral = 1

	// ScriptTypePath specifies a glob file pattern for
	// database commands stored in files
	ScriptTypePath = 2
)

// DBConn represents a database connection including the driver and connection string.
// Uses sql.DB, so make sure your database driver package supports it and is imported.
type DBConn struct {
	// Driver is your database driver name, passed as the first
	// argument to sql.Open
	Driver string

	// String is the database connection string, passed as the
	// second argument to sql.Open
	String string
}

// Script represents a database script run either as a setup or teardown routine.
// Command can either be a literal script, or a path (using globbing patterns) to a
// script file or files.
type Script struct {
	// Type is the Script type to use.
	Type int

	// Command is either a literal database command, or a file
	// glob pattern, depending on the 'Type'.
	Command string
}

// NewScript returns a Script that represents a literal database command to run.
func NewScript(command string) Script {
	return Script{
		Type:    ScriptTypeLiteral,
		Command: command,
	}
}

// NewScriptPath returns a Script that represents a glob path to a script files or files to run.
func NewScriptPath(path string) Script {
	return Script{
		Type:    ScriptTypePath,
		Command: path,
	}
}

// App represents settings and arguments for your Go HTTP API executable.
type App struct {
	// BuildArguments is a list of build arguments to include when baloon
	// tries to buld your App executable. They are run via "go build yourArgsHere..."
	BuildArguments []string

	// RunArguments is a list of command line arguments to
	// include when your Go executable is run.
	RunArguments []string

	// WaitForOutputLine specifies a line of text that baloon should wait
	// to appear in either stdout or stderr in order to signal that the App is
	// ready to start excepting HTTP requests.
	WaitForOutputLine string

	// WaitTimeout is how long baloon should wait for the
	// 'WaitForOutputLine' to appear.
	WaitTimeout time.Duration
}

// FixtureConfig is a configuration object for your test Fixture.
type FixtureConfig struct {
	// AppRoot is an absolute path to the root of your Go application directory,
	// where your main.go file is located.
	AppRoot string

	// DatabaseSetups is a list of one or more database setup commands to run
	// before the test suite is run.
	DatabaseSetups []DB

	// AppSetup specifies configuration settings for your Go app executable.
	AppSetup App

	// DatabaseTeardowns is a list of one or more database teardown
	// commands to run after the test suite has run.
	DatabaseTeardowns []DB
}

// UnitTest represents database commands and a func to run at the beginning or end of each unit test.
type UnitTest struct {
	// DatabaseRoutines is a list of one or more database setup commands to
	// run before each unit, or at the end of each unit test.
	DatabaseRoutines []DB

	// Func is a function to run before each unit test is run.
	Func func()
}

// NewFixture returns a Fixture, but also verifies that everything has been set up correctly.
func NewFixture(config FixtureConfig) (Fixture, error) {
	var fixture Fixture

	// check AppRoot is set
	if len(config.AppRoot) == 0 {
		return fixture, fmt.Errorf("AppRoot is missing")
	}

	// must be absolute path
	if !filepath.IsAbs(config.AppRoot) {
		return fixture, fmt.Errorf("Please use an absolute path, or try using something like filepath.Abs(\"./../\")")
	}

	// check AppRoot exists
	_, err := os.Stat(config.AppRoot)
	if os.IsNotExist(err) {
		return fixture, fmt.Errorf("AppRoot directory does not exist")
	} else if err != nil {
		return fixture, fmt.Errorf("Error determining if AppRoot exists: %s", err.Error())
	}

	// check wait for output set
	if config.AppSetup.WaitForOutputLine == "" {
		return fixture, fmt.Errorf("AppSetup.WaitForOutputLine has not been set")
	}

	// default timeout to 10 seconds
	if config.AppSetup.WaitTimeout <= 0 {
		config.AppSetup.WaitTimeout = time.Second * 10
	}

	fixture.config = config

	return fixture, nil
}
