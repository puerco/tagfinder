package scanner

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/chainguard-dev/tagfinder/pkg/spdx"
	"github.com/nozzle/throttler"
)

type scannerImplementation interface {
	BuildFileList(string) ([]string, error)
	ScanPath(*Options, string) ([]spdx.Tag, error)
	ScanFile(*Options, string) ([]spdx.Tag, error)
	ParseLine(line string) *spdx.Tag
}

type defaultImplementation struct{}

func (di *defaultImplementation) BuildFileList(path string) ([]string, error) {
	fileList := []string{}

	// Scan the directory for files
	if err := fs.WalkDir(
		os.DirFS(path), ".", func(path string, entry fs.DirEntry, err error,
		) error {
			if err != nil {
				return err
			}
			if entry.Type().IsRegular() {
				fileList = append(fileList, path)
			}
			return nil
		}); err != nil {
		return nil, fmt.Errorf("scanning directory: %w", err)
	}
	return fileList, nil
}

func (di *defaultImplementation) ScanPath(opts *Options, path string) ([]spdx.Tag, error) {
	fmt.Printf("Scanning %s with %d threads, max %d lines", path, opts.Threads, opts.Lines)
	// Scan the directory to finda all files
	fileList, err := di.BuildFileList(path)
	if err != nil {
		return nil, fmt.Errorf("building file list: %w", err)
	}

	numThreads := 5
	if opts.Threads > 0 {
		numThreads = opts.Threads
	}

	mtx := sync.Mutex{}

	t := throttler.New(numThreads, len(fileList))
	tags := []spdx.Tag{}
	for _, filePath := range fileList {
		filePath := filePath
		go func(f string) {
			foundTags, err := di.ScanFile(opts, filepath.Join(path, filePath))
			if err != nil {
				t.Done(err)
				return
			}
			if len(foundTags) > 0 {
				mtx.Lock()
				tags = append(tags, foundTags...)
				mtx.Unlock()
			}
			t.Done(nil)
		}(filePath)
		t.Throttle()
	}

	if t.Err() != nil {
		return nil, fmt.Errorf("scanning path: %w", t.Err())
	}

	return tags, nil
}

// ScanFile searches n lines from a file returning al SPDX lines found.
func (di *defaultImplementation) ScanFile(opts *Options, path string) (tags []spdx.Tag, err error) {
	f, err := os.Open(path)
	if err != nil {
		return tags, fmt.Errorf("opening %s: %w", path, err)
	}

	// Scan the lines
	tags = []spdx.Tag{}

	fileScanner := bufio.NewScanner(f)
	lines := 0
	for fileScanner.Scan() {
		if tag := ParseLine(fileScanner.Text()); tag != nil {
			tags = append(tags, *tag)
		}
		lines++
		if lines != 0 && lines > opts.Lines {
			break
		}
	}

	return tags, nil
}

// ParseLine checks a line to see if it contains an SPDX tag
func (di *defaultImplementation) ParseLine(line string) *spdx.Tag {
	if res := tagRegExp.FindStringSubmatch(line); len(res) > 0 {
		return &spdx.Tag{
			Name:  res[1],
			Value: res[2],
		}
	}
	return nil
}
