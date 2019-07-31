package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRefLocal(t *testing.T) {
	locals := []Ref{
		MustCreateRef("#/location"),
		MustCreateRef("#/location/item"),
	}
	remotes := []Ref{
		MustCreateRef("file.yaml#//a/b/c"),
		MustCreateRef("path/to/file/file.yaml#//a/b/c"),
		MustCreateRef("file://a/b/c"),
		MustCreateRef("http://example.com//a/b/c"),
	}

	for _, ref := range locals {
		assert.Truef(t, isRefLocal(ref), "expected ref: %s to be considered local", ref.String())
	}

	for _, ref := range remotes {
		assert.False(t, isRefLocal(ref), "expected ref: %s to be considered remote", ref.String())
	}
}
