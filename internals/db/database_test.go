package database_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	database "github.com/Vince-maple-byte/KeyData/internals/db"
	"github.com/Vince-maple-byte/KeyData/internals/record"
)

func TestDatabaseCreation(t *testing.T) {
	temp := t.TempDir()
	store := filepath.Join(temp, "store")
	wal := filepath.Join(temp, "wal")
	walPath := filepath.Join(wal, "memtable1.wal")

	err := os.Mkdir(store, 0600)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(wal, 0600)
	if err != nil {
		t.Fatal(err)
	}

	db, err := database.CreateDatabase(store, filepath.Join(wal, "memtable1.wal"))

	if err != nil {
		t.Fatalf("Error in creating the database struct:\n%v", err)
	}

	if db.Dir != store {
		t.Errorf("Incorrect data directory saved\nExpected:%s\nActual:%s", store, db.Dir)
	}

	if db.WalPath != walPath {
		t.Errorf("Incorrect data directory saved\nExpected:%s\nActual:%s", wal, db.WalPath)
	}

	if db.Memtable == nil {
		t.Error("Unable to create the Memtable")
	}

	if db.SSTFiles == nil {
		t.Error("Unable to create the SSTable slice")
	}
}

func TestDatabaseCorrectNumberOfSSTFiles(t *testing.T) {
	temp := t.TempDir()
	store := filepath.Join(temp, "store")
	wal := filepath.Join(temp, "wal")

	err := os.Mkdir(store, 0600)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(wal, 0600)
	if err != nil {
		t.Fatal(err)
	}

	db, err := database.CreateDatabase(store, filepath.Join(wal, "memtable1.wal"))

	if err != nil {
		t.Fatalf("Error in creating the database struct:\n%v", err)
	}

	for i := range 3200 * 3 {
		ok, err := db.Put(fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i))

		if !ok {
			t.Fatalf("Something went wrong when adding data into the database:\n%v", err)
		}
	}

	dir, err := os.ReadDir(store)

	if err != nil {
		t.Fatalf("Not able to read into the directory\n%v", err)
	}

	if len(dir) != len(db.SSTFiles) {
		t.Errorf("The length of the database struct: %d, does not match the length of the actual amount of file stored in the directory: %d", len(db.SSTFiles), len(dir))
	}

	for _, file := range db.SSTFiles {
		file.Close()
	}
}

func TestDatabaseAfterCompactionHappens(t *testing.T) {
	temp := t.TempDir()
	store := filepath.Join(temp, "store")
	wal := filepath.Join(temp, "wal")

	err := os.Mkdir(store, 0600)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(wal, 0600)
	if err != nil {
		t.Fatal(err)
	}

	db, err := database.CreateDatabase(store, filepath.Join(wal, "memtable1.wal"))

	if err != nil {
		t.Fatalf("Error in creating the database struct:\n%v", err)
	}

	for range 3200 * 4 {
		ok, err := db.Put("key", "val")

		if !ok {
			t.Fatalf("Something went wrong when adding data into the database:\n%v", err)
		}

		if err != nil {
			t.Errorf("Received an error here:\n%v", err)
		}
	}

	dir, err := os.ReadDir(store)

	if err != nil {
		t.Fatalf("Not able to read into the directory\n%v", err)
	}

	if len(db.SSTFiles) != 1 {
		t.Errorf("The length of the database struct: %d, does not match the length of the actual amount of file stored in the directory: %d", len(db.SSTFiles), len(dir))
	}

	for _, file := range db.SSTFiles {
		t.Logf("File name: %s", file.FileName)
		file.Close()
	}
}

func TestDatabasePutValue(t *testing.T) {

	temp := t.TempDir()
	store := filepath.Join(temp, "store")
	wal := filepath.Join(temp, "wal")

	err := os.Mkdir(store, 0700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(wal, 0700)
	if err != nil {
		t.Fatal(err)
	}

	db, err := database.CreateDatabase(store, filepath.Join(wal, "memtable1.wal"))

	if err != nil {
		t.Fatalf("Error in creating the database struct:\n%v", err)
	}

	ok, err := db.Put("key1", "val1")

	if !ok {
		t.Fatalf("Unable to add a new key/value pair\nError:%v", err)
	}

	if err != nil {
		t.Fatalf("Error:%v", err)
	}

	if db.Memtable.Size < 1 {
		t.Errorf("Did not properly add the key/value pair into the memtable\nExpected Size:%d\nActual Size:%d",
			1, db.Memtable.Size)
	}
}

func TestDatabaseGetValue(t *testing.T) {

	temp := t.TempDir()
	store := filepath.Join(temp, "store")
	wal := filepath.Join(temp, "wal")

	err := os.Mkdir(store, 0700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(wal, 0700)
	if err != nil {
		t.Fatal(err)
	}

	db, err := database.CreateDatabase(store, filepath.Join(wal, "memtable1.wal"))

	if err != nil {
		t.Fatalf("Error in creating the database struct:\n%v", err)
	}

	for i := range 10 {
		ok, err := db.Put(fmt.Sprint("key", i), fmt.Sprint("val", i))

		if !ok {
			t.Fatalf("Unable to add a new key/value pair\nError:%v", err)
		}

		if err != nil {
			t.Fatalf("Error:%v", err)
		}
	}

	for i := range 10 {
		value, err := db.Get(fmt.Sprint("key", i))

		if err != nil {
			t.Errorf("Unable to get the value for %s key\nError:%v", fmt.Sprint("key", i), err)
		}

		if record.GetContents(value).Payload != fmt.Sprint("val", i) {
			t.Errorf("Improper value saved\nExpected:%s\nActual:%s", fmt.Sprint("val", i), record.GetContents(value).Payload)
		}
	}

	db.Close()
}

func TestDatabaseDeleteValue(t *testing.T) {
	temp := t.TempDir()
	store := filepath.Join(temp, "store")
	wal := filepath.Join(temp, "wal")

	err := os.Mkdir(store, 0700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(wal, 0700)
	if err != nil {
		t.Fatal(err)
	}

	db, err := database.CreateDatabase(store, filepath.Join(wal, "memtable1.wal"))

	if err != nil {
		t.Fatalf("Error in creating the database struct:\n%v", err)
	}

	for i := range 10 {
		ok, err := db.Put(fmt.Sprint("key", i), fmt.Sprint("val", i))

		if !ok {
			t.Fatalf("Unable to add a new key/value pair\nError:%v", err)
		}

		if err != nil {
			t.Fatalf("Error:%v", err)
		}
	}

	for i := range 10 {

		_, err = db.Delete(fmt.Sprint("key", i))

		if err != nil {
			t.Errorf("Unable to get the value for %s key\nError:%v", fmt.Sprint("key", i), err)
		}

		//The memtable should return an error stating that the record is deleted
		_, err := db.Get(fmt.Sprint("key", i))

		if err == nil {
			t.Fatal("Expecting an error in which stating that the record has been deleted")
		}
	}
}

//Test for restarting the Database with the Wal file being able to restart the Memtable
// Test for restarting the Database where the SSTables will be loaded into the SSTFile array
