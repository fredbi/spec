package spec

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getCwd() string {
	cwd, _ := os.Getwd()
	return cwd
}

// tests that paths are normalized correctly
func TestNormalizePaths(t *testing.T) {
	type testNormalizePathsTestCases []struct {
		refPath   string
		base      string
		expOutput string
	}
	testCases := func() testNormalizePathsTestCases {
		testCases := testNormalizePathsTestCases{
			{
				// http basePath, absolute refPath
				refPath:   "http://www.anotherexample.com/another/base/path/swagger.json#/definitions/Pet",
				base:      "http://www.example.com/base/path/swagger.json",
				expOutput: "http://www.anotherexample.com/another/base/path/swagger.json#/definitions/Pet",
			},
			{
				// http basePath, relative refPath
				refPath:   "another/base/path/swagger.json#/definitions/Pet",
				base:      "http://www.example.com/base/path/swagger.json",
				expOutput: "http://www.example.com/base/path/another/base/path/swagger.json#/definitions/Pet",
			},
			{
				refPath:   "#/definitions/Pet",
				base:      "http://www.example.com/base/path/swagger.json",
				expOutput: "http://www.example.com/base/path/swagger.json#/definitions/Pet",
			},
		}
		switch runtime.GOOS {
		case "windows":
			testCases = append(testCases, testNormalizePathsTestCases{
				{
					refPath:   `file://C:\another\base\path\swagger.json#/definitions/Pet`,
					base:      "http://www.example.com/base/path/swagger.json",
					expOutput: "/c:/another/base/path/swagger.json#/definitions/Pet",
				},
				{
					// file basePath, absolute refPath
					refPath:   `C:\another\base\path\swagger.json#/definitions/Pet`,
					base:      "http://www.example.com/base/path/swagger.json",
					expOutput: "/c:/another/base/path/swagger.json#/definitions/Pet",
				},
				{
					// file basePath, absolute refPath, no fragment
					refPath:   `c:\another\base\path.json`,
					base:      `c:\base\path.json`,
					expOutput: `/c:/another/base/path.json`,
				},
				{
					// file basePath, absolute refPath, no fragment
					refPath:   `C:\another\base\path.json`,
					base:      `c:\base\path.json`,
					expOutput: `/c:/another/base/path.json`,
				},
				{
					// file basePath, absolute refPath
					refPath:   `c:\another\base\path.json#/definitions/Pet`,
					base:      `c:\base/path.json`,
					expOutput: `/c:/another/base/path.json#/definitions/Pet`,
				},
				{
					// file basePath, relative refPath
					refPath:   `another\base\path.json#/definitions/Pet`,
					base:      `c:\base\path.json`,
					expOutput: `c:/base/another/base/path.json#/definitions/Pet`,
				},
				{
					// file scheme, backslashed
					refPath:   `another\base\path.json#/definitions/Pet`,
					base:      `file://c:/base/path.json`,
					expOutput: `/c:/base/another/base/path.json#/definitions/Pet`,
				},
				{
					// absolute refPath with file scheme, slashed
					refPath:   `file:///C:/another/base/path/swagger.json#/definitions/Pet`,
					base:      `http://www.example.com/base/path/swagger.json`,
					expOutput: `/c:/another/base/path/swagger.json#/definitions/Pet`,
				},
				{
					// absolute refPath with file scheme, backslashed
					refPath:   `file://\another/base/path/swagger.json#/definitions/Pet`,
					base:      `http://www.example.com/base/path/swagger.json`,
					expOutput: `/another/base/path/swagger.json#/definitions/Pet`,
				},
				{
					refPath:   "#/definitions/Pet",
					base:      `C:\a\b.json`,
					expOutput: "/c:/a/b.json#/definitions/Pet",
				},
			}...)
		default:
			// linux case
			testCases = append(testCases, testNormalizePathsTestCases{
				{
					// file basePath, absolute refPath
					refPath:   "file:///another/base/path/swagger.json#/definitions/Pet",
					base:      "http://www.example.com/base/path/swagger.json",
					expOutput: "/another/base/path/swagger.json#/definitions/Pet",
				},
				{
					// file basePath, absolute refPath
					refPath:   "/another/base/path/swagger.json#/definitions/Pet",
					base:      "http://www.example.com/base/path/swagger.json",
					expOutput: "/another/base/path/swagger.json#/definitions/Pet",
				},
				{
					// file basePath, absolute refPath, no fragment
					refPath:   "/another/base/path.json",
					base:      "/base/path.json",
					expOutput: "/another/base/path.json",
				},
				{
					// file basePath, absolute refPath
					refPath:   "/another/base/path.json#/definitions/Pet",
					base:      "/base/path.json",
					expOutput: "/another/base/path.json#/definitions/Pet",
				},
				{
					// file basePath, relative refPath
					refPath:   "another/base/path.json#/definitions/Pet",
					base:      "/base/path.json",
					expOutput: "/base/another/base/path.json#/definitions/Pet",
				},
				{
					// file scheme
					refPath:   `another/base/path.json#/definitions/Pet`,
					base:      `file:///base/path.json`, // formally non-legit, but common
					expOutput: `file:///base/another/base/path.json#/definitions/Pet`,
				},
				{
					// absolute refPath
					refPath:   `/another/base/path/swagger.json#/definitions/Pet`,
					base:      `http://www.example.com/base/path/swagger.json`,
					expOutput: `/another/base/path/swagger.json#/definitions/Pet`,
				},
				{
					// absolute refPath with file scheme
					refPath:   `file:///another/base/path/swagger.json#/definitions/Pet`,
					base:      `http://www.example.com/base/path/swagger.json`,
					expOutput: `/another/base/path/swagger.json#/definitions/Pet`,
				},
				{
					refPath:   "#/definitions/Pet",
					base:      `/a/b.json`,
					expOutput: "/a/b.json#/definitions/Pet",
				},
			}...)
		}
		return testCases
	}()

	for _, tcase := range testCases {
		out := normalizePaths(tcase.refPath, normalizedAbsPath(tcase.base))
		assert.Equalf(t, tcase.expOutput, out, "expected normalizePath(%q, %q)=%q, but got %q", tcase.refPath, tcase.base, tcase.expOutput, out)
	}
}

