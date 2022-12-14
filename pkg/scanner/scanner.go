package scanner

import (
	"regexp"

	"github.com/chainguard-dev/tagfinder/pkg/spdx"
)

type Scanner struct {
	impl    scannerImplementation
	Options Options
}

func New(opts ...FnOption) *Scanner {
	return &Scanner{
		impl:    &defaultImplementation{},
		Options: *buildOptions(opts),
	}
}

type ScanResults struct {
	Tags         map[string][]spdx.Tag
	IgnoredFiles []string
}

var tagRegExp regexp.Regexp

func init() {
	tagRegExp = *regexp.MustCompile(`^\s*[#/*]+\s*SPDX-(\S*):\s*(.*)`)
}

// ScanPath scans a path
func (s *Scanner) ScanPath(path string) ([]spdx.Tag, error) {
	return s.impl.ScanPath(&s.Options, path)
}

// ScanPath scans a directory
func ScanPath(path string, passedOpts ...FnOption) ([]spdx.Tag, error) {
	di := defaultImplementation{}
	return di.ScanPath(buildOptions(passedOpts), path)
}

// ParseLine checks a line to see if it contains an SPDX tag
func ParseLine(line string) *spdx.Tag {
	di := defaultImplementation{}
	return di.ParseLine(line)
}

// ScanFile reads a file and returns all tags found
func ScanFile(path string, passedOpts ...FnOption) (tags []spdx.Tag, err error) {
	di := defaultImplementation{}
	return di.ScanFile(buildOptions(passedOpts), path)
}
