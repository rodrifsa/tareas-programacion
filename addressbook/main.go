package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

/*
var (
a = 1
b = 2
...
)
is a shortcut for
var a = 1
var b = 2
...
*/
var (
	// here we define the subcommands and the flags.
	// Aquí definimos los subcomandos y las banderas.
	// if you run the program with
	// Si ejecutas el programa con

	// addressbook.exe delete -id=2

	// then 'delete' is the subcommand
	// entonces 'delete' es el subcomando

	// and '-id=2' is a flag with a value
	// y '-id=2' es una bandera con un valor

	// the subcommand 'select'
	//el subcomando 'select'

	selectCmd = flag.NewFlagSet("select", flag.ExitOnError)
	// the subcommand 'insert'
	// el subcomando 'insetar'
	insertCmd = flag.NewFlagSet("insert", flag.ExitOnError)
	// the flag '-lastname' for the subcommand 'insert'
	// La bandera '-lastname' para el subcomando 'insertar'.
	insertLastnameArg = insertCmd.String("lastname", "", "the lastname of the contact")
	// the flag '-firstname' for the subcommand 'insert'
	// La bandera '-firstname' para el subcomando 'insertar'.
	insertFirstnameArg = insertCmd.String("firstname", "", "the firstname of the contact")
	// the flag '-dayofbirth' for the subcommand 'insert'
	// La bandera '-dayofbirth' para el subcomando 'insertar'.
	insertDayOfBirthArg = insertCmd.String("dayofbirth", "", "the day of birth in the format YYYY-MM-DD")
	// the subcommand 'update'
	// el subcomando 'update'
	updateCmd           = flag.NewFlagSet("update", flag.ExitOnError)
	updateIdArg         = updateCmd.Int("id", 0, "id of the contact to be updated")
	updateFirstnameArg  = updateCmd.String("firstname", "", "the first name of the contact")
	updateLastnameArg   = updateCmd.String("lastname", "", "the last name of the contact")
	updateDayOfBirthArg = updateCmd.String("dayofbirth", "", "the day of birth of the contact")

	// the subcommand 'delete'
	deleteCmd = flag.NewFlagSet("delete", flag.ExitOnError)
	// the flag '-id' for the subcommand 'delete'
	deleteIdArg = deleteCmd.Int("id", 0, "id of the contact to be deleted")
	// the subcommand 'deleteall'

	deleteAllCmd = flag.NewFlagSet("delete-all", flag.ExitOnError)
)

// main is the main entry point of the program
func main() {
	// we let 'run' do all the work and just handle the error, if we get one
	err := run()
	// we got an error
	if err != nil {
		// print the error to os.Stderr which is the error output of the terminal
		fmt.Fprintf(os.Stderr, "an error happened: %v\n", err.Error())
		// exit the program with a error code (1)
		os.Exit(1)
	}
	// exit the program with code 0, which means: no error
	os.Exit(0)
}

