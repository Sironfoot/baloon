# baloon
Baloon is a Setup and Teardown test fixture library for end-to-end testing of HTTP APIs written in Go.

> **WARNING! Baloon API is unstable and may undergo changes. Please use with caution for the time being.**

Baloon will setup a database with your sample data, compile and run your Go executable, run your tests, and teardown your database afterwards. It also supports setup and teardown routines per unit test.

Baloon is designed to be used in conjunction with an API testing library such [baloo](https://github.com/h2non/baloo) (which inspired me to write this test fixture library, hence the name baloon). The goal is to make HTTP API testing less brittle by providing clean and repeatable setup and teardown routines for your main external dependencies, namely databases and compiling/running your program.

## Installation

```bash
go get github.com/sironfoot/baloon
```

## Requirements

- Go 1.7+

## Setup

Baloon needs to be run in your Go test's TestMain function. TestMain is a special test function in Go that will run instead of your tests, providing you with the opportunity to run setup and teardown code before running your tests.

```go
func TestMain(m *testing.M) {
	// insert setup code here

	// run all our tests
	code := m.Run()

	// insert teardown code here

	// exit
	os.Exit(code)
}
```
In this example, we're assuming we have our tests inside a /tests directory at the root of our Go HTTP app. Create a main_test.go file in this dir, this is where we put our TestMain function that includes all our setup code. All code listed below goes in this function.

#### 1. App Root

Lets start by getting an absolute path to your Go app's root directory (containing your main.go):

```go
import "path/filepath"

func TestMain(m *testing.M) {
	appRootPath, err := filepath.Abs("./../")
}
```

#### 2. Database Setup

Now we setup our fixture. Create one or more Database setup routines.

```go
databaseSetups := []baloon.DB{
	// create the initial database
	baloon.DB{
		Connection: baloon.DBConn{
			Driver: "postgres",
			String: "postgres://user:pw@localhost:5432/?sslmode=disable",
		},
		Scripts: []baloon.Script{
			baloon.NewScript("CREATE DATABASE northwind;"),
		},
	},
	// setup tables, stored procedures etc.
	baloon.DB{
		Connection: baloon.DBConn{
			Driver: "postgres",
			String: "postgres://user:pw@localhost:5432/northwind?sslmode=disable",
		},
		Scripts: []baloon.Script{
			baloon.NewScriptPath("./sql/create tables.sql"),
			baloon.NewScriptPath("./sql/create functions.sql")
			baloon.NewScriptPath("./sql/create sprocs.sql")
		},
	},
}
```

We have 2 setups here because we need to connect the the database server instance (sans any particular database) to first create a database, then a second setup connects to our newly created database to add the tables, sprocs etc.

Baloon uses the "database/sql" package, so will support any database that supports that, but make sure your database driver is imported:

```go
import "_ "github.com/lib/pq"
```

Scripts can be literal scripts (```CREATE DATABASE northwind;```), or paths to files containing scripts (```./sql/create tables.sql```). Paths are relative to your app root (see 1. App Root above). Paths support globbing patterns (e.g. ```./sql/*.sql```).

#### 3. App Executable Setup

Here we provide instructions on how to run our Go HTTP API executable.

```go
appSetup := baloon.App{
	Arguments: []string{
		"-port", "8080",
		"-db_name", "northwind",
		"-db_port", "5432",
		"-ready_statement", "Test App is Ready",
	},
	WaitForOutputLine: "Test App is Ready",
	WaitTimeout:       5 * time.Second,
}
```

Baloon will automatically compile our app into the root dir (with a random filename) using ```go build -o "./filename"```. It will run our app with the arguments provided, and delete our app executable afterwards.

WaitForOutputLine tells Baloon to wait for a line of text to appear in the stdout or stderr to signal that our app is ready to start accepting HTTP requests. So configure your app to output an appropriate line, or use the standard ```Listening and serving HTTP on :8080``` that most Go HTTP Web frameworks output. If our app takes a few seconds to startup & initialise, we don't want tests executing against our app before it's ready.

#### 4. Database Teardown

Same as setup but runs after all our tests have finished. Here we just delete our database.

```go
databaseTeardowns := []baloon.DB{
	baloon.DB{
		Connection: baloon.DBConn{
			Driver: "postgres",
			String: "postgres://user:pw@localhost:5432/?sslmode=disable",
		},
		Scripts: []baloon.Script{
			baloon.NewScript("DROP DATABASE IF EXISTS northwind;"),
		},
	},
}
```

#### 5. Putting It All Together

```go
var fixture baloon.Fixture

func TestMain(m *testing.M) {
    // code from above goes here

	setup := baloon.FixtureConfig{
		AppRoot: appRoot,
		DatabaseSetups: databaseSetups,
		AppSetup: appSetup,
		DatabaseTeardowns: databaseTeardowns
	}

	fixture, err = baloon.NewFixture(setup)
	if err != nil {
		log.Fatal(err)
	}

	err = fixture.Setup()
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	err = fixture.Teardown()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}
```

#### 6. Per Unit Test Setup and teardown

We can run setup and teardown routines per individual unit test. A use case is to add sample data to our database to test against, but have that data reset after each test, as some tests might insert or delete data.

We can also run bespoke code during unit test setup and teardown. For instance, getting an example admin and non-admin user ID if our primary key IDs are auto-generated and therefore will be different after each setup.

In the TestMain func after creating a new Fixture:

```go
fixture, err = baloon.NewFixture(setup)
if err != nil {
	log.Fatal(err)
}

fixture.AddTestSetup(baloon.TestSetup{
	DatabaseSetups: []baloon.DB{
		baloon.DB{
			Connection: baloon.DBConn{
				Driver: "postgres",
				String: "postgres://user:pw@localhost:5432/northwind?sslmode=disable",
			},
			Scripts: []baloon.Script{
				baloon.NewScriptPath("./tests/testData/*.sql"),
			},
		},
	},
	Func: func() {
		adminUserID = getAdminUserID("admin@example.com")
		stdUserID = getStandardUserID("user@example.com")
	},
})

fixture.AddTestTeardown(baloon.TestTeardown{
	DatabaseTeardowns: []baloon.DB{
		baloon.DB{
			Connection: baloon.DBConn{
				Driver: "postgres",
				String: "postgres://user:pw@localhost:5432/northwind?sslmode=disable",
			},
			Scripts: []baloon.Script{
				baloon.NewScript("DELETE FROM orders;"),
				baloon.NewScript("DELETE FROM customers;"),
				baloon.NewScript("DELETE FROM products;"),
			},
		},
	},
	Func: func() { },
})
```

Then use these in each unit test:

```go
func TestCustomers_List(t *testing.T) {
	fixture.TestSetup()
	defer fixture.TestTeardown()

	// test code Here
}
```
## Tips

#### Dropping Database Connections

Drop database commands can fail if there are open/active connections to the database. To make the setup and teardown more reliable, it's a good idea to drop these connections as well. The below example is for postgres:

```go
sqlDropConnections :=
	`SELECT pg_terminate_backend(pg_stat_activity.pid)
	 FROM pg_stat_activity
	 WHERE  pg_stat_activity.datname = 'northwind'
	 	AND pid <> pg_backend_pid();`

databaseTeardowns := []baloon.DB{
	baloon.DB{
		Connection: baloon.DBConn{
			Driver: "postgres",
			String: "postgres://user:pw@localhost:5432/?sslmode=disable",
		},
		Scripts: []baloon.Script{
			baloon.NewScript(sqlDropConnections)
			baloon.NewScript("DROP DATABASE IF EXISTS northwind;"),
		},
	},
}
```
#### Drop Database During Setup

Further to the above, it's advisable to attempt to drop any databases during setup as well as teardown. The reasoning is that if our setup/teardown routines fail, we might be left with the Test database still alive, causing database setup routines to fail trying to create a database that already exists.

```go
databaseSetups := []baloon.DB{
	// create the initial database
	baloon.DB{
		Connection: baloon.DBConn{
			Driver: "postgres",
			String: "postgres://user:pw@localhost:5432/?sslmode=disable",
		},
		Scripts: []baloon.Script{
			baloon.NewScript(sqlDropConnections)
			baloon.NewScript("DROP DATABASE IF EXISTS northwind;"),
			baloon.NewScript("CREATE DATABASE northwind;"),
		},
	},
	// setup tables, stored procedures etc.
	// ...snip
}
```

#### Custom Setup and Teardown Code

If you want to run any custom setup and teardown code, simply add it to your TestMain() func.

```go
func TestMain(m *testing.M) {
	// fixture setup config (snip...)
	fixture, err = baloon.NewFixture(setup)

	// custom setup code goes here..
	err = fixture.Setup()
	// ...or here

	code := m.Run()

	// custom teardown code goes here...
	err = fixture.Teardown()
	// ...or here

	os.Exit(code)
}
```

# Licence

MIT - Dominic Pettifer
