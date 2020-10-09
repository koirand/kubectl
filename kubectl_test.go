package kubectl

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
	"time"
)

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

type pods struct {
	Items []pod `json:"items"`
}

type pod struct {
	Status struct {
		Phase string `json:"phase"`
	} `json:"status"`
}

func TestApply(t *testing.T) {
	k := NewKubectl()
	defer func() {
		exec.Command("kubectl", "delete", "pod", "foo").Run()
	}()

	// Normal
	if err := k.Apply(
		manifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}

	// Error
	if err := k.Apply(
		manifest,
		map[string]string{},
	); err == nil {
		t.Fatal("Expected error but not")
	}
}

func TestDelete(t *testing.T) {
	k := NewKubectl()
	defer func() {
		exec.Command("kubectl", "delete", "pod", "foo").Run()
	}()

	if err := k.Apply(
		manifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}

	// Error
	if err := k.Apply(
		manifest,
		map[string]string{},
	); err == nil {
		t.Fatal("Expected error but not")
	}

	// Normal
	if err := k.Delete(
		manifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}
}

func TestExec(t *testing.T) {
	k := NewKubectl()
	defer func() {
		exec.Command("kubectl", "delete", "pod", "foo").Run()
	}()

	if err := k.Apply(
		manifest,
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
		"false",
	)
	if err == nil {
		t.Fatal("Expected error but not")
	}
}

func TestGetByName(t *testing.T) {
	k := NewKubectl()
	defer func() {
		exec.Command("kubectl", "delete", "pod", "foo").Run()
	}()

	if err := k.Apply(
		manifest,
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
	_, err = k.GetByName("pod", "bar", "default")
	if err == nil {
		t.Fatal("Expected error but not")
	}
}

func TestGetWithLabel(t *testing.T) {
	k := NewKubectl()
	defer func() {
		exec.Command("kubectl", "delete", "pod", "foo").Run()
	}()

	if err := k.Apply(
		manifest,
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
	if len(pods.Items) == 0 {
		t.Fatal("Failed to get pods")
	}

	// Not Exist
	out, err = k.GetByLabel("pod", "app=bar", "default")
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
