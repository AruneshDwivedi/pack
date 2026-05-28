//go:build linux || windows

package client

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/buildpacks/pack/internal/style"
	"github.com/buildpacks/pack/internal/volume"
)

func processVolumes(imgOS string, volumes []string) (processed []string, warnings []string, err error) {
	// Pack only supports Linux containers, so mount specs are always parsed
	// with Linux semantics regardless of the host OS.
	parser := volume.NewLinuxParser()
	for _, v := range volumes {
		parsed, err := parser.ParseMountRaw(v)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "platform volume %q has invalid format", v)
		}

		sensitiveDirs := []string{"/cnb", "/layers", "/workspace"}
		if imgOS == "windows" {
			sensitiveDirs = []string{`c:/cnb`, `c:\cnb`, `c:/layers`, `c:\layers`, `c:/workspace`, `c:\workspace`}
		}
		for _, p := range sensitiveDirs {
			if strings.HasPrefix(strings.ToLower(parsed.Target), p) {
				warnings = append(warnings, fmt.Sprintf("Mounting to a sensitive directory %s", style.Symbol(parsed.Target)))
			}
		}

		processed = append(processed, fmt.Sprintf("%s:%s:%s", parsed.Source, parsed.Target, processMode(parsed.Mode)))
	}
	return processed, warnings, nil
}

func processMode(mode string) string {
	if mode == "" {
		return "ro"
	}

	return mode
}
