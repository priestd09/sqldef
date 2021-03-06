// Integration test of mysqldef command.
//
// Test requirement:
//   - go command
//   - `mysql -uroot` must succeed
package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMysqldefDryRun(t *testing.T) {
	resetTestDatabase()
	writeFile("schema.sql", `
	  CREATE TABLE users (
	    name varchar(40),
	    created_at datetime NOT NULL
	  );
	`)

	dryRun := assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--dry-run", "--file", "schema.sql")
	apply := assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--file", "schema.sql")
	assertEquals(t, dryRun, "--- dry run ---\n"+apply)
}

func TestMysqldefExport(t *testing.T) {
	resetTestDatabase()
	out := assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--export")
	assertEquals(t, out, "-- No table exists\n")

	mustExecute("mysql", "-uroot", "mysqldef_test", "-e", `
	  CREATE TABLE users (
	    name varchar(40),
	    created_at datetime NOT NULL
	  );
	`)
	out = assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--export")
	assertEquals(t, out,
		"CREATE TABLE `users` (\n"+
			"  `name` varchar(40) DEFAULT NULL,\n"+
			"  `created_at` datetime NOT NULL\n"+
			") ENGINE=InnoDB DEFAULT CHARSET=latin1;\n",
	)
}

func TestMysqldefHelp(t *testing.T) {
	_, err := execute("mysqldef", "--help")
	if err != nil {
		t.Errorf("failed to run --help: %s", err)
	}

	out, err := execute("mysqldef")
	if err == nil {
		t.Errorf("no database must be error, but successfully got: %s", out)
	}
}

func TestMain(m *testing.M) {
	resetTestDatabase()
	mustExecute("go", "build")
	status := m.Run()
	os.Exit(status)
}

func mustExecute(command string, args ...string) string {
	out, err := execute(command, args...)
	if err != nil {
		log.Printf("command: '%s %s'", command, strings.Join(args, " "))
		log.Printf("out: '%s'", out)
		log.Fatal(err)
	}
	return out
}

func assertedExecute(t *testing.T, command string, args ...string) string {
	out, err := execute(command, args...)
	if err != nil {
		t.Errorf("failed to execute '%s %s' (out: '%s'): %s", command, strings.Join(args, " "), out, err)
	}
	return out
}

func assertEquals(t *testing.T, actual string, expected string) {
	if expected != actual {
		t.Errorf("expected '%s' but got '%s'", expected, actual)
	}
}

func execute(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	out, err := cmd.Output()
	return string(out), err
}

func resetTestDatabase() {
	mustExecute("mysql", "-uroot", "-e", "DROP DATABASE IF EXISTS mysqldef_test;")
	mustExecute("mysql", "-uroot", "-e", "CREATE DATABASE mysqldef_test;")
}

func writeFile(path string, content string) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	file.Write(([]byte)(content))
}