// this is the function that does all the work for the main function
func run() error {
	// open the database
	db, err := openDatabase()
	if err != nil {
		return fmt.Errorf("could not open the database: %w", err)
	}

	// defer makes sure that the following function is executed, before
	// the run function returns
	// that means 'db.Close()' is executed and we don't need to care about
	// that anymore after that line
	defer db.Close()
	// createTable only creates the table, if it is not there yet
	createTable(db)
	// os.Args is a slice of all strings of the commandline
	// e.g. when running
	//
	// addressbook.exe delete -id=2
	//
	// os.Args would be []string{"addressbook.exe", "delete", "-id=2"}
	// len(x) returns the length of x, where x might be a slice, an array or a map.
	// len(os.Args) < 2 means os.Args just has one value (the program name),
	// therefor no subcommand is given
	if len(os.Args) < 2 {
		// we print here the content of the table (default)
		return selectContacts(db)
	}
	// since os.Args[0] is the program name, os.Args[1] has the subcommand (what comes after the program name)
	switch os.Args[1] {
	// the 'select' subcommand
	case "select":

		return selectContacts(db)
	// the 'insert' subcommand
	case "insert":
		// first parse the rest of os.Args to get the flags/arguments
		insertCmd.Parse(os.Args[2:])
		// create a new Contact struct
		var c Contact
		// set the properties of the Contact based on the flags/arguments
		c.Firstname = *insertFirstnameArg
		c.Lastname = *insertLastnameArg
		c.DayOfBirth = *insertDayOfBirthArg
		// if we got an empty Fristname...
		if c.Firstname == "" {
			// print the usage information of the program
			insertCmd.Usage()
			// and return the error messages
			return fmt.Errorf("pass '-firstname' ")
		}
		// if we got an empty Lastname...
		if c.Lastname == "" {
			// print the usage information of the program
			insertCmd.Usage()

			// and return the error messages
			return fmt.Errorf("pass '-lastname' ")
		}
		// if we got an empty DayOfBirth...
		if c.DayOfBirth == "" {
			// print the usage information of the program
			insertCmd.Usage()
			// and return the error messages
			return fmt.Errorf("pass '-dayofbirth' ")
		}
		// we need to add the time, so that sqlite could treat it as a datetime, see https://www.sqlitetutorial.net/sqlite-date/
		c.DayOfBirth = c.DayOfBirth + " 00:00:00.000"
		// if we are here, so errors happened so far.
		// just insert the new contact and return the error (which might be nil == no error)
		return insertContact(db, c)
	case "update":
		updateCmd.Parse(os.Args[2:])
		var c Contact
		c.ID = *updateIdArg
		c.Firstname = *updateFirstnameArg
		c.Lastname = *updateLastnameArg
		c.DayOfBirth = *updateDayOfBirthArg
		if c.ID == 0 {
			updateCmd.Usage()
			return fmt.Errorf("pass '-id'")
		}
		if c.Firstname == "" {
			updateCmd.Usage()
			return fmt.Errorf("pass '-firstname'")
		}
		if c.Lastname == "" {
			updateCmd.Usage()
			return fmt.Errorf("pass '-lastname'")
		}
		if c.DayOfBirth == "" {
			updateCmd.Usage()
			return fmt.Errorf("pass '-dayofbirth'")
		}

		return updateContact(db, c)

	// the 'delete' subcommand
	case "delete":
		// first parse the rest of os.Args to get the flags/arguments
		deleteCmd.Parse(os.Args[2:])
		var c Contact
		// get the id flag/argument

		c.ID = *deleteIdArg
		// if the id was not set (0) or not correctly set (<0)
		if c.ID <= 0 {
			// print the usage information of the program
			deleteCmd.Usage()
			// and return the error messages
			return fmt.Errorf("pass '-id'")
		}
		// if we are here, so errors happened so far.
		// just delete the contact and return the error (which might be nil == no error)
		return deleteContact(db, c)
	// the 'delete-all' subcommand
	case "delete-all":
		// just delete all contacts and return the error
		return deleteAllContacts(db)
	// some unknown/wrong subcommand
	default:
		// return the error
		return fmt.Errorf("expected 'select', 'insert', 'update', 'delete' or 'delete-all' subcommand")
	}
	return nil
}

// openDatabase opens the sqlite database and returns the DB struct of the sql
// standard library or an error, if the database could not be opened.
func openDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "addressbook.db")
	// an error happened: we return no database, but the error
	if err != nil {
		return nil, err
	}
	// no error happened: we return the db no error
	return db, nil
}

// createTable uses the given db to create a new contacts table
func createTable(db *sql.DB) error {
	// this is the sql code to create the table contacts
	sqlcreate := "" +
		"CREATE TABLE IF NOT EXISTS contacts (" +
		" id INTEGER PRIMARY KEY AUTOINCREMENT," +
		" firstname TEXT NOT NULL," +
		" lastname TEXT NOT NULL," +
		" dayofbirth TEXT NOT NULL," +
		" UNIQUE(dayofbirth,lastname,firstname) " +
		")"
	// this executes the sql code to create the table
	_, err := db.Exec(sqlcreate)
	// return the error (might be nil)

	return err
}

