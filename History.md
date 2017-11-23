# Changelog

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
