package scanner

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanPath(t *testing.T) {
	s := Scanner{}
	_, err := s.ScanPath("./test_data/example.c")
	require.Error(t, err)
}

func TestScanFileWithOptions(t *testing.T) {
	di := defaultImplementation{}
	tags, err := di.ScanFile(&Options{Lines: 100}, "./test_data/example.c")
	require.NoError(t, err)
	require.Len(t, tags, 3)

	tags, err = di.ScanFile(&Options{Lines: 3}, "./test_data/example.c")
	require.NoError(t, err)
	require.Len(t, tags, 2)
}

func TestParseLine(t *testing.T) {
	di := defaultImplementation{}
	for _, tc := range []struct {
		shouldNil bool
		tag       string
		name      string
		value     string
	}{
		{
			// Normal
			false, "// SPDX-FileType: DOCUMENTATION", "FileType", "DOCUMENTATION",
		},
		{
			// With spaces in the front
			false, "  // SPDX-FileType: DOCUMENTATION", "FileType", "DOCUMENTATION",
		},
		{
			// Hash comment
			false, "# SPDX-FileType: DOCUMENTATION", "FileType", "DOCUMENTATION",
		},
		{
			// Multiword value
			false, "# SPDX-FileCopyrightText: 2019 Jane Doe <jane@example.com>", "FileCopyrightText", "2019 Jane Doe <jane@example.com>",
		},
	} {
		res := di.ParseLine(tc.tag)
		if tc.shouldNil {
			require.Nil(t, res)
		} else {
			require.NotNil(t, res)
			require.Equal(t, tc.name, res.Name)
			require.Equal(t, tc.value, res.Value)
		}
	}
}
