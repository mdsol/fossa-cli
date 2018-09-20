package dep_test

import (
	"testing"

	"github.com/fossas/fossa-cli/buildtools/dep"
	"github.com/fossas/fossa-cli/pkg"
	"github.com/stretchr/testify/assert"
)

func TestIsIgnored(t *testing.T) {
	manifest := dep.Manifest{Ignored: []string{"apple", "orange*"}}

	// File listed in ignored list is ignored.
	valid := manifest.IsIgnored("apple")
	assert.Equal(t, valid, true)

	// Wildcard entry properly ignores its own package.
	valid = manifest.IsIgnored("orange")
	assert.Equal(t, valid, true)

	// Wildcard entry properly ignores other packages.
	valid = manifest.IsIgnored("orange/blood")
	assert.Equal(t, valid, true)

	// File not listed in ignored list is not ignored.
	valid = manifest.IsIgnored("apple/fuji")
	assert.Equal(t, valid, false)
}

func TestReadLockfile(t *testing.T) {
	// Reading a valid lockfile returns expected data.
	lockfile, err := dep.ReadLockfile("testdata/Gopkg.lock")
	expectedLockfile := dep.Lockfile{
		Projects: []dep.Project{
			dep.Project{
				Name:     "cat/fossa",
				Packages: []string{"."},
				Revision: "1",
				Version:  "v0.3.0",
			},
		},
		Normalized: map[string]pkg.Import{
			"cat/fossa": pkg.Import{
				Target: "v0.3.0",
				Resolved: pkg.ID{
					Type:     pkg.Go,
					Name:     "cat/fossa",
					Revision: "1",
					Location: "",
				},
			},
		},
	}

	assert.Equal(t, err, nil)
	assert.Equal(t, lockfile, expectedLockfile)

	// Reading an invalid lockfile returns an expected error.
	_, err = dep.ReadLockfile("NotAFile")
	assert.Error(t, err)
}

func TestReadManifest(t *testing.T) {
	// Reading a valid manifest returns expected data.
	manifest, err := dep.ReadManifest("testdata/Gopkg.toml")
	expectedManifest := dep.Manifest{
		Ignored: []string{
			"cat/puma",
			"cat/big/*",
		},
	}

	assert.Equal(t, err, nil)
	assert.Equal(t, manifest, expectedManifest)

	// Reading an invalid manifest returns an expected error.
	_, err = dep.ReadManifest("NotAFile")
	assert.Error(t, err)
}
