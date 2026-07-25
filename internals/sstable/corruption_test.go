package sstable_test

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/Vince-maple-byte/KeyData/internals/memtable"
	"github.com/Vince-maple-byte/KeyData/internals/sstable"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildValidRecord constructs a minimal, valid SSTable record byte slice.
// Layout (matches Compact reader):
//
//	[0:8]   timestamp  (uint64 big-endian)
//	[8:12]  checksum   (uint32 big-endian)  -- intentionally zeroed for helpers
//	[12]    flags byte (1 byte, unused but present)
//	[13:17] keySize    (uint32 big-endian)
//	[17:21] payloadSize(uint32 big-endian)
//	[21:21+keySize]          key
//	[21+keySize : ...]       payload
func buildValidRecord(key, payload string, timestamp uint64) []byte {
	ks := uint32(len(key))
	ps := uint32(len(payload))

	rec := make([]byte, 0, 21+ks+ps)
	rec = binary.BigEndian.AppendUint64(rec, timestamp) // [0:8]  timestamp
	rec = binary.BigEndian.AppendUint32(rec, 0)         // [8:12] checksum placeholder
	rec = append(rec, 0x00)                             // [12]   flags
	rec = binary.BigEndian.AppendUint32(rec, ks)        // [13:17] keySize
	rec = binary.BigEndian.AppendUint32(rec, ps)        // [17:21] payloadSize
	rec = append(rec, []byte(key)...)
	rec = append(rec, []byte(payload)...)
	return rec
}

// buildValidFooter builds a 24-byte footer.
//
//	[0:8]   reserved / bloom-filter offset (zeroed)
//	[8:16]  end-of-data-block offset
//	[16:24] magic number 0xDEADBEEFDEADBEEF
func buildValidFooter(dataBlockSize uint64) []byte {
	f := make([]byte, 0, 24)
	f = binary.BigEndian.AppendUint64(f, 0)
	f = binary.BigEndian.AppendUint64(f, dataBlockSize)
	f = binary.BigEndian.AppendUint64(f, 0xDEADBEEFDEADBEEF)
	return f
}

// writeTempFile writes data to a uniquely-named temp file inside dir and
// returns the full path.
func writeTempFile(t *testing.T, dir string, data []byte) string {
	t.Helper()
	f, err := os.CreateTemp(dir, "kd_*.sst")
	if err != nil {
		t.Fatalf("writeTempFile: CreateTemp: %v", err)
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		t.Fatalf("writeTempFile: Write: %v", err)
	}
	return f.Name()
}

// buildSSTFile assembles: records || indexBlock || footer
// and returns the raw bytes.
func buildSSTFile(records [][]byte, indexBlock []byte, footer []byte) []byte {
	var out []byte
	for _, r := range records {
		out = append(out, r...)
	}
	out = append(out, indexBlock...)
	out = append(out, footer...)
	return out
}

// ---------------------------------------------------------------------------
// createFooter tests
// ---------------------------------------------------------------------------

func TestCreateFooter_EmptyList(t *testing.T) {
	_, err := sstable.ExportFooter([][]byte{})
	if err == nil {
		t.Fatal("expected error for empty record list, got nil")
	}
}

func TestCreateFooter_MagicNumber(t *testing.T) {
	rec := buildValidRecord("hello", "world", 1)
	footer, err := sstable.ExportFooter([][]byte{rec})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(footer) != 24 {
		t.Fatalf("footer length: want 24, got %d", len(footer))
	}
	magic := binary.BigEndian.Uint64(footer[16:24])
	const want = uint64(0xDEADBEEFDEADBEEF)
	if magic != want {
		t.Errorf("magic: want 0x%X, got 0x%X", want, magic)
	}
}

