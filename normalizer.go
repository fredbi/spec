// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// normalizePaths returns the absolute path of refPath, given a ref path and a base absolute path.
//
// This support URI schemes http, https and file.
//
// 1. if refPath is absolute remote, return it
// 2. remove "file" scheme (file is implicit)
// 3. normalize url-unescaped
// 4. normalize all paths to slashes
// * if refPath is absolute file, return it slashed
// * if refPath is relative, join it with basePath keeping the scheme, hosts, and ports if exists
//
// base could be a directory or a fully qualified file path.
//
// We assume the "base" parameter has already been properly normalized.
func normalizePaths(refPath, base string) string {
	if strings.HasPrefix(refPath, "http") {
		// 1. if refPath is absolute remote, return it
		return refPath
	}

	if strings.HasPrefix(refPath, "file://") {
		// 2. remove "file" scheme (file is implicit)
		refPath = strings.TrimPrefix(refPath, "file://")
	}

	refURL := normalizeURL(refPath)

	if path.IsAbs(refURL.Path) {
		// this is an implied file ref: if the file path is absolute, just return it
		return refURL.String()
	}

	var baseURL *url.URL
	if strings.HasPrefix(base, "http") || strings.HasPrefix(base, "file://") {
		baseURL = mustURL(url.Parse(base))
	} else {
		baseURL = normalizeURL(base)
	}
	baseURL.Fragment = refURL.Fragment

	if strings.HasPrefix(refPath, "#") {
		return baseURL.String()
	}

	baseURL.Path = path.Dir(baseURL.Path)
	baseURL.Path = path.Join(baseURL.Path, refURL.Path)
	return baseURL.String()
}

// normalizeFileRef is the same a normalizePaths but with a Ref object as input
func normalizeFileRef(ref *Ref, base string) *Ref {
	if ref == nil || ref.String() == "" {
		return mustRefPtr(NewRef(base))
	}
	debugLog("normalizing ref: %s against base: %s", ref.String(), base)
	return mustRefPtr(NewRef(normalizePaths(ref.String(), base)))
}

// denormalizePaths returns to simplest notation on file $ref,
// i.e. strips the absolute path and sets a path relative to the base path.
//
// This is currently used when we rewrite ref after a circular ref has been detected
//
// relativeBase and originalRelativeBase are assumed to be already normalized.
//
// TODO(fredbi): change interface with pointers / retrn pointer that sucks
func denormalizeFileRef(ref *Ref, relativeBase, originalRelativeBase string) *Ref {
	debugLog("denormalizeFileRef for: %s", ref.String())

	if ref == nil || ref.String() == "" || isRefLocal(*ref) {
		return ref
	}

	relativeBaseURL := mustURL(url.Parse(relativeBase))
	relativeBaseURL.Fragment = ""

	if relativeBaseURL.IsAbs() && strings.HasPrefix(ref.String(), relativeBase) {
		// this should work for absolute URI (e.g. http://...): we have an exact match, just trim prefix
		r, _ := NewRef(strings.TrimPrefix(ref.String(), relativeBase))
		return &r
	}

	if relativeBaseURL.IsAbs() {
		// other absolute URL get unchanged (i.e. with a non-empty scheme)
		return ref
	}

	// for relative file URIs:
	originalRelativeBaseURL, _ := url.Parse(originalRelativeBase)
	originalRelativeBaseURL.Fragment = ""
	if strings.HasPrefix(ref.String(), originalRelativeBaseURL.String()) {
		// the resulting ref is in the expanded spec: return a local ref
		return mustRefPtr(NewRef(strings.TrimPrefix(ref.String(), originalRelativeBaseURL.String())))
	}

	// check if we may set a relative path, considering the original base path for this spec.
	// Example:
	//   spec is located at /mypath/spec.json
	//   my normalized ref points to: /mypath/item.json#/target
	//   expected result: item.json#/target
	parts := strings.Split(ref.String(), "#")
	relativePath, err := filepath.Rel(path.Dir(originalRelativeBaseURL.String()), parts[0])
	if err != nil {
		// there is no common ancestor (e.g. different drives on windows)
		// leaves the ref unchanged
		return ref
	}
	if len(parts) == 2 {
		relativePath += "#" + parts[1]
	}
	return mustRefPtr(NewRef(relativePath))
}

// normalizedAbsPath returns a normalized absolute path for cache
//
// * all URIs or paths are normalized to slashes.
// * on windows, drive letters are lower cased.
func normalizedAbsPath(pth string) string {
	if strings.HasPrefix(pth, "http") {
		// does not change fully qualified URI
		return pth
	}

	if strings.HasPrefix(pth, "file://") {
		return "file://" + normalizedAbsPath(strings.TrimPrefix(pth, "file://"))
	}

	if !filepath.IsAbs(pth) {
		wd, _ := os.Getwd()
		return filepath.ToSlash(filepath.Join(wd, pth))
	}

	return normalizeURL(pth).String()
}

func mustURL(u *url.URL, err error) *url.URL {
	if err != nil {
		msg := fmt.Sprintf("invalid URL: %v", err)
		panic(msg)
	}
	return u
}

func mustString(s string, err error) string {
	if err != nil {
		msg := fmt.Sprintf("invalid string: %v", err)
		panic(msg)
	}
	return s
}

func mustRefPtr(ref Ref, err error) *Ref {
	if err != nil {
		msg := fmt.Sprintf("invalid ref: %v", err)
		panic(msg)
	}
	return &ref
}

// normalizeURL gives a common way to represent URLs without scheme internally, in particular
// when it comes to represent windows paths.
func normalizeURL(ref string) *url.URL {
	if runtime.GOOS == "windows" {
		// on windows, URIs starting with absolute file path are actually invalid:
		// drive letter is parsed as scheme, and path is actually rendered in the opaque section of the URL.
		// Rewrite the URL with correct path
		u := mustURL(url.Parse(ref))
		if len(u.Scheme) > 0 {
			u.Path = "/" + u.Scheme + ":" + filepath.ToSlash(u.Path) // turns cases like: C:/a/b/c into /C:/a/b//c
			u.Scheme = ""
			return u
		}
	}
	return mustURL(url.Parse(ref))
}
