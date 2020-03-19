package main

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/koirand/kubectl"
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

func main() {
	// Create pod
	if err := kubectl.Apply(
		manifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		log.Fatal(err)
	}

	// Wait for pod running
	p := pod{}
	for {
		out, err := kubectl.Get("pod", "foo", "default")
		if err != nil {
			log.Fatal(err)
		}
		if err := json.Unmarshal(out, &p); err != nil {
			log.Fatal(err)
		}
		log.Println(p.Status.Phase)
		if p.Status.Phase == "Running" {
			break
		}
		time.Sleep(time.Second)
	}

	// Exec command
	out, err := kubectl.Exec(
		"foo",
		"default",
		"echo",
		"foo",
		"bar",
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(strings.TrimSpace(string(out)))

	// Delete pod
	if err := kubectl.Delete(
		manifest,
		map[string]string{
			"Name": "foo",
		},
	); err != nil {
		log.Fatal(err)
	}
}