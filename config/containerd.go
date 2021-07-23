package config

type ContainerdConfiguration struct {
	Version         string   `ngx:"version"`
	DataRoot        string   `ngx:"data-root"`
	RegistryMirrors []string `ngx:"registry-mirrors"`
	StraitVersion   bool     `ngx:"strait-version"`
}

func DefaultContainerdConfiguration() *ContainerdConfiguration {
	return &ContainerdConfiguration{
		Version:  "v1.21.2",
		DataRoot: "/var/lib/containerd",
	}
}
