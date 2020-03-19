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
spec:
  containers:
    - name: nginx
      image: nginx:latest
`

type pod struct {
	Status struct {
		Phase string `json:"phase"`
	} `json:"status"`
}

func TestApply(t *testing.T) {
	defer func() {
		exec.Command("kubectl", "delete", "pod", "foo").Run()
	}()

	// Normal
	if err := Apply(
		manifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}

	// Error
	if err := Apply(
		manifest,
		map[string]string{},
	); err == nil {
		t.Fatal("Expected error but not")
	}
}

func TestDelete(t *testing.T) {
	defer func() {
		exec.Command("kubectl", "delete", "pod", "foo").Run()
	}()

	if err := Apply(
		manifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}

	// Error
	if err := Apply(
		manifest,
		map[string]string{},
	); err == nil {
		t.Fatal("Expected error but not")
	}

	// Normal
	if err := Delete(
		manifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}
}

func TestExec(t *testing.T) {
	defer func() {
		exec.Command("kubectl", "delete", "pod", "foo").Run()
	}()

	if err := Apply(
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
		out, err := Get("pod", "foo", "default")
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
	out, err := Exec(
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
	_, err = Exec(
		"foo",
		"default",
		"false",
	)
	if err == nil {
		t.Fatal("Expected error but not")
	}
}

func TestGet(t *testing.T) {
	defer func() {
		exec.Command("kubectl", "delete", "pod", "foo").Run()
	}()

	if err := Apply(
		manifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		t.Fatal(err)
	}

	// Normal
	p := pod{}
	out, err := Get("pod", "foo", "default")
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
	_, err = Get("pod", "bar", "default")
	if err == nil {
		t.Fatal("Expected error but not")
	}
}
