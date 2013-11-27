# safefile

[![Build Status](https://travis-ci.org/dchest/safefile.png)](https://travis-ci.org/dchest/safefile)

Go package safefile implements safe "atomic" saving of files.

Instead of truncating and overwriting the destination file, it creates a
temporary file in the same directory, writes to it, and then renames the
temporary file to the original name when calling Commit.

## Installation

```
$ go get github.com/dchest/safefile
```

## Documentation
	
 <http://godoc.org/github.com/dchest/safefile>

## Example

```
f, err := safefile.Create("/home/ken/report.txt")
if err != nil {
	// ...
}
// Created temporary file /home/ken/133a7876287381fa-0.tmp
defer f.Close()

_, err = io.WriteString(f, "Hello world")
if err != nil {
	// ...
}
// Wrote "Hello world" to /home/ken/133a7876287381fa-0.tmp

err = f.Commit()
if err != nil {
    // ...
}
// Renamed /home/ken/133a7876287381fa-0.tmp to /home/ken/report.txt

```