// Contact is a struct which can be used to collect the values of
// a dataset in the contacts table
type Contact struct {
	ID         int
	Firstname  string
	Lastname   string
	DayOfBirth string
}

// insertContact inserts a new given Contact into the contacts table
func insertContact(db *sql.DB, c Contact) error {
	// this is the sql code to insert into the contacts table
	sqlinsert := fmt.Sprintf("INSERT into contacts ( firstname, lastname, dayofbirth ) VALUES ('%s', '%s', '%s')",
		c.Firstname, c.Lastname, c.DayOfBirth)
	// this executes the sql code to insert the contact
	_, err := db.Exec(sqlinsert)
	return err
}

// deleteContact deletes the given contact
func deleteContact(db *sql.DB, c Contact) error {
	// this is the sql code to delete a dataset from the contacts table
	sqldelete := fmt.Sprintf("DELETE FROM contacts WHERE id = %v", c.ID)
	// this executes the sql code to delete the dataset

	_, err := db.Exec(sqldelete)
	return err
}

// deleteAllContacts deletes all data from the contacts table
func deleteAllContacts(db *sql.DB) error {
	// this is the sql code to delete all from the contacts table
	sqldeleteall := "DELETE FROM contacts"
	// this executes the sql code to delete all
	_, err := db.Exec(sqldeleteall)
	return err
}

// selectContacts selects all data from the contacts table
func selectContacts(db *sql.DB) error {
	// this is the sql code to select all from the contacts table
	sqlselect := "SELECT id, firstname, lastname, dayofbirth from contacts"
	// this executes the sql code to select all
	rows, err := db.Query(sqlselect)
	// we got an error, so return the error and leave the function
	if err != nil {
		return err
	}
	// print the column names

	// %-20s means:
	// - %s: print a string
	// - %20s: print a string and pad with space up to the length of 20
	// - %-20s: like %ns but but pad after the value (left-aligned)
	fmt.Fprintln(os.Stdout, fmt.Sprintf("| ID | %-20s | %-20s | %-20s |", strings.ToUpper("firstname"), strings.ToUpper("lastname"),
		strings.ToUpper("dayofbirth")))
	// rows.Next() returns true as long as there is data to scan
	// if it returns false, then the for loop stops.
	for rows.Next() {
		// create a new Contact
		var c Contact
		// scan the values to the fields of the Contact
		err = rows.Scan(&c.ID, &c.Firstname, &c.Lastname, &c.DayOfBirth)
		// we got an error, so return the error and leave the function
		if err != nil {
			return err
		}
		// print a new line
		dashes := strings.Repeat("-", 20)
		fmt.Fprintln(os.Stdout, fmt.Sprintf("| -- | %s | %s | %s |", dashes, dashes, dashes))
		// remove the time
		c.DayOfBirth = strings.Replace(c.DayOfBirth, " 00:00:00.000", "", 1)
		// print the data of the dataset
		// %-20s: see above (print column names)

		// %2d: pad the decimal (%d) with 2 zeros
		fmt.Fprintf(os.Stdout, "| %2d | %-20s | %-20s | %-20s |\n", c.ID, c.Firstname, c.Lastname, c.DayOfBirth)
	}
	// we got no error, so the returned error is nil
	return nil
}

func updateContact(db *sql.DB, c Contact) error {
	// this is the sql code to update a dataset from the contacts table
	sqlupdate := fmt.Sprintf("UPDATE contacts SET firstname= '%s', lastname= '%s', dayofbirth= '%s' WHERE id = %d", c.Firstname, c.Lastname, c.DayOfBirth, c.ID)
	// this executes the sql code to update the dataset

	_, err := db.Exec(sqlupdate)
	return err
}
