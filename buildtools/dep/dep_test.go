package dep_test

import (
	"testing"

	"github.com/fossas/fossa-cli/buildtools"
	"github.com/fossas/fossa-cli/buildtools/dep"
	"github.com/fossas/fossa-cli/pkg"
	"github.com/stretchr/testify/assert"
)

func TestIsIgnored(t *testing.T) {
	manifest := dep.Manifest{Ignored: []string{"apple", "orange*"}}

	// Regular ignore
	valid := manifest.IsIgnored("apple")
	assert.Equal(t, valid, true)

	// Wildcard ignore
	valid = manifest.IsIgnored("orange")
	assert.Equal(t, valid, true)

	// Wildcard ignore other packages
	valid = manifest.IsIgnored("orange/blood")
	assert.Equal(t, valid, true)

	// Don't ignore file that isn't in the ignored list
	valid = manifest.IsIgnored("apple/fuji")
	assert.Equal(t, valid, false)
}

func TestResolverResolver(t *testing.T) {
	resolver := dep.Resolver{
		Manifest: dep.Manifest{
			Ignored: []string{"badpackage"},
		},
		Lockfile: dep.Lockfile{
			Normalized: map[string]pkg.Import{"goodpackage": pkg.Import{}},
		},
	}

	// Package that can be ignored returns a sentinel error
	_, err := resolver.Resolve("badpackage")
	assert.Equal(t, buildtools.ErrPackageIsIgnored, err)

	// Package that cannot be ignored is not
	_, err = resolver.Resolve("goodpackage")
	assert.Equal(t, err, nil)
}

func TestLockfileResolver(t *testing.T) {
	lockfile := dep.Lockfile{Normalized: map[string]pkg.Import{"goodpackage": pkg.Import{}}}

	// Package not included in lockfile returns a sentinel error
	_, err := lockfile.Resolve("badpackage")
	assert.Equal(t, buildtools.ErrNoRevisionForPackage, err)

	// Package included in Lockfile does not return an error
	_, err = lockfile.Resolve("goodpackage")
	assert.Equal(t, err, nil)
}
