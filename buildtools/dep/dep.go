// Package dep provides functions for working with the dep tool.
package dep

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/fossas/fossa-cli/buildtools"
	"github.com/fossas/fossa-cli/files"
	"github.com/fossas/fossa-cli/pkg"
)

// A Resolver contains both the Lockfile and Manifest information. Resolver implements Resolver.
type Resolver struct {
	Manifest Manifest
	Lockfile Lockfile
}

// Manifest contains the contents of a dep toml file. Manifest implements Resolver.
// Ignored is a result of toml unmarshalling. IgnoreMap is for easy search of ignored packages
type Manifest struct {
	Ignored []string
}

// A Lockfile contains the contents of a dep lockfile. Lockfiles are resolvers.
type Lockfile struct {
	Projects   []Project
	normalized map[string]pkg.Import // A normalized map of package import paths to revisions.
}

// A Project is a single imported repository within a dep project.
type Project struct {
	Name     string
	Packages []string
	Revision string
	Version  string
}

// Resolve determines if the Gopkg.toml dep file has any rules that apply
func (r Resolver) Resolve(importpath string) (pkg.Import, error) {
	if r.Manifest.IsIgnored(importpath) {
		return pkg.Import{}, buildtools.ErrPackageIsIgnored
	}

	return r.Lockfile.Resolve(importpath)
}

// IsIgnored finds packages to ignore while handling the possibility of a wildcard
func (m Manifest) IsIgnored(importpath string) bool {

	for _, ignoredPackage := range m.Ignored {
		if strings.HasSuffix(ignoredPackage, "*") {
			ignoredWildcard := ignoredPackage[:len(ignoredPackage)-1]
			if strings.HasPrefix(importpath, ignoredWildcard) {
				return true
			}
		}

		if ignoredPackage == importpath {
			return true
		}
	}

	return false
}

// Resolve returns the revision of an imported Go package contained within the
// lockfile. If the package is not found, buildtools.ErrNoRevisionForPackage is
// returned.
func (l Lockfile) Resolve(importpath string) (pkg.Import, error) {
	rev, ok := l.normalized[importpath]
	if !ok {
		return pkg.Import{}, buildtools.ErrNoRevisionForPackage
	}
	return rev, nil
}

// New constructs a golang.Resolver
func New(lockfilePath string, manifestPath string) (Resolver, error) {
	var err error

	resolver := Resolver{}
	resolver.Lockfile, err = ReadLockfile(lockfilePath)
	if err != nil {
		return Resolver{}, err
	}

	resolver.Manifest, err = ReadManifest(manifestPath)
	if err != nil {
		return Resolver{}, err
	}

	normalized := make(map[string]pkg.Import)
	for _, project := range resolver.Lockfile.Projects {
		for _, pk := range project.Packages {
			importpath := path.Join(project.Name, pk)
			normalized[importpath] = pkg.Import{
				Target: project.Version,
				Resolved: pkg.ID{
					Type:     pkg.Go,
					Name:     importpath,
					Revision: project.Revision,
					Location: "",
				},
			}
		}
	}

	resolver.Lockfile.normalized = normalized
	return resolver, nil
}

// ReadLockFile accepts the filepath of a lockfile which it parses into a lockfile object
func ReadLockfile(filepath string) (Lockfile, error) {
	var lockfile Lockfile

	err := files.ReadTOML(&lockfile, filepath)
	if err != nil {
		return Lockfile{}, errors.New(fmt.Sprintf("No lockfile Gopkg.lock found: %+v", err))
	}
	return lockfile, nil
}

// ReadManifest accepts the filepath of a manifest which it parses into a manifest object
func ReadManifest(filepath string) (Manifest, error) {
	var manifest Manifest
	err := files.ReadTOML(&manifest, filepath)
	if err != nil {
		return Manifest{}, errors.New(fmt.Sprintf("No manifest Gopkg.toml found: %+v", err))
	}

	return manifest, nil
}
