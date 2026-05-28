//go:build linux || windows

package client

import (
	"testing"

	h "github.com/buildpacks/pack/testhelpers"
)

func TestProcessVolumes(t *testing.T) {
	t.Run("normalizes bind, named, and modeless volumes", func(t *testing.T) {
		processed, warnings, err := processVolumes("linux", []string{
			"/host/src:/target:rw",
			"my-vol:/data",
			"/host/ro:/readonly:ro",
		})
		h.AssertNil(t, err)
		h.AssertEq(t, len(warnings), 0)
		h.AssertSliceContainsInOrder(t, processed,
			"/host/src:/target:rw",
			"my-vol:/data:ro",
			"/host/ro:/readonly:ro",
		)
	})

	t.Run("warns when mounting to a sensitive directory", func(t *testing.T) {
		_, warnings, err := processVolumes("linux", []string{"/host:/cnb/buildpacks"})
		h.AssertNil(t, err)
		h.AssertEq(t, len(warnings), 1)
		h.AssertContains(t, warnings[0], "sensitive directory")
	})

	t.Run("errors on an invalid spec", func(t *testing.T) {
		_, _, err := processVolumes("linux", []string{"/data:rw"})
		h.AssertError(t, err, "has invalid format")
	})
}
