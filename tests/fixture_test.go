package baloon_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/sironfoot/baloon"
)

func TestNewFixture(t *testing.T) {
	testRootPath, _ := filepath.Abs("./")

	tests := []struct {
		Message       string
		Config        baloon.FixtureConfig
		ReturnsError  string
		ContainsError string
	}{
		{
			Message:      "Should return error when AppRoot is missing.",
			Config:       baloon.FixtureConfig{},
			ReturnsError: "AppRoot is missing",
		},
		{
			Message: "Should return error when AppRoot is not an absolute path.",
			Config: baloon.FixtureConfig{
				AppRoot: "./",
			},
			ContainsError: "Please use an absolute path",
		},
		{
			Message: "Should return error when AppRoot path doesn't exist.",
			Config: baloon.FixtureConfig{
				AppRoot: "/blah/blah/nonsense",
			},
			ReturnsError: "AppRoot directory does not exist",
		},
		{
			Message: "Should return error when AppSetup.WaitForOutputLine has not been set",
			Config: baloon.FixtureConfig{
				AppRoot: testRootPath,
			},
			ReturnsError: "AppSetup.WaitForOutputLine has not been set",
		},
	}

	for _, test := range tests {
		_, err := baloon.NewFixture(test.Config)
		if test.ReturnsError != "" {
			if err == nil || err.Error() != test.ReturnsError {
				t.Errorf(test.Message)
			}
		} else {
			if err == nil || !strings.Contains(err.Error(), test.ContainsError) {
				t.Errorf(test.Message)
			}
		}
	}
}
