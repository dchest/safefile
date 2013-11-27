// Copyright 2013 Dmitry Chestnykh. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package safefile implements safe "atomic" saving of files.
//
// Instead of truncating and overwriting the destination file, it creates a
// temporary file in the same directory, writes to it, and then renames the
// temporary file to the original name when calling Commit.
//
// Example:
//
//  f, err := safefile.Create("/home/ken/report.txt")
//  if err != nil {
//  	// ...
//  }
//  // Created temporary file /home/ken/133a7876287381fa-0.tmp
//
//  defer f.Close()
//
//  _, err = io.WriteString(f, "Hello world")
//  if err != nil {
//  	// ...
//  }
//  // Wrote "Hello world" to /home/ken/133a7876287381fa-0.tmp
//
//  err = f.Commit()
//  if err != nil {
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
	origName  string
	closeFunc func(*File) error
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
		return &File{
			File:      f,
			origName:  filename,
			closeFunc: closeUncommitted,
		}, nil
	}
}

// OrigName returns the original filename given to Create.
func (f *File) OrigName() string {
	return f.origName
}

// Close closes temporary file and removes it.
// If the file has been committed, Close is noop.
func (f *File) Close() error {
	return f.closeFunc(f)
}

func closeUncommitted(f *File) error {
	err0 := f.File.Close()
	err1 := os.Remove(f.Name())
	f.closeFunc = closeAgainError
	if err0 != nil {
		return err0
	}
	return err1
}

func closeAfterFailedRename(f *File) error {
	// just remove temporary file.
	f.closeFunc = closeAgainError
	return os.Remove(f.Name())
}

func closeCommitted(f *File) error {
	// noop
	return nil
}

func closeAgainError(f *File) error {
	return os.ErrInvalid
}

// Commit safely closes the file by syncing temporary file,
// closing it and renaming to the original file name.
//
// In case of success, the temporary file is closed and
// no longer exists on disk. It is safe to call Close on
// after Commit: the operation will do nothing.
//
// In case of error, the temporary file is still opened
// and exists on disk; it must be closed by callers by
// calling Close or by trying to commit again.
func (f *File) Commit() error {
	// Sync to disk.
	err := f.Sync()
	if err != nil {
		return err
	}
	// Close underlying os.File.
	err = f.File.Close()
	if err != nil {
		return err
	}
	// Rename.
	err = os.Rename(f.Name(), f.origName)
	if err != nil {
		f.closeFunc = closeAfterFailedRename
		return err
	}
	f.closeFunc = closeCommitted
	return nil
}

// WriteFile is a safe analog of ioutil.WriteFile.
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	f, err := Create(filename, perm)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := f.Write(data)
	if err != nil {
		return err
	}
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
		return err
	}
	return f.Commit()
}
