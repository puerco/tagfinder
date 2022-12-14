package scanner

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"sync"

	"github.com/nozzle/throttler"

	"github.com/chainguard-dev/tagfinder/pkg/spdx"
)

type Scanner struct {
	Options Options
}

type ScanResults struct {
	Tags         map[string][]spdx.Tag
	IgnoredFiles []string
}

var tagRegExp regexp.Regexp

func init() {
	tagRegExp = *regexp.MustCompile(`^\s*[#/*]+\s*SPDX-(\S*):\s*(.*)`)
}

func buildFileList(path string) ([]string, error) {
	fileList := []string{}

	// Scan the directory for files
	if err := fs.WalkDir(
		os.DirFS(path), ".", func(path string, entry fs.DirEntry, err error,
		) error {
			if entry.Type().IsRegular() {
				fileList = append(fileList, path)
			}
			return nil
		}); err != nil {
		return nil, fmt.Errorf("scanning directory: %w", err)
	}
	return fileList, nil
}

// ScanPath scans a path
func (s *Scanner) ScanPath(path string) ([]spdx.Tag, error) {
	return scanPathWithOptions(path, &s.Options)
}

// ScanPath scans a directory
func ScanPath(path string, passedOpts ...FnOption) ([]spdx.Tag, error) {
	opts := &Options{}
	for _, o := range passedOpts {
		o(opts)
	}
	return scanPathWithOptions(path, opts)
}

func scanPathWithOptions(path string, opts *Options) ([]spdx.Tag, error) {
	fmt.Printf("Scanning %s with %d threads, max %d lines", path, opts.Threads, opts.Lines)
	// Scan the directory to finda all files
	fileList, err := buildFileList(path)
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
			foundTags, err := scanFileWithOptions(filePath, opts)
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

// ParseLine checks a line to see if it contains an SPDX tag
func ParseLine(line string) *spdx.Tag {
	if res := tagRegExp.FindStringSubmatch(line); len(res) > 0 {
		return &spdx.Tag{
			Name:  res[1],
			Value: res[2],
		}
	}
	return nil
}

// scanFileWithOptions searches n lines from a file returning al SPDX lines found.
// When n is zero, the complete file is read
func scanFileWithOptions(path string, opts *Options) (tags []spdx.Tag, err error) {
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

// ScanFile reads a file and returns all tags found
func (s *Scanner) ScanFile(
	path string, passedOpts ...FnOption,
) (tags []spdx.Tag, err error) {
	opts := &Options{}
	for _, o := range passedOpts {
		o(opts)
	}
	return scanFileWithOptions(path, opts)
}