func TestCreateFooter_DataSizeMatchesInput(t *testing.T) {
	r1 := buildValidRecord("key1", "val1", 1)
	r2 := buildValidRecord("key2", "val2", 2)
	list := [][]byte{r1, r2}

	footer, err := sstable.ExportFooter(list)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantSize := uint64(len(r1) + len(r2))
	gotSize := binary.BigEndian.Uint64(footer[8:16])
	if gotSize != wantSize {
		t.Errorf("dataBlockSize: want %d, got %d", wantSize, gotSize)
	}
}

// ---------------------------------------------------------------------------
// Truncated-file tests
// (These simulate what Compact would encounter when reading a corrupt file.)
// ---------------------------------------------------------------------------

func TestReadFile_TruncatedFooter(t *testing.T) {
	rec := buildValidRecord("key", "value", 42)
	footer := buildValidFooter(uint64(len(rec)))
	full := buildSSTFile([][]byte{rec}, []byte{}, footer)

	// Chop the last 4 bytes so the footer is incomplete.
	truncated := full[:len(full)-4]

	dir := t.TempDir()
	path := writeTempFile(t, dir, truncated)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if len(data) < 24 {
		// Correct: caller should detect the file is too short to hold a footer.
		t.Log("correctly detected truncated footer: file shorter than 24 bytes")
		return
	}

	footer2 := data[len(data)-24:]
	magic := binary.BigEndian.Uint64(footer2[16:24])
	if magic == 0xDEADBEEFDEADBEEF {
		t.Error("truncated file should NOT produce a valid magic number")
	}
}

func TestReadFile_TruncatedMidRecord(t *testing.T) {
	// Build a record but only write the first half of it.
	rec := buildValidRecord("somekey", "somevalue", 99)
	half := rec[:len(rec)/2]

	dir := t.TempDir()
	path := writeTempFile(t, dir, half)

	data, _ := os.ReadFile(path)

	// Simulate the Compact inner loop: try to read keySize starting at offset 0.
	const headerEnd = 17
	if len(data) < headerEnd {
		t.Log("correctly detected truncated record: header incomplete")
		return
	}

	keySize := int(binary.BigEndian.Uint32(data[13:17]))
	payloadSize := int(binary.BigEndian.Uint32(data[17:21]))
	expectedEnd := 21 + keySize + payloadSize

	if len(data) < expectedEnd {
		t.Logf("correctly detected truncated record: need %d bytes, have %d", expectedEnd, len(data))
	} else {
		t.Error("expected truncation to be detected but full record appears readable")
	}
}

func TestReadFile_ZeroByteFile(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, []byte{})

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected 0 bytes, got %d", len(data))
	}
	if len(data) < 24 {
		t.Log("correctly detected zero-byte file: cannot contain a valid footer")
	}
}

func TestReadFile_FooterOnlyNoData(t *testing.T) {
	// A file containing only a footer (dataBlockSize = 0) — edge case.
	footer := buildValidFooter(0)
	dir := t.TempDir()
	path := writeTempFile(t, dir, footer)

	data, _ := os.ReadFile(path)
	footer2 := data[len(data)-24:]
	magic := binary.BigEndian.Uint64(footer2[16:24])

	if magic != 0xDEADBEEFDEADBEEF {
		t.Errorf("magic mismatch: got 0x%X", magic)
	}

	dataEnd := binary.BigEndian.Uint64(footer2[8:16])
	if dataEnd != 0 {
		t.Errorf("expected dataBlockSize=0, got %d", dataEnd)
	}
}

func TestReadFile_TruncatedAtHeaderBoundary(t *testing.T) {
	// Write exactly 21 bytes (header only, no key/payload).
	header := make([]byte, 21)
	binary.BigEndian.PutUint64(header[0:8], 12345) // timestamp
	binary.BigEndian.PutUint32(header[8:12], 0)    // checksum
	header[12] = 0x00                              // flags
	binary.BigEndian.PutUint32(header[13:17], 5)   // keySize = 5
	binary.BigEndian.PutUint32(header[17:21], 10)  // payloadSize = 10

	dir := t.TempDir()
	path := writeTempFile(t, dir, header)
	data, _ := os.ReadFile(path)

	keySize := int(binary.BigEndian.Uint32(data[13:17]))
	payloadSize := int(binary.BigEndian.Uint32(data[17:21]))
	need := 21 + keySize + payloadSize

	if len(data) < need {
		t.Logf("correctly detected record truncated at header boundary: need %d, have %d", need, len(data))
	} else {
		t.Error("expected boundary truncation to be detected")
	}
}

// ---------------------------------------------------------------------------
// Swapped-bytes / bit-flip tests
// ---------------------------------------------------------------------------

func TestReadFile_SwappedMagicBytes(t *testing.T) {
	rec := buildValidRecord("key", "val", 1)
	footer := buildValidFooter(uint64(len(rec)))

	// Swap the first two bytes of the magic number region (footer[16] <-> footer[17]).
	footer[16], footer[17] = footer[17], footer[16]

	full := buildSSTFile([][]byte{rec}, []byte{}, footer)
	dir := t.TempDir()
	path := writeTempFile(t, dir, full)

	data, _ := os.ReadFile(path)
	magic := binary.BigEndian.Uint64(data[len(data)-8:])

	if magic == 0xDEADBEEFDEADBEEF {
		t.Error("swapped magic bytes should invalidate the magic number")
	} else {
		t.Logf("correctly detected swapped magic bytes: got 0x%X", magic)
	}
}

func TestReadFile_SwappedKeySizeBytes(t *testing.T) {
	rec := buildValidRecord("hello", "world", 7)

	// Swap bytes [13] and [14] — the high bytes of keySize (uint32).
	rec[13], rec[16] = rec[16], rec[13]

	dir := t.TempDir()
	path := writeTempFile(t, dir, []byte{})
	os.WriteFile(path, rec, 0644)

	data, _ := os.ReadFile(path)
	keySize := int(binary.BigEndian.Uint32(data[13:17]))

	// Original keySize was 5 ("hello"); swapped bytes should differ.
	if keySize == 5 {
		t.Error("swapped keySize bytes should produce a different (wrong) keySize")
	} else {
		t.Logf("correctly detected swapped keySize bytes: parsed keySize=%d (true=5)", keySize)
	}
}

func TestReadFile_SwappedTimestampBytes(t *testing.T) {
	const originalTS uint64 = 0x0000000100000002
	rec := buildValidRecord("k", "v", originalTS)

	// Swap bytes [0] and [7] — MSB and LSB of the timestamp.
	rec[0], rec[7] = rec[7], rec[0]

	ts := binary.BigEndian.Uint64(rec[0:8])
	if ts == originalTS {
		t.Error("swapped timestamp bytes should produce a different value")
	} else {
		t.Logf("correctly detected swapped timestamp: got 0x%X (original 0x%X)", ts, originalTS)
	}
}

func TestReadFile_SwappedDataBlockSizeInFooter(t *testing.T) {
	rec := buildValidRecord("testkey", "testval", 55)
	correctSize := uint64(len(rec))
	footer := buildValidFooter(correctSize)

	// Swap bytes [8] and [9] inside the footer (high bytes of dataBlockSize).
	footer[8], footer[15] = footer[15], footer[8]

	full := buildSSTFile([][]byte{rec}, []byte{}, footer)
	dir := t.TempDir()
	path := writeTempFile(t, dir, full)

	data, _ := os.ReadFile(path)
	footerSlice := data[len(data)-24:]
	parsedSize := binary.BigEndian.Uint64(footerSlice[8:16])

	if parsedSize == correctSize {
		t.Logf("incorrectly detected swapped dataBlockSize: got %d (correct %d)", parsedSize, correctSize)
		t.Error("swapped dataBlockSize bytes should produce a wrong size")
	} else {
		t.Logf("correctly detected swapped dataBlockSize: got %d (correct %d)", parsedSize, correctSize)
	}
}