func TestDenormalizeFileRef(t *testing.T) {
	type testDenormalizeTestCases []struct {
		refPath      string
		base         string
		originalBase string
		expOutput    string
	}
	testCases := func() testDenormalizeTestCases {
		testCases := testDenormalizeTestCases{
			{
				refPath:      "http://example.com/a/b/c.json#/definitions/items",
				base:         "http://example.com/a/b/c.json",
				originalBase: "http://example.com/root.json",
				expOutput:    "#/definitions/items",
			},
			{
				refPath:      "http://example.com/a.json#/definitions/items",
				base:         "http://example.com/a.json",
				originalBase: "http://example.com/a.json",
				expOutput:    "#/definitions/items",
			},
			{
				refPath:      "",
				base:         "http://example.com/a/b/c.json",
				originalBase: "http://example.com/root.json",
				expOutput:    "",
			},
			{
				refPath:      "#/someplace/here",
				base:         "http://example.com/a/b/c.json",
				originalBase: "http://example.com/root.json",
				expOutput:    "#/someplace/here",
			},
			{
				refPath:      "http://example1.com/a/b/c.json#/someplace/here",
				base:         "http://example1.com/a/b/c.json",
				originalBase: "http://example2.com/root.json",
				expOutput:    "#/someplace/here",
			},
			{
				refPath:      "file:///a/b/c.json#/someplace/here",
				base:         "file:///a/b/c.json",
				originalBase: "file:///x/root.json",
				expOutput:    "#/someplace/here",
			},
		}
		return testCases
	}()
	for _, tcase := range testCases {
		r := MustCreateRef(tcase.refPath)
		out := denormalizeFileRef(&r, tcase.base, tcase.originalBase)
		require.Equalf(t, tcase.expOutput, out.String(), "expected denormalizeFileRef(%q, %q, %q)=%q, but got %q", tcase.refPath, tcase.base, tcase.originalBase, tcase.expOutput, out)
	}
}

func TestNormalizedAbsPath(t *testing.T) {
	fixtures := []struct{ path, expected string }{
		{`file:///a/b/c`, `file://` + filepath.ToSlash(testAbsPath(`/a/b/c`))},
		{`file://a/b/c`, `file://` + filepath.ToSlash(testAbsPath(`a/b/c`))},
		{`https://example.com/a/b/c`, `https://example.com/a/b/c`},
		{`http://example.com/a/b/c`, `http://example.com/a/b/c`},
	}
	switch runtime.GOOS {
	case "windows":
		fixtures = append(fixtures, []struct{ path, expected string }{
			{`file://C:\a\b\c`, `file://c:/a/b/c`}, // note that under this form, the drive letter is not lower cased
			{`C:\a\b\c`, `c:/a/b/c`},
			{`\a\b\c`, filepath.ToSlash(testAbsPath(`\a\b\c`))},
			{`a\b\c`, filepath.ToSlash(testAbsPath(`a\b\c`))},
			{`.\a\b\c`, filepath.ToSlash(testAbsPath(`.\a\b\c`))},
		}...)
	default:
		fixtures = append(fixtures, []struct{ path, expected string }{
			{`/a/b/c`, `/a/b/c`},
			{`a/b/c`, testAbsPath(`a/b/c`)},
			{`./a/b/c`, testAbsPath(`./a/b/c`)},
			{`file:///a/b/c`, `file:///a/b/c`},
		}...)
	}

	for _, fixture := range fixtures {
		assert.Equal(t, fixture.expected, normalizedAbsPath(fixture.path))
	}
}
