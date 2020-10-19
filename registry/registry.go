package registry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"

	"cloud.google.com/go/storage"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/docs"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

type RegistryConfig struct {
	Source string `hcl:"source"`
	Name   string `hcl:"name"`
	Bucket string `hcl:"bucket"`
}

type Registry struct {
	config RegistryConfig
}

// Documentation implements Documented.
func (r *Registry) Documentation() (*docs.Documentation, error) {
	doc, err := docs.New(docs.FromConfig(&RegistryConfig{}))
	if err != nil {
		return nil, err
	}

	doc.Description("Upload build artifcats to Google Cloud Storage")

	doc.Example(`
build {
	use "archive" {
		sources = ["./"]
		output_name = "server.zip"
		overwrite_existing = true
	}

		registry {
		use "gcs" {
			source = "server.zip"
			name = "${gitrefpretty()}.zip"
			bucket = "staging.gcp-project-name.appspot.com"
		}
		}
	}
`)

	doc.Output("gcs.Artifact")

	_ = doc.SetField("source", "The build artifact to upload to GCS", docs.Summary())

	_ = doc.SetField("name", "the name of the object to create on GCS", docs.Summary())

	_ = doc.SetField("bucket", "the name of the GCS Bucket", docs.Summary())

	return doc, nil
}

// Config implements Configurable.
func (r *Registry) Config() (interface{}, error) {
	return &r.config, nil
}

// ConfigSet implements ConfigurableNotify.
func (r *Registry) ConfigSet(config interface{}) error {
	c, ok := config.(*RegistryConfig)
	if !ok {
		// The Waypoint SDK should ensure this never gets hit
		return fmt.Errorf("Expected *RegisterConfig as parameter")
	}

	// validate the config
	if c.Source == "" {
		return errors.New("Source artifcat should not be empty")
	}

	if c.Name == "" {
		return errors.New("Name of the object should not be empty")
	}

	if c.Bucket == "" {
		return errors.New("Bucket should not be empty")
	}

	return nil
}

// PushFunc implements Registry.
func (r *Registry) PushFunc() interface{} {
	// return a function which will be called by Waypoint
	return r.push
}

// A PushFunc does not have a strict signature, you can define the parameters
// you need based on the Available parameters that the Waypoint SDK provides.
// Waypoint will automatically inject parameters as specified
// in the signature at run time.
//
// Available input parameters:
// - context.Context
// - *component.Source
// - *component.JobInfo
// - *component.DeploymentConfig
// - *datadir.Project
// - *datadir.App
// - *datadir.Component
// - hclog.Logger
// - terminal.UI
// - *component.LabelSet
//
// In addition to default input parameters the builder.Binary from the Build step
// can also be injected.
//
// The output parameters for PushFunc must be a Struct which can
// be serialzied to Protocol Buffers binary format and an error.
// This Output Value will be made available for other functions
// as an input parameter.
// If an error is returned, Waypoint stops the execution flow and
// returns an error to the user.
func (r *Registry) push(ctx context.Context, source *component.Source, ui terminal.UI) (*Artifact, error) {
	u := ui.Status()
	defer u.Close()

	u.Update("Pushing artifact to registry: " + r.config.Name)

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	object := client.Bucket(r.config.Bucket).Object(r.config.Name)
	wc := object.NewWriter(ctx)

	f, err := os.Open(path.Join(source.Path, r.config.Source))
	if err != nil {
		u.Step(terminal.StatusError, "Opening source file failed")
		return nil, err
	}

	if _, err := io.Copy(wc, f); err != nil {
		u.Step(terminal.StatusError, "Uploading file to Google Cloud Storage failed")
		return nil, err
	}

	if err := wc.Close(); err != nil {
		u.Step(terminal.StatusError, "Error closing GCS writer after file upload")
		return nil, err
	}

	attrs, err := object.Attrs(ctx)
	if err != nil {
		u.Step(terminal.StatusError, "Error fetching uploaded object attributes")
		return nil, err
	}

	ml := attrs.MediaLink
	sourceURL, err := stripQueryParams(ml)
	if err != nil {
		u.Step(terminal.StatusError, "Error parsing uploaded object url")
		return nil, err
	}

	u.Step(terminal.StatusOK, "Artifact saved to Google Cloud Storage: '"+sourceURL+"'")

	return &Artifact{Source: sourceURL}, nil
}

func stripQueryParams(u string) (string, error) {
	pu, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	pu.RawQuery = ""

	return pu.String(), nil
}
