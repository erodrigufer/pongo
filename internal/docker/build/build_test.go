package dockerBuild

import (
	"testing"

	"github.com/docker/docker/client"
)

// TestImageBuild, tests that the ImageBuild function actually builds an image
// properly. It requires that a Docker daemon is present in the host running
// the test, otherwise the test will fail.
func TestImageBuild(t *testing.T) {
	// Start a Docker client.
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("error: could not start Docker client: %v", err)
	}

	imageName := "testbuildfunc"
	if err = ImageBuild(client, "./ImagesTest", imageName); err != nil {
		t.Errorf("error: failed to properly build the image '%s': %v", imageName, err)
	}

	if err = ImageBuild(client, "invalidPath", "invalidImage"); err == nil {
		t.Errorf("error: the Docker client should have returned an error, due to an invalid path to the Dockerfile")
	}

	//TODO: a cleanup function to erase the test image that was just built.
}
