package baloon

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"
)

// Fixture represents a test fixture. You usually have one per test suite.
type Fixture struct {
	config        FixtureConfig
	TestSetups    []TestSetup
	TestTeardowns []TestTeardown

	appPath                  string
	appProcess               *exec.Cmd
	alreadyAttemptedSetup    bool
	alreadyAttemptedTeardown bool
}

// Setup runs the fixture setup. Call this only once before running all your tests, usually in func MainTest()
func (fixture *Fixture) Setup() error {
	if fixture.alreadyAttemptedSetup {
		return fmt.Errorf("Setup() has already been called. Only run this function once for the test suite.")
	}

	fixture.alreadyAttemptedSetup = true

	for i, dbSetup := range fixture.config.DatabaseSetups {
		err := dbSetup.Run(fixture.config.AppRoot)
		if err != nil {
			return fmt.Errorf("Error running Database Setup at index %d: %s", i, err.Error())
		}
	}

	appRoot := fixture.config.AppRoot
	appName := path.Base(appRoot)

	// build app
	cmdFile, err := ioutil.TempFile(appRoot, appName+"_")
	if err != nil {
		return fmt.Errorf("Error generating random file for program: %s", err.Error())
	}

	appPath := cmdFile.Name()
	fixture.appPath = appPath
	cmdFile.Close()

	cmd := exec.Command("go", "build", "-o", appPath)
	cmd.Dir = appRoot

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Error building program: %s", err.Error())
	}

	// run app
	appProcess := exec.Command(appPath, fixture.config.AppSetup.Arguments...)
	appProcess.Dir = appRoot

	fixture.appProcess = appProcess

	outReader, err := appProcess.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Error getting stdout pipe from running program: %s", err.Error())
	}
	defer outReader.Close()

	errReader, err := appProcess.StderrPipe()
	if err != nil {
		return fmt.Errorf("Error getting stderr pipe from running program: %s", err.Error())
	}
	defer errReader.Close()

	outScanner := bufio.NewScanner(outReader)
	errScanner := bufio.NewScanner(errReader)

	err = appProcess.Start()
	if err != nil {
		return fmt.Errorf("Error running program under test: %s", err.Error())
	}

	outDone := make(chan struct{})
	go func() {
		for outScanner.Scan() {
			if outScanner.Text() == fixture.config.AppSetup.WaitForOutputLine {
				close(outDone)
				break
			}
		}
	}()

	errDone := make(chan struct{})
	go func() {
		for errScanner.Scan() {
			if errScanner.Text() == fixture.config.AppSetup.WaitForOutputLine {
				close(errDone)
				break
			}
		}
	}()

	select {
	case <-outDone:
		return nil
	case <-errDone:
		return nil
	case <-time.After(fixture.config.AppSetup.WaitTimeout):
		return fmt.Errorf("Timeout waiting for program to start. Was looking for output line \"%s\".",
			fixture.config.AppSetup.WaitForOutputLine)
	}
}

// Teardown runs the fixture teardown routines. Call this only once after running all your tests,
// usually in func MainTest() after the call to m.Run()
func (fixture *Fixture) Teardown() error {
	if !fixture.alreadyAttemptedSetup {
		return fmt.Errorf("Please run Setup() first before calling Teardown()")
	}

	if fixture.alreadyAttemptedTeardown {
		return fmt.Errorf("Teardown() has already been called. Only run this function once for the test suite.")
	}

	fixture.alreadyAttemptedTeardown = true

	// shut down app
	err := fixture.appProcess.Process.Kill()
	if err != nil {
		return fmt.Errorf("Error shutting down program: %s", err.Error())
	}

	// delete program file
	err = os.Remove(fixture.appPath)
	if err != nil {
		return fmt.Errorf("Error trying to delete complile binary: %s", err.Error())
	}

	// run database teardown
	for i, dbSetup := range fixture.config.DatabaseTeardowns {
		err := dbSetup.Run(fixture.config.AppRoot)
		if err != nil {
			return fmt.Errorf("Error running Database Teardown at index %d: %s", i, err.Error())
		}
	}

	return nil
}

// AddTestSetup adds a TestSetup routine to the test Fixture
func (fixture *Fixture) AddTestSetup(setup TestSetup) {
	fixture.TestSetups = append(fixture.TestSetups, setup)
}

// TestSetup will run all TestSetup routines. This is run at the start of each individual test,
// e.g. func TestSomething(t *testing.T), within your test suite.
func (fixture *Fixture) TestSetup() error {
	if !fixture.alreadyAttemptedSetup {
		return fmt.Errorf("Please run Setup() first before calling TestSetup()")
	}

	if fixture.alreadyAttemptedTeardown {
		return fmt.Errorf("Fixture has already been teared down")
	}

	for i, testSetup := range fixture.TestSetups {
		for dbIndex, dbSetup := range testSetup.DatabaseSetups {
			err := dbSetup.Run(fixture.config.AppRoot)
			if err != nil {
				return fmt.Errorf("Error running Database Setup at index %d for TestSetup at index %d: %s",
					dbIndex, i, err.Error())
			}
		}

		if testSetup.Func != nil {
			testSetup.Func()
		}
	}

	return nil
}

// AddTestTeardown adds a TestTeardown routine to the test Fixture
func (fixture *Fixture) AddTestTeardown(teardown TestTeardown) {
	fixture.TestTeardowns = append(fixture.TestTeardowns, teardown)
}

// TestTeardown will run all TestTeardown routines. This is run at the end of each individual test,
// e.g. func TestSomething(t *testing.T), within your test suite.
func (fixture *Fixture) TestTeardown() error {
	if !fixture.alreadyAttemptedSetup {
		return fmt.Errorf("Please run Setup() first before calling TestTeardown()")
	}

	if fixture.alreadyAttemptedTeardown {
		return fmt.Errorf("Fixture has already been teared down")
	}

	for i, testTeardown := range fixture.TestTeardowns {
		for dbIndex, dbSetup := range testTeardown.DatabaseTeardowns {
			err := dbSetup.Run(fixture.config.AppRoot)
			if err != nil {
				return fmt.Errorf("Error running Database Setup at index %d for TestTeardown at index %d: %s",
					dbIndex, i, err.Error())
			}
		}

		if testTeardown.Func != nil {
			testTeardown.Func()
		}
	}

	return nil
}

// Close will attempt to free up any resources created by the Fixture.
// Make sure to call this before any log.Fatal() or os.Exit() calls.
func (fixture *Fixture) Close() {
	// attempt to clean things up as best we can
	if !fixture.alreadyAttemptedTeardown {
		fixture.Teardown()
	}

	// kill process if it's running
	if fixture.appProcess != nil && fixture.appProcess.Process != nil {
		fixture.appProcess.Process.Kill()
	}

	// delete executable if it exists
	_, err := os.Stat(fixture.appPath)
	if err == nil {
		os.Remove(fixture.appPath)
	}
}