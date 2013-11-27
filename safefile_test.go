package safefile

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func ensureFileContains(name, data string) error {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	if string(b) != data {
		return fmt.Errorf("wrong data in file: expected %s, got %s", data, string(b))
	}
	return nil
}

func tempFileName(count int) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("safefile-test-%d-%x", count, time.Now().UnixNano()))
}

var testData = "Hello, this is a test data"

func testInTempDir() error {
	name := tempFileName(0)
	defer os.Remove(name)
	f, err := Create(name, 0666)
	if err != nil {
		return err
	}
	if name != f.OrigName() {
		f.Close()
		return fmt.Errorf("name %q differs from OrigName: %q", name, f.OrigName())
	}
	_, err = io.WriteString(f, testData)
	if err != nil {
		f.Close()
		return err
	}
	err = f.Commit()
	if err != nil {
		f.Close()
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return ensureFileContains(name, testData)
}

func TestFile(t *testing.T) {
	err := testInTempDir()
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func TestWriteFile(t *testing.T) {
	name := tempFileName(1)
	err := WriteFile(name, []byte(testData), 0666)
	if err != nil {
		t.Fatalf("%s", err)
	}
	err = ensureFileContains(name, testData)
	if err != nil {
		os.Remove(name)
		t.Fatalf("%s", err)
	}
	os.Remove(name)
}

func TestAbandon(t *testing.T) {
	name := tempFileName(3)
	f, err := Create(name, 0666)
	if err != nil {
		t.Fatalf("%s", err)
	}
	err = f.Close()
	if err != nil {
		t.Fatalf("Abandon failed: %s", err)
	}
	// Make sure temporary file doesn't exist.
	_, err = os.Stat(f.Name())
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("%s", err)
	}
}

func TestDoubleCommit(t *testing.T) {
	name := tempFileName(3)
	f, err := Create(name, 0666)
	if err != nil {
		t.Fatalf("%s", err)
	}
	err = f.Commit()
	if err != nil {
		os.Remove(name)
		t.Fatalf("First commit failed: %s", err)
	}
	err = f.Commit()
	if err != ErrAlreadyCommitted {
		os.Remove(name)
		t.Fatalf("Second commit didn't fail: %s", err)
	}
	err = f.Close()
	if err != nil {
		os.Remove(name)
		t.Fatalf("Close failed: %s", err)
	}
	os.Remove(name)
}
