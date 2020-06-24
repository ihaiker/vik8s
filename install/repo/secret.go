package repo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ihaiker/vik8s/install/tools"
	"strings"
)

type Repo struct {
	Url      string `flag:"" help:"Choose a container registry to pull images."`
	User     string `flag:"user" help:"docker registry user"`
	Password string `flag:"password" help:"docker registry password"`

	//输出 Secret kubernetes yaml 定义
	Secret string `flag:"-"`

	//输出imagePullSecrets定义
	PullSecrets string `flag:"-"`
}

func (repo *Repo) String() string {
	return repo.Url
}

func (repo *Repo) Set(namespace string) {
	repo.Url = Suffix(repo.Url)
	if repo.User != "" {
		auth := tools.Json{
			"auths": tools.Json{
				"http://" + strings.TrimRight(repo.Url, "/"): tools.Json{
					"username": repo.User,
					"password": repo.Password,
					"auth":     base64.StdEncoding.EncodeToString([]byte(repo.User + ":" + repo.Password)),
				},
				//两个都加上，这样就不用判断了
				"https://" + strings.TrimRight(repo.Url, "/"): tools.Json{
					"username": repo.User,
					"password": repo.Password,
					"auth":     base64.StdEncoding.EncodeToString([]byte(repo.User + ":" + repo.Password)),
				},
			},
		}
		base, _ := json.Marshal(auth)
		repo.Secret = fmt.Sprintf(`---
apiVersion: v1
kind: Secret
type: kubernetes.io/dockerconfigjson
metadata:
  name: docker-auth
  namespace: %s
data:
  .dockerconfigjson: %s
`, namespace, base64.StdEncoding.EncodeToString(base))
		repo.PullSecrets = "imagePullSecrets:\n  - name: docker-auth"
	}
}

func (repo *Repo) Default(namespace, def string) {
	if repo.Url == "" {
		repo.Url = def
	}
	repo.Set(namespace)
}

func (repo *Repo) QuayIO(namespace string) {
	repo.Default(namespace, QuayIO(repo.Url))
}
