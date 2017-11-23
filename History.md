# Changelog

## v2.0.0 - 2017-11-23

- **BREAKING CHANGE**: unit test setup/teardown functions accept a `testing.T` and instead of returning an error, will call T.Fatal() on database errors.
- **BREAKING CHANGE**: custom unit test setup/teardown code can make use of a `testing.T` struct from the unit test to perform error handling.

## v1.0.1 - 2017-11-23

- fixed: prevent errors during teardown when App process is shut down or doesn't exist.
- fixed: check file exists before trying to delete it.
- fixed: docs should use 'defer fixture.Close()' use case.

## v1.0.0 - 2017-11-20

- added: Support for custom Go build commands

## v0.1.0 - 2017-11-16

- First release
- added: Setup/teardown database
- added: Build/run app executable
- added: Per unit test setup/teardown
