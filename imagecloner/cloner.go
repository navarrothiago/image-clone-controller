package imagecloner

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/navarrothiago/image-clone-controller/config"
)

type cloner struct {
	log logr.Logger
	cfg config.Config
}

// Clone will check the availability on the image in public repository and
// copy the image to target repository.
func (c *cloner) Clone(ctx context.Context, sourceImage name.Reference, targetImage name.Reference) error {
	if err := c.isExistInPublic(ctx, sourceImage); err != nil {
		return err
	}

	img, err := remote.Image(sourceImage, remote.WithContext(ctx))
	if err != nil {
		return err
	}

	err = remote.Write(targetImage, img, remote.WithContext(ctx),
		remote.WithAuth(authn.FromConfig(authn.AuthConfig{Username: c.cfg.DockerUsername, Password: c.cfg.DockerPassword})))
	if err != nil {
		return err
	}
	return nil
}

func (c *cloner) isExistInPublic(ctx context.Context, sourceImage name.Reference) error {
	c.log.Info(fmt.Sprintf("Check if %s exists in public repository", sourceImage))
	_, err := remote.Head(sourceImage, remote.WithContext(ctx))
	if err != nil {
		c.log.Info(fmt.Sprintf("Image %s was not found in public repository", sourceImage))
		if e, ok := err.(*transport.Error); ok {
			c.log.Error(err, fmt.Sprintf(`error reading the source image:%s, errorcode:%d`, sourceImage.Name(), e.StatusCode))
		}
		return err
	}
	c.log.Info(fmt.Sprintf("Image %s was found in public repository!!", sourceImage))
	return nil
}

// IsExistInClones will check the image is previously cloned to target repository
func (c *cloner) IsExistInClones(ctx context.Context, targetImage name.Reference) (error, bool) {
	c.log.Info(fmt.Sprintf("Check if %s exists in cache (target repository)", targetImage))
	_, err := remote.Head(targetImage, remote.WithContext(ctx),
		remote.WithAuth(authn.FromConfig(authn.AuthConfig{Username: c.cfg.DockerUsername, Password: c.cfg.DockerPassword})))
	if err != nil {
		c.log.Info(fmt.Sprintf("Image %s was not found in cache (target repository)", targetImage))
		if e, ok := err.(*transport.Error); ok {
			c.log.Error(err, fmt.Sprintf(`error reading the source image:%s, errorcode:%d`, targetImage.Name(), e.StatusCode))
		}
		return err, false
	}
	c.log.Info(fmt.Sprintf("Image %s was found in cache (target repository)!!", targetImage))
	return nil, true
}

func NewCloner(l logr.Logger, c config.Config) Cloner {
	return &cloner{log: l, cfg: c}
}
