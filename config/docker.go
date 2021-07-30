package config

type DockerCertsConfiguration struct {
	Enable bool `ngx:"enable" help:"Use TLS"`
	Custom bool `flag:"-"`

	CaCertPath       string `ngx:"ca" flag:"ca" help:"Trust certs signed only by this CA"`
	CaPrivateKeyPath string `ngx:"ca-key" flag:"-"`

	ServerCertPath string `ngx:"server" flag:"cert" help:"Path to TLS certificate file"`
	ServerKeyPath  string `ngx:"server-key" flag:"key" help:"Path to TLS key file"`

	ClientCertPath string `ngx:"cert" flag:"-"`
	ClientKeyPath  string `ngx:"key" flag:"-"`
}

type DockerStorageConfiguration struct {
	Driver string   `ngx:"driver" help:"storage driver to use"`
	Opt    []string `ngx:"opt" help:"Storage driver options"`
}

type DockerDNSConfiguration struct {
	List   []string `ngx:"list" help:"DNS server to use"`
	Opt    []string `ngx:"opt" help:"DNS options to use"`
	Search []string `ngx:"search" help:"DNS search domains to use"`
}

type DockerConfiguration struct {
	Version       string   `ngx:"version" help:"docker version"`
	StraitVersion bool     `ngx:"strait-version" help:"Strict check DOCKER version if inconsistent will upgrade" def:"false"`
	DataRoot      string   `ngx:"data-root" help:"docker data root"`
	Hosts         []string `ngx:"hosts" help:"Daemon socket(s) to connect to" def:"fd://"`
	DaemonJson    string   `ngx:"daemon-json" help:"docker cfg file, if set this option, other option will ignore."`

	InsecureRegistries []string `help:"it replaces the daemon insecure registries with a new set of insecure registries."`
	RegistryMirrors    []string `ngx:"registry-mirrors" help:"preferred DockerConfiguration registry mirror"`

	Storage *DockerStorageConfiguration `ngx:"storage" flag:"storage"`
	DNS     *DockerDNSConfiguration     `flag:"dns" ngx:"dns" `
	TLS     *DockerCertsConfiguration   `ngx:"tls" flag:"tls"`
}

func DefaultDockerConfiguration() *DockerConfiguration {
	return &DockerConfiguration{
		Version:       "v19.03.15",
		DataRoot:      "/var/lib/docker",
		StraitVersion: false,
		TLS:           new(DockerCertsConfiguration),
	}
}
