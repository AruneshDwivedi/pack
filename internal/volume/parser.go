// Package volume provides a minimal parser for Docker-style volume mount
// specifications (`source:target[:mode]`).
//
// It is a parse-only port of the Linux logic in
// github.com/docker/docker/volume/mounts, which is not part of the moby/moby
// split modules and pulls in heavy daemon-side dependencies (safepath, idtools,
// selinux, ...) that pack never exercises. Only the spec parsing and mode
// validation pack relies on is reproduced here. SELinux relabel modes
// (`:z`/`:Z`) are intentionally not handled (pre-existing gap in pack).
//
// See https://github.com/buildpacks/pack/issues/2470 for the follow-up to
// replace this with a maintained upstream implementation.
package volume

import (
	"fmt"
	"path"
	"strings"
)

// ParsedVolume is the subset of a parsed mount specification that pack needs.
type ParsedVolume struct {
	Source string
	Target string
	Mode   string
}

// Parser parses raw volume mount specifications.
type Parser interface {
	ParseMountRaw(raw string) (ParsedVolume, error)
}

// NewLinuxParser returns a Parser for Linux-container mount specifications.
// Pack only supports Linux containers, so this is the only parser variant.
func NewLinuxParser() Parser {
	return &linuxParser{}
}

type linuxParser struct{}

var rwModes = map[string]bool{
	"rw": true,
	"ro": true,
}

var propagationModes = map[string]bool{
	"private":  true,
	"rprivate": true,
	"slave":    true,
	"rslave":   true,
	"shared":   true,
	"rshared":  true,
}

var consistencyModes = map[string]bool{
	"consistent": true,
	"cached":     true,
	"delegated":  true,
}

// copyModes is the set of recognized copy-mode tokens. Pack does not act on
// the copy mode; it only needs the token to be considered valid.
var copyModes = map[string]bool{
	"nocopy": true,
}

func validMountMode(mode string) bool {
	if mode == "" {
		return true
	}

	var rw, prop, copyc, cons int
	for _, o := range strings.Split(mode, ",") {
		switch {
		case rwModes[o]:
			rw++
		case propagationModes[o]:
			prop++
		case copyModes[o]:
			copyc++
		case consistencyModes[o]:
			cons++
		default:
			return false
		}
	}
	if rw > 1 || prop > 1 || copyc > 1 || cons > 1 {
		return false
	}
	return true
}

func validateTarget(target string) error {
	if target == "" {
		return fmt.Errorf("field Target must not be empty")
	}
	clean := path.Clean(strings.ReplaceAll(target, `\`, `/`))
	if clean == "/" {
		return fmt.Errorf("invalid specification: destination can't be '/'")
	}
	if !path.IsAbs(strings.ReplaceAll(target, `\`, `/`)) {
		return fmt.Errorf("invalid mount path: '%s' mount path must be absolute", target)
	}
	return nil
}

func (p *linuxParser) ParseMountRaw(raw string) (ParsedVolume, error) {
	arr := strings.SplitN(raw, ":", 4)
	if arr[0] == "" {
		return ParsedVolume{}, errInvalidSpec(raw)
	}

	var source, target, mode string
	switch len(arr) {
	case 1:
		target = arr[0]
	case 2:
		if validMountMode(arr[1]) {
			// Destination + Mode is not a valid volume - volumes
			// cannot include a mode. e.g. /foo:rw
			return ParsedVolume{}, errInvalidSpec(raw)
		}
		source = arr[0]
		target = arr[1]
	case 3:
		source = arr[0]
		target = arr[1]
		mode = arr[2]
	default:
		return ParsedVolume{}, errInvalidSpec(raw)
	}

	if !validMountMode(mode) {
		return ParsedVolume{}, fmt.Errorf("invalid mode: %v", mode)
	}

	if err := validateTarget(target); err != nil {
		return ParsedVolume{}, fmt.Errorf("%v: %v", errInvalidSpec(raw), err)
	}

	return ParsedVolume{Source: source, Target: target, Mode: mode}, nil
}

func errInvalidSpec(spec string) error {
	return fmt.Errorf("invalid volume specification: '%s'", spec)
}
