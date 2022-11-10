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

func ImageBuild(dockerClient *client.Client, srcPath string) error {
	timeOut, err := time.ParseDuration("120s")
	if err != nil {
		return fmt.Errorf("error: parsing time duration: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	tar, err := archive.TarWithOptions(srcPath, &archive.TarOptions{})
	if err != nil {
		return fmt.Errorf("error: could not create tar from path %s: %w", srcPath, err)
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{"ctfsmd"},
		Remove:     true,
	}
	res, err := dockerClient.ImageBuild(ctx, tar, opts)
	if err != nil {
		return fmt.Errorf("error: docker daemon could not build image from path %s: %w", srcPath, err)
	}
	defer res.Body.Close()

	err = print(res.Body)
	if err != nil {
		return err
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

func print(rd io.Reader) error {
	var lastLine string

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		lastLine = scanner.Text()
		fmt.Println(scanner.Text())
	}

	errLine := &ErrorLine{}
	json.Unmarshal([]byte(lastLine), errLine)
	if errLine.Error != "" {
		return errors.New(errLine.Error)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
