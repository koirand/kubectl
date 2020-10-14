package kubectl_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/koirand/kubectl"
)

const replicasetManifest string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: {{ .Name }}
spec:
  selector:
    matchLabels:
      app: {{ .Name }}
  replicas: 2
  template:
    metadata:
      labels:
        app: {{ .Name }}
    spec:
      containers:
        - name: nginx
          image: nginx:latest
`

const podManifest string = `
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

type replicaset struct {
	Status struct {
		AvailableReplicas string `json:"availableReplicas"`
	} `json:"status"`
}

type pods struct {
	Items []pod `json:"items"`
}

type pod struct {
	Status struct {
		Phase string `json:"phase"`
	} `json:"status"`
}

func TestApply(t *testing.T) {
	k := kubectl.NewKubectl()
	defer func() {
		exec.Command("kubectl", "delete", "pod", "foo").Run()
	}()

	// Normal
	if err := k.Apply(
		podManifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}

	// Error
	if err := k.Apply(
		podManifest,
		map[string]string{}, // no param
	); err == nil {
		t.Fatal("Expected error but not")
	}
}

func TestDelete(t *testing.T) {
	k := kubectl.NewKubectl()
	defer func() {
		exec.Command("kubectl", "delete", "pod", "foo").Run()
	}()

	if err := k.Apply(
		podManifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}

	// Error
	if err := k.Delete(
		podManifest,
		map[string]string{}, // no param
	); err == nil {
		t.Fatal("Expected error but not")
	}

	// Normal
	if err := k.Delete(
		podManifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}
}

func TestPatch(t *testing.T) {
	k := kubectl.NewKubectl()
	defer func() {
		exec.Command("kubectl", "delete", "replicaset", "foo").Run()
	}()

	if err := k.Apply(
		replicasetManifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}

	// Normal
	if err := k.Patch(
		"replicaset",
		"foo",
		`{"spec":{"replicas":0}}`,
	); err != nil {
		t.Fatal(err)
	}

	// Error
	if err := k.Patch(
		"replicaset",
		"invalid", // invalid name
		`{"spec":{"replicas":0}}`,
	); err == nil {
		t.Fatal("Expected error but not")
	}
}

func TestExec(t *testing.T) {
	k := kubectl.NewKubectl()
	defer func() {
		exec.Command("kubectl", "delete", "pod", "foo").Run()
	}()

	if err := k.Apply(
		podManifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}

	// Wait for pod running
	p := pod{}
	for {
		out, err := k.GetByName("pod", "foo", "default")
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(out, &p); err != nil {
			t.Fatal(err)
		}
		if p.Status.Phase == "Running" {
			break
		}
		time.Sleep(time.Second)
	}

	// Normal
	out, err := k.Exec(
		"foo",
		"default",
		"echo",
		"foo",
		"bar",
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(out)) != "foo bar" {
		t.Fatalf("Expected %s for Output, but %s:", "foo bar", out)
	}

	// Error
	_, err = k.Exec(
		"foo",
		"default",
		"false", // fail command
	)
	if err == nil {
		t.Fatal("Expected error but not")
	}
}

func TestGetByName(t *testing.T) {
	k := kubectl.NewKubectl()
	defer func() {
		exec.Command("kubectl", "delete", "pod", "foo").Run()
	}()

	if err := k.Apply(
		podManifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}

	// Normal
	p := pod{}
	out, err := k.GetByName("pod", "foo", "default")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out, &p); err != nil {
		t.Fatal(err)
	}
	if p.Status.Phase == "" {
		t.Fatal("Failed to get pod status")
	}

	// Error
	_, err = k.GetByName("pod", "invalid", "default") // invalid name
	if err == nil {
		t.Fatal("Expected error but not")
	}
}

func TestGetWithLabel(t *testing.T) {
	k := kubectl.NewKubectl()
	defer func() {
		exec.Command("kubectl", "delete", "replicaset", "foo").Run()
	}()

	if err := k.Apply(
		replicasetManifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}

	// Normal
	pods := pods{}
	out, err := k.GetByLabel("pod", "app=foo", "default")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out, &pods); err != nil {
		t.Fatal(err)
	}
	if len(pods.Items) != 2 {
		t.Fatal("Failed to get pods")
	}

	// Not Exist
	out, err = k.GetByLabel("pod", "app=invalid", "default") // invalid label
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out, &pods); err != nil {
		t.Fatal(err)
	}
	if len(pods.Items) != 0 {
		t.Fatalf("Pods count is expected 0, but was %v", len(pods.Items))
	}
}