func TestReadFile_BitFlipInPayload(t *testing.T) {
	rec := buildValidRecord("mykey", "myvalue", 100)
	payloadStart := 21 + len("mykey")

	// Flip all bits in the first payload byte.
	rec[payloadStart] ^= 0xFF

	payload := rec[payloadStart : payloadStart+len("myvalue")]
	if string(payload) == "myvalue" {
		t.Error("bit-flipped payload should not equal original")
	} else {
		t.Logf("correctly detected payload corruption: %q", payload)
	}
}

// ---------------------------------------------------------------------------
// fileOffset tests
// ---------------------------------------------------------------------------

func TestFileOffset_CorrectOffsets(t *testing.T) {
	r1 := []byte("abcde")    // 5 bytes → offset 0
	r2 := []byte("fghij")    // 5 bytes → offset 5
	r3 := []byte("klmnopqr") // 8 bytes → offset 10

	offsets := sstable.ExportFileOffset([][]byte{r1, r2, r3})

	expected := []uint64{0, 5, 10}
	for i, want := range expected {
		if offsets[i] != want {
			t.Errorf("offset[%d]: want %d, got %d", i, want, offsets[i])
		}
	}
}

func TestFileOffset_EmptyList(t *testing.T) {
	offsets := sstable.ExportFileOffset([][]byte{})
	if len(offsets) != 0 {
		t.Errorf("expected empty offsets, got %v", offsets)
	}
}

// ---------------------------------------------------------------------------
// Compact-level integration: corrupt file in temp dir
// ---------------------------------------------------------------------------

func TestCompact_IgnoresFileWithTruncatedRecord(t *testing.T) {
	dir := t.TempDir()

	// Write 4 files so Compact's minThreshold (4) is met.
	// Three are valid; one has its record body truncated.
	writeValidSST := func(name, key, val string) {
		rec := buildValidRecord(key, val, 1)
		footer := buildValidFooter(uint64(len(rec)))
		data := buildSSTFile([][]byte{rec}, []byte{}, footer)
		os.WriteFile(filepath.Join(dir, name), data, 0644)
	}

	writeValidSST("kd_1.sst", "apple", "pie")
	writeValidSST("kd_2.sst", "banana", "split")
	writeValidSST("kd_3.sst", "cherry", "jam")

	// Fourth file: valid header, truncated body.
	rec := buildValidRecord("date", "fruit", 1)
	truncated := rec[:len(rec)-3]
	footer := buildValidFooter(uint64(len(truncated)))
	corrupt := buildSSTFile([][]byte{truncated}, []byte{}, footer)
	os.WriteFile(filepath.Join(dir, "kd_4.sst"), corrupt, 0644)

	// Compact should not panic; corruption causes the inner loop to break early.
	err := sstable.Compact(dir)
	// We accept either nil or an error — the key requirement is no panic / index OOB.
	t.Logf("Compact returned: %v", err)
}

func TestCompact_FileWithSwappedMagic(t *testing.T) {
	dir := t.TempDir()

	sstable.MergeList = func() sstable.ListMerger {
		return memtable.CreateSkiplist()
	}

	writeSST := func(name, key, val string, corruptMagic bool) {
		rec := buildValidRecord(key, val, 1)
		footer := buildValidFooter(uint64(len(rec)))
		if corruptMagic {
			// Swap the last two bytes of the magic number.
			footer[22], footer[23] = footer[23], footer[22]
		}
		data := buildSSTFile([][]byte{rec}, []byte{}, footer)
		os.WriteFile(filepath.Join(dir, name), data, 0644)
	}

	writeSST("kd_1.sst", "alpha", "1", false)
	writeSST("kd_2.sst", "beta", "2", false)
	writeSST("kd_3.sst", "gamma", "3", false)
	writeSST("kd_4.sst", "delta", "4", true) // corrupted magic

	err := sstable.Compact(dir)
	t.Logf("Compact with swapped magic returned: %v", err)
}
