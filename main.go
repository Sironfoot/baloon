package baloon

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// ScriptTypeLiteral ...
	ScriptTypeLiteral = 1
	// ScriptTypePath ...
	ScriptTypePath = 2
)

// DBConn ...
type DBConn struct {
	Driver string
	String string
}

// Script ...
type Script struct {
	Type    int
	Command string
}

// NewScript ...
func NewScript(command string) Script {
	return Script{
		Type:    ScriptTypeLiteral,
		Command: command,
	}
}

// NewScriptPath ...
func NewScriptPath(path string) Script {
	return Script{
		Type:    ScriptTypePath,
		Command: path,
	}
}

// App ...
type App struct {
	Arguments         []string
	WaitForOutputLine string
	WaitTimeout       time.Duration
}

// FixtureConfig ...
type FixtureConfig struct {
	AppRoot           string
	DatabaseSetups    []DB
	AppSetup          App
	DatabaseTeardowns []DB
}

// TestSetup ...
type TestSetup struct {
	DatabaseSetups []DB
	Func           func()
}

// TestTeardown ...
type TestTeardown struct {
	DatabaseTeardowns []DB
	Func              func()
}

func truncate(text string, maxLength int, affix string) string {
	if len(text) <= maxLength {
		return text
	}

	return text[:maxLength] + affix
}

// NewFixture ...
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
