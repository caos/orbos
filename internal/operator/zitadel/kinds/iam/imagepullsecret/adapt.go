package imagepullsecret

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/dockerconfigsecret"
)

func AdaptFunc(
	namespace string,
	name string,
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

	query, err := dockerconfigsecret.AdaptFuncToEnsure(name, namespace, labels, data)
	if err != nil {
		return nil, nil, err
	}
	destroy, err := dockerconfigsecret.AdaptFuncToDestroy(name, namespace)
	if err != nil {
		return nil, nil, err
	}
	return query, destroy, nil
}
