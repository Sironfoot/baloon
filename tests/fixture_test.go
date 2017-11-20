package baloon_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestFixtureSetup(t *testing.T) {
	appRootPath, _ := filepath.Abs("./app/")

	// can't run multiple times
	fixture, err := baloon.NewFixture(baloon.FixtureConfig{
		AppRoot: appRootPath,
		AppSetup: baloon.App{
			RunArguments: []string{
				"-ready_statement", "Running",
			},
			WaitForOutputLine: "Running",
			WaitTimeout:       time.Second * 2,
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	err = fixture.Setup()
	if err != nil {
		t.Errorf("Should have run successfully, but got error: %s", err.Error())
	}

	err = fixture.Setup()
	if err == nil {
		t.Errorf("Should get an error if running Setup() more than once.")
	} else if err.Error() != "Setup() has already been called. Only run this function once for the test suite." {
		t.Errorf("Wrong error returned when running Setup() more than once. Error was: %s", err.Error())
	}

	fixture.Close()

	// check program runs correctly, and timeouts are dealt with
	fixture, err = baloon.NewFixture(baloon.FixtureConfig{
		AppRoot: appRootPath,
		AppSetup: baloon.App{
			RunArguments: []string{
				"-ready_statement", "Hello world",
			},
			WaitForOutputLine: "Running",
			WaitTimeout:       time.Millisecond * 100,
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	err = fixture.Setup()
	if err == nil {
		t.Errorf("Should return error about program timeout")
	} else if err.Error() != "Timeout waiting for program to start. Was looking for output line \"Running\"." {
		t.Errorf("Wrong error returned about program timeout. Error was: %s", err.Error())
	}

	fixture.Close()
}

func TestFixtureTeardown(t *testing.T) {
	appRootPath, _ := filepath.Abs("./app/")

	// can't run multiple times
	fixture, err := baloon.NewFixture(baloon.FixtureConfig{
		AppRoot: appRootPath,
		AppSetup: baloon.App{
			RunArguments: []string{
				"-ready_statement", "Running",
			},
			WaitForOutputLine: "Running",
			WaitTimeout:       time.Second * 2,
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	err = fixture.Setup()
	if err != nil {
		t.Fatal(err)
	}

	err = fixture.Teardown()
	if err != nil {
		t.Errorf("Should have run successfully, but got error: %s", err.Error())
	}

	err = fixture.Teardown()
	if err == nil {
		t.Errorf("Should return an error if attempting to run Teardown() twice")
	} else if err.Error() != "Teardown() has already been called. Only run this function once for the test suite." {
		t.Errorf("Wrong error returned when running Teardown() more than once. Error was: %s", err.Error())
	}

	fixture.Close()

	// can't run Teardown before Setup
	fixture, err = baloon.NewFixture(baloon.FixtureConfig{
		AppRoot: appRootPath,
		AppSetup: baloon.App{
			RunArguments: []string{
				"-ready_statement", "Running",
			},
			WaitForOutputLine: "Running",
			WaitTimeout:       time.Second * 2,
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	err = fixture.Teardown()
	if err == nil {
		t.Errorf("Should return an error if attempting to run Teardown() before Setup()")
	} else if err.Error() != "Please run Setup() first before calling Teardown()" {
		t.Errorf("Wrong error returned when running Teardown() before Setup(). Error was: %s", err.Error())
	}

	fixture.Close()
}

func TestBuildArguments(t *testing.T) {
	appRootPath, _ := filepath.Abs("./app/")
	programName := "baloon_test"

	fixture, err := baloon.NewFixture(baloon.FixtureConfig{
		AppRoot: appRootPath,
		AppSetup: baloon.App{
			BuildArguments: []string{
				"-o", "./" + programName,
			},
			RunArguments: []string{
				"-ready_statement", "Running",
			},
			WaitForOutputLine: "Running",
			WaitTimeout:       time.Second * 2,
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	err = fixture.Setup()
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(filepath.Join(appRootPath, programName))
	if err != nil && os.IsNotExist(err) {
		t.Errorf("Should generate program executable '%s' but wasn't present", programName)
	} else if err != nil {
		t.Fatal(err)
	}

	err = fixture.Teardown()
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(filepath.Join(appRootPath, programName))
	if err == nil || !os.IsNotExist(err) {
		t.Errorf("Should delete program executable afterward.")
	}
}
