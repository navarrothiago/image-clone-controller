package config

// Conf is the application config object
type Config struct {
	DockerRegistry string `env:"DOCKER_REGISTRY" envDefault:"index.docker.io"`
	DockerUsername string `env:"DOCKER_USERNAME"`
	DockerPassword string `env:"DOCKER_PASSWORD"`
}
