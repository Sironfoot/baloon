package baloon_test

import (
	"testing"

	"github.com/sironfoot/baloon"
)

func TestNewScript(t *testing.T) {
	command := "Hello world"

	script := baloon.NewScript(command)
	if script.Type != baloon.ScriptTypeLiteral {
		t.Errorf("Should return Type as ScriptTypeLiteral but got %d", script.Type)
	}

	if script.Command != command {
		t.Errorf("Should return the command we set '%s' but got '%s'", command, script.Command)
	}
}

func TestNewScriptPath(t *testing.T) {
	command := "Hello world"

	script := baloon.NewScriptPath(command)
	if script.Type != baloon.ScriptTypePath {
		t.Errorf("Should return Type as ScriptTypePath but got %d", script.Type)
	}

	if script.Command != command {
		t.Errorf("Should return the command we set '%s' but got '%s'", command, script.Command)
	}
}
