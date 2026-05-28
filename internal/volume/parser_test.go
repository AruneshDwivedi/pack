package volume_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"github.com/buildpacks/pack/internal/volume"
	h "github.com/buildpacks/pack/testhelpers"
)

func TestVolumeParser(t *testing.T) {
	spec.Run(t, "VolumeParser", testVolumeParser, spec.Report(report.Terminal{}))
}

func testVolumeParser(t *testing.T, when spec.G, it spec.S) {
	var parser volume.Parser

	it.Before(func() {
		parser = volume.NewLinuxParser()
	})

	when("valid specs", func() {
		it("parses a bind mount with mode", func() {
			v, err := parser.ParseMountRaw("/host/src:/cnb/target:rw")
			h.AssertNil(t, err)
			h.AssertEq(t, v.Source, "/host/src")
			h.AssertEq(t, v.Target, "/cnb/target")
			h.AssertEq(t, v.Mode, "rw")
		})

		it("parses a bind mount without mode", func() {
			v, err := parser.ParseMountRaw("/host/src:/workspace/data")
			h.AssertNil(t, err)
			h.AssertEq(t, v.Source, "/host/src")
			h.AssertEq(t, v.Target, "/workspace/data")
			h.AssertEq(t, v.Mode, "")
		})

		it("parses a named volume", func() {
			v, err := parser.ParseMountRaw("my-volume:/data:ro")
			h.AssertNil(t, err)
			h.AssertEq(t, v.Source, "my-volume")
			h.AssertEq(t, v.Target, "/data")
			h.AssertEq(t, v.Mode, "ro")
		})

		it("parses a destination-only spec", func() {
			v, err := parser.ParseMountRaw("/data")
			h.AssertNil(t, err)
			h.AssertEq(t, v.Source, "")
			h.AssertEq(t, v.Target, "/data")
		})

		it("accepts a propagation mode", func() {
			v, err := parser.ParseMountRaw("/host:/data:rslave")
			h.AssertNil(t, err)
			h.AssertEq(t, v.Mode, "rslave")
		})
	})

	when("invalid specs", func() {
		it("rejects an empty source", func() {
			_, err := parser.ParseMountRaw(":/data")
			h.AssertError(t, err, "invalid volume specification")
		})

		it("rejects destination + mode", func() {
			_, err := parser.ParseMountRaw("/data:rw")
			h.AssertError(t, err, "invalid volume specification")
		})

		it("rejects an unknown mode", func() {
			_, err := parser.ParseMountRaw("/host:/data:bogus")
			h.AssertError(t, err, "invalid mode: bogus")
		})

		it("rejects a non-absolute target", func() {
			_, err := parser.ParseMountRaw("/host:relative/target")
			h.AssertError(t, err, "mount path must be absolute")
		})

		it("rejects '/' as the target", func() {
			_, err := parser.ParseMountRaw("/host:/")
			h.AssertError(t, err, "destination can't be '/'")
		})

		it("rejects too many segments", func() {
			_, err := parser.ParseMountRaw("a:b:c:d")
			h.AssertError(t, err, "invalid volume specification")
		})
	})
}
