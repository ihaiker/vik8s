package config

type DockerCerts struct {
	Enable       bool   `ngx:"enable" help:"Use TLS"`
	Custom       bool   `flag:"-"`
	CaCert       string `ngx:"ca-cert" flag:"ca" help:"Trust certs signed only by this CA"`
	CaPrivateKey string `ngx:"ca-key" flag:"-"`
	ServerCert   string `ngx:"cert" flag:"cert" help:"Path to TLS certificate file"`
	ServerKey    string `ngx:"key" flag:"key" help:"Path to TLS key file"`
	ClientCert   string `ngx:"client-cert" flag:"-"`
	ClientKey    string `ngx:"client-key" flag:"-"`
}

type DockerStorage struct {
	Driver string   `ngx:"driver" help:"storage driver to use"`
	Opt    []string `ngx:"opt" help:"Storage driver options"`
}

type DockerDNS struct {
	List   []string `ngx:"list" help:"DNS server to use"`
	Opt    []string `ngx:"opt" help:"DNS options to use"`
	Search []string `ngx:"search" help:"DNS search domains to use"`
}

type DockerConfiguration struct {
	Version       string   `ngx:"version" help:"docker version"`
	StraitVersion bool     `ngx:"strait-version" help:"Strict check DOCKER version if inconsistent will upgrade" def:"false"`
	DataRoot      string   `ngx:"data-root" help:"docker data root"`
	Hosts         []string `ngx:"hosts" help:"Daemon socket(s) to connect to"`
	DaemonJson    string   `ngx:"daemon-json" help:"docker cfg file, if set this option, other option will ignore."`

	InsecureRegistries []string `help:"it replaces the daemon insecure registries with a new set of insecure registries."`
	RegistryMirrors    []string `ngx:"registry-mirrors" help:"preferred DockerConfiguration registry mirror"`

	Storage *DockerStorage `ngx:"storage" flag:"storage"`
	DNS     *DockerDNS     `flag:"dns" ngx:"dns" `
	TLS     *DockerCerts   `ngx:"tls" flag:"tls"`
}

func DefaultDockerConfiguration() *DockerConfiguration {
	return &DockerConfiguration{
		Version:       "v19.3.12",
		DataRoot:      "/var/lib/docker",
		StraitVersion: false,
		TLS:           new(DockerCerts),
	}
}
