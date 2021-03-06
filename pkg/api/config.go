package api

import (
	"strings"
)

const DefaultArtifacts = "/tmp/artifacts"

// Default sets default values after loading but before validation
func (config *ReleaseBuildConfiguration) Default() {
	defArtifacts := func(p *string) {
		if *p == "" {
			*p = DefaultArtifacts
		}
	}
	defLeases := func(l []StepLease) {
		for i := range l {
			if l[i].Count == 0 {
				l[i].Count = 1
			}
		}
	}
	def := func(s *LiteralTestStep) {
		defArtifacts(&s.ArtifactDir)
		defLeases(s.Leases)
	}
	defTest := func(t *TestStepConfiguration) {
		defArtifacts(&t.ArtifactDir)
		if s := t.MultiStageTestConfigurationLiteral; s != nil {
			defLeases(s.Leases)
			for i := range s.Pre {
				def(&s.Pre[i])
			}
			for i := range s.Test {
				def(&s.Test[i])
			}
			for i := range s.Post {
				def(&s.Post[i])
			}
		}
	}
	for _, step := range config.RawSteps {
		if test := step.TestStepConfiguration; test != nil {
			defTest(test)
		}
	}
	for _, test := range config.Tests {
		defTest(&test)
	}
}

// ImageStreamFor guesses at the ImageStream that will hold a tag.
// We use this to decipher the user's intent when they provide a
// naked tag in configuration; we support such behavior in order to
// allow users a simpler workflow for the most common cases, like
// referring to `pipeline:src`. If they refer to an ambiguous image,
// however, they will get bad behavior and will need to specify an
// ImageStream as well, for instance release-initial:installer.
// We also return whether the stream is explicit or inferred.
func (config *ReleaseBuildConfiguration) ImageStreamFor(image string) (string, bool) {
	if config.IsPipelineImage(image) || config.BuildsImage(image) {
		return PipelineImageStream, true
	} else {
		return StableImageStream, false
	}
}

// DependencyParts returns the imageStream and tag name from a user-provided
// reference to an image in the test namespace
func (config *ReleaseBuildConfiguration) DependencyParts(dependency StepDependency) (string, string, bool) {
	if !strings.Contains(dependency.Name, ":") {
		stream, explicit := config.ImageStreamFor(dependency.Name)
		return stream, dependency.Name, explicit
	} else {
		parts := strings.Split(dependency.Name, ":")
		return parts[0], parts[1], true
	}
}
