package memtable

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Vince-maple-byte/KeyData/internals/record"
)

func TestWriteForSingleEntryInMemtable(t *testing.T) {

	baseDir := t.TempDir()

	walDir := filepath.Join(baseDir, "dir1")
	os.Mkdir(walDir, 0700)

	dataDir := filepath.Join(baseDir, "dir2")
	os.Mkdir(dataDir, 0700)

	mem := CreateMemtable(walDir, dataDir)
	prevSize := mem.size
	_, err := mem.Write("key", "val", "PUT")

	if err != nil {
		t.Fatalf("Was not able to process the delete operation for the memtable")
	}

	if mem.size <= prevSize {
		t.Errorf("The write operation did not successfully increase the memtable size")
	}
}

func TestWriteForMultipleEntriesInMemtable(t *testing.T) {
	baseDir := t.TempDir()

	walDir := filepath.Join(baseDir, "dir1")
	os.Mkdir(walDir, 0700)

	dataDir := filepath.Join(baseDir, "dir2")
	os.Mkdir(dataDir, 0700)
	mem := CreateMemtable(walDir, dataDir)

	tests := []struct {
		key, value, op string
	}{
		{"key1", "val1", "PUT"},
		{"key2", "val2", "PUT"},
		{"key3", "val3", "PUT"},
		{"key4", "val4", "PUT"},
		{"key5", "val5", "PUT"},
	}

	for _, test := range tests {
		prevSize := mem.size
		_, err := mem.Write(test.key, test.value, test.op)

		if err != nil {
			t.Fatalf("Was not able to process the delete operation for the memtable")
		}

		if mem.size <= prevSize {
			t.Errorf("The delete operation did not successfully increase the memtable size")
		}
	}

}

func TestGetForMemtable(t *testing.T) {
	baseDir := t.TempDir()

	walDir := filepath.Join(baseDir, "dir1")
	os.Mkdir(walDir, 0700)

	dataDir := filepath.Join(baseDir, "dir2")
	os.Mkdir(dataDir, 0700)
	mem := CreateMemtable(walDir, dataDir)

	mem.Write("key", "val", "PUT")
	actual, err := mem.Get("key")

	if err != nil {
		t.Errorf("The read was not able to complete (could mean that the key was not saved into the memtable)")
	}

	if record.GetContents(actual).Key != "key" {
		t.Errorf("The key that was retrieved was not the same as the key/value pair that was written to the memtable")
	}
}

func TestDeleteForMemtable(t *testing.T) {
	baseDir := t.TempDir()

	walDir := filepath.Join(baseDir, "dir1")
	os.Mkdir(walDir, 0700)

	dataDir := filepath.Join(baseDir, "dir2")
	os.Mkdir(dataDir, 0700)
	mem := CreateMemtable(walDir, dataDir)

	mem.Write("key", "val", "PUT")
	prevSize := mem.size
	_, err := mem.Write("key", "", "DELETE")

	if err != nil {
		t.Fatalf("Was not able to process the delete operation for the memtable")
	}

	if mem.size <= prevSize {
		t.Errorf("The delete operation did not successfully increase the memtable size")
	}

}

func TestWrite_FlushResetsState(t *testing.T) {
	t.Skip("requires sstable wiring and writable data dir — run as integration test")

	baseDir := t.TempDir()

	walDir := filepath.Join(baseDir, "dir1")
	os.Mkdir(walDir, 0700)

	dataDir := filepath.Join(baseDir, "dir2")
	os.Mkdir(dataDir, 0700)
	mt := CreateMemtable(walDir, dataDir)

	for i := 0; i < MAX_SIZE; i++ {
		key := "key" + string(rune(i))
		_, err := mt.Write(key, "value", "PUT")
		if err != nil {
			t.Fatalf("unexpected error at write %d: %v", i, err)
		}
	}

	if mt.size != 0 {
		t.Errorf("expected size to reset to 0 after flush, got %d", mt.size)
	}
}
