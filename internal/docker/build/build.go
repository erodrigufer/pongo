// dockerBuild provides functions to build Docker images using the Docker SDK.
package dockerBuild

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

// Reference material used to program this functions:
// https://blog.loginradius.com/engineering/build-push-docker-images-golang/

// ImageBuild, uses the docker client specified to build an image out of the
// files that happen to be on the path specified by srcPath. The specified path
// should have a Dockerfile and all other files required to build the image.
// It names the created image as 'imageName'. If the specified path does not
// exist or does not contain a Dockerfile, ImageBuild will return an error.
func ImageBuild(dockerClient *client.Client, srcPath, imageName string) error {
	// Create a context with a timeout for the Docker daemon's action.
	timeOut, err := time.ParseDuration("5m")
	if err != nil {
		return fmt.Errorf("error: parsing time duration: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	// Every time a Docker image is built locally (even when using the Docker
	// CLI) all the files required to build the image are bundled in a tar file,
	// the tar file is then sent to the Docker daemon. The tar file with
	// Dockerfile and any other required files is the so-called 'build context'.
	tar, err := archive.TarWithOptions(srcPath, &archive.TarOptions{})
	if err != nil {
		return fmt.Errorf("error: could not create tar from path %s: %w", srcPath, err)
	}
	defer tar.Close()

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{imageName},
		Remove:     true,
	}
	res, err := dockerClient.ImageBuild(ctx, tar, opts)
	if err != nil {
		return fmt.Errorf("error: docker daemon could not build image from path %s: %w", srcPath, err)
	}
	defer res.Body.Close()

	// Scan the build text stream for errors.
	err = scanBuildStream(res.Body)
	if err != nil {
		return fmt.Errorf("an error was encountered during the docker image build: %w", err)
	}

	return nil
}

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

// scanBuildStream, scan the text stream output of the Docker daemon for the
// build being performed, if an error is encountered during the build this
// function returns the error.
func scanBuildStream(rd io.Reader) error {
	var lastLine string

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		lastLine = scanner.Text()
		// Uncomment the following line if you want to get a line-by-line
		// stream output of the Docker daemon building the image.
		// fmt.Println(scanner.Text())
	}

	errLine := &ErrorLine{}
	// If the last line does not contain an error, then the unmarshalling
	// will not populate the fields of the errLine struct.
	json.Unmarshal([]byte(lastLine), errLine)
	// An error was encountered.
	if errLine.Error != "" {
		return errors.New(errLine.Error)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error: scan of Docker build stream encountered an error: %w", err)
	}

	return nil
}
