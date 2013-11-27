// Copyright 2013 Dmitry Chestnykh. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package safefile implements safe "atomic" saving of files.
//
// Instead of truncating and overwriting the destination file, it creates a
// temporary file in the same directory, writes to it, and then renames the
// temporary file to the original name on close.
//
// Example:
//
//  f, err := safefile.Create("/home/ken/report.txt")
//  if err != nil {
//  	// ...
//  }
//  // Created temporary file /home/ken/133a7876287381fa-0.tmp
//
//  _, err = io.WriteString(f, "Hello world")
//  if err != nil {
//  	// ...
//  }
//  // Wrote "Hello world" to /home/ken/133a7876287381fa-0.tmp
//
//  err = f.Close()
//  if err != nil {
//      // ...
//      // Due to close error, temporary file removed.
//      // ...
//  }
//  // Renamed /home/ken/133a7876287381fa-0.tmp to /home/ken/report.txt
//
package safefile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type File struct {
	*os.File
	origName string
}

func makeTempName(origname string, counter int) (tempname string, err error) {
	origname = filepath.Clean(origname)
	if len(origname) == 0 || origname[len(origname)-1] == filepath.Separator {
		return "", os.ErrInvalid
	}
	return filepath.Join(filepath.Dir(origname), fmt.Sprintf("%x-%d.tmp", time.Now().UnixNano(), counter)), nil
}

// Create creates a file in the same directory as filename
func Create(filename string, perm os.FileMode) (*File, error) {
	counter := 0
	for {
		tempname, err := makeTempName(filename, counter)
		if err != nil {
			return nil, err
		}
		f, err := os.OpenFile(tempname, os.O_RDWR|os.O_CREATE, perm)
		if err != nil {
			if os.IsExist(err) {
				counter++
				continue
			}
			return nil, err
		}
		return &File{f, filename}, nil
	}
}

// OrigName returns the original filename given to Create.
func (f *File) OrigName() string {
	return f.origName
}

// CloseEx safely closes the file by syncing temporary file,
// closing it and renaming to the original file name.
//
// In case of error, the temporary file is closed, and if
// deleteOnError is true, it is also removed.
func (f *File) CloseEx(deleteOnError bool) error {
	// Sync to disk.
	err := f.Sync()
	if err != nil {
		f.File.Close() // ignore close error
		if deleteOnError {
			os.Remove(f.Name())
		}
		return err
	}
	// Close underlying os.File.
	err = f.File.Close()
	if err != nil {
		if deleteOnError {
			os.Remove(f.Name())
		}
		return err
	}
	// Rename.
	err = os.Rename(f.Name(), f.origName)
	if err != nil {
		if deleteOnError {
			os.Remove(f.Name())
		}
		return err
	}
	return nil
}

// Close calls CloseEx(true).
func (f *File) Close() error {
	return f.CloseEx(true)
}

// WriteFile is a safe analog of ioutil.WriteFile.
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	f, err := Create(filename, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err != nil {
		f.Close()
		return err
	}
	if err == nil && n < len(data) {
		f.Close()
		err = io.ErrShortWrite
		return err
	}
	return f.Close()
}
