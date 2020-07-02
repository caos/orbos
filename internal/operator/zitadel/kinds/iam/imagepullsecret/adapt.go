package imagepullsecret

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/dockerconfigsecret"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
) (
	resources.QueryFunc,
	resources.DestroyFunc,
	error,
) {
	data := `{
		"auths": {
				"docker.pkg.github.com": {
						"auth": "aW1ncHVsbGVyOmU2NTAxMWI3NDk1OGMzOGIzMzcwYzM5Zjg5MDlkNDE5OGEzODBkMmM="
				}
		}
}`

	return dockerconfigsecret.AdaptFunc("public-github-packages", namespace, labels, data)
}
