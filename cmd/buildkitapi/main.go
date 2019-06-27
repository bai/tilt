package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/pkg/errors"

	"github.com/windmilleng/tilt/internal/build"
	"github.com/windmilleng/tilt/internal/dockerfile"
)

// A small utility for running Buildkit on the dockerfile
// in the current directory printing out all the buildkit api
// response protobufs.
func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	d, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	d.NegotiateAPIVersion(ctx)

	df, err := ioutil.ReadFile("Dockerfile")
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Dockerfile not found in current directory")
		}
		return err
	}

	archive, err := build.TarDfOnly(ctx, dockerfile.Dockerfile(df))
	if err != nil {
		return err
	}

	opts := types.ImageBuildOptions{}
	opts.Version = types.BuilderBuildKit
	opts.Dockerfile = string(df)

	response, err := d.ImageBuild(ctx, archive, opts)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	return readDockerOutput(ctx, response.Body)
}

func readDockerOutput(ctx context.Context, reader io.Reader) error {
	decoder := json.NewDecoder(reader)

	for decoder.More() {
		message := jsonmessage.JSONMessage{}
		err := decoder.Decode(&message)
		if err != nil {
			return errors.Wrap(err, "decoding docker output")
		}

		if messageIsFromBuildkit(message) {
			err := writeBuildkitStatus(message.Aux)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func writeBuildkitStatus(aux *json.RawMessage) error {
	var resp controlapi.StatusResponse
	var dt []byte
	if err := json.Unmarshal(*aux, &dt); err != nil {
		return err
	}
	if err := (&resp).Unmarshal(dt); err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(resp)
}

func messageIsFromBuildkit(msg jsonmessage.JSONMessage) bool {
	return msg.ID == "moby.buildkit.trace"
}