package kubectl_test

import (
	"github.com/koirand/kubectl"
)

func Example() {
	const manifest string = `
apiVersion: v1
kind: Pod
metadata:
  name: {{ .Name }}
labels:
  app: {{ .Name }}
spec:
containers:
  - name: nginx
    image: nginx:latest
`

	k := kubectl.NewKubectl()

	if err := k.Apply(
		podManifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		panic(err)
	}
}
