package safefile

import (
	"path/filepath"
	"fmt"
	"io"
	"io/ioutil"
	"os"
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

func testInTempDir() error {
	data := "Hello, safe file"
	name := filepath.Join(os.TempDir(), fmt.Sprintf("safefile-test1-%x", time.Now().UnixNano()))
	defer os.Remove(name)
	f, err := Create(name, 0666)
	if err != nil {
		return err
	}
	if name != f.OrigName() {
		return fmt.Errorf("name %q differs from OrigName: %q", name, f.OrigName())
	}
	_, err = io.WriteString(f, data)
	if err != nil {
		f.Close()
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return ensureFileContains(name, data)
}

func TestFile(t *testing.T) {
	err := testInTempDir()
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func TestWriteFile(t *testing.T) {
	data := "Testing WriteFile"
	name := filepath.Join(os.TempDir(), fmt.Sprintf("safefile-test2-%x", time.Now().UnixNano()))
	err := WriteFile(name, []byte(data), 0666)
	if err != nil {
		t.Fatalf("%s", err)
	}
	err = ensureFileContains(name, data) 
	if err != nil {
		os.Remove(name)
		t.Fatalf("%s", err)
	}
	os.Remove(name)
}
