// Package dep provides functions for working with the dep tool.
package dep

import (
	"fmt"
	"path"
	"strings"

	"github.com/fossas/fossa-cli/buildtools"
	"github.com/fossas/fossa-cli/files"
	"github.com/fossas/fossa-cli/pkg"
)

// A Resolver contains both the Lockfile and Manifest information. Resolver implements golang.Resolver.
type Resolver struct {
	Manifest Manifest
	Lockfile Lockfile
}

// Manifest contains the ignored packages in a dep toml file. Manifest implements golang.Resolver.
type Manifest struct {
	Ignored []string
}

// A Lockfile contains the Projects in a dep lockfile and a corresponding map for retrieving project data.
type Lockfile struct {
	Projects   []Project
	Normalized map[string]pkg.Import
}

// A Project is a single imported repository within a dep project.
type Project struct {
	Name     string
	Packages []string
	Revision string
	Version  string
}

// Resolve returns the revision of an imported Go package contained within the
// lockfile and checks to see if it should be ignored. If the package cannot be
// ignored and is not found, buildtools.ErrNoRevisionForPackage is returned.
func (r Resolver) Resolve(importpath string) (pkg.Import, error) {
	if r.Manifest.IsIgnored(importpath) {
		return pkg.Import{}, buildtools.ErrPackageIsIgnored
	}

	revision, ok := r.Lockfile.Normalized[importpath]
	if !ok {
		return pkg.Import{}, buildtools.ErrNoRevisionForPackage
	}

	return revision, nil
}

// IsIgnored checks if a Go package can be ignored according to a dep manifest.
func (m Manifest) IsIgnored(importpath string) bool {
	for _, ignoredPackage := range m.Ignored {
		if strings.HasSuffix(ignoredPackage, "*") {
			ignoredPrefix := ignoredPackage[:len(ignoredPackage)-1]
			if strings.HasPrefix(importpath, ignoredPrefix) {
				return true
			}
		}

		if ignoredPackage == importpath {
			return true
		}
	}

	return false
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

	return resolver, nil
}

// ReadLockFile creates and returns a Lockfile object using the provided filepath
func ReadLockfile(filepath string) (Lockfile, error) {
	var lockfile Lockfile

	err := files.ReadTOML(&lockfile, filepath)
	fmt.Println(err)
	if err != nil {
		return Lockfile{}, fmt.Errorf("No lockfile Gopkg.lock found: %+v", err)
	}

	normalized := make(map[string]pkg.Import)
	for _, project := range lockfile.Projects {
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

	lockfile.Normalized = normalized
	return lockfile, nil
}

// ReadManifest crestes and returns a Manifest object using the provided filepath
func ReadManifest(filepath string) (Manifest, error) {
	var manifest Manifest
	err := files.ReadTOML(&manifest, filepath)
	if err != nil {
		return Manifest{}, fmt.Errorf("No manifest Gopkg.toml found: %+v", err)
	}

	return manifest, nil
}
