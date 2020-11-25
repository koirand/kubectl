package kubectl

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"text/template"
)

type Kubectl interface {
	Apply(manifest string, data interface{}) error
	Delete(manifest string, data interface{}) error
	Patch(resource string, name string, namespace string, patch string) error
	Exec(name string, namespace string, commands ...string) ([]byte, error)
	GetByName(resource string, name string, namespace string) ([]byte, error)
	GetByLabel(resource string, label string, namespace string) ([]byte, error)
	DeleteByLabel(resources []string, label string, namespace string) error
}

type kubectl struct {
}

func NewKubectl() Kubectl {
	return &kubectl{}
}

func (c *kubectl) Apply(manifest string, data interface{}) error {
	tpl, err := template.New("template").Parse(manifest)
	if err != nil {
		return fmt.Errorf("Cannot perse template: %s", err)
	}

	var stderr bytes.Buffer
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stderr = &stderr
	stdin, _ := cmd.StdinPipe()
	tpl.Execute(stdin, data)
	stdin.Close()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Cannot apply from manifest: %s: %s", err, stderr.String())
	}
	return nil
}

func (c *kubectl) Patch(resource string, name string, namespace string, patch string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("kubectl", "patch", resource, name, "-n", namespace, "--patch", patch)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Cannot patch: %s: %s", err, stderr.String())
	}
	return nil
}

func (c *kubectl) Delete(manifest string, data interface{}) error {
	tpl, err := template.New("template").Parse(manifest)
	if err != nil {
		return fmt.Errorf("Cannot perse template: %s", err)
	}

	var stderr bytes.Buffer
	cmd := exec.Command("kubectl", "delete", "-f", "-")
	cmd.Stderr = &stderr
	stdin, _ := cmd.StdinPipe()
	tpl.Execute(stdin, data)
	stdin.Close()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Cannot delete from manifest: %s: %s", err, stderr.String())
	}
	return nil
}

func (c *kubectl) Exec(name string, namespace string, commands ...string) ([]byte, error) {
	args := []string{"exec", name, "-n", namespace}
	for _, c := range commands {
		args = append(args, c)
	}

	var stderr bytes.Buffer
	cmd := exec.Command("kubectl", args...)
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Cannot exec command: %s: %s", err, stderr.String())
	}
	return out, nil
}

func (c *kubectl) GetByName(resource string, name string, namespace string) ([]byte, error) {
	var stderr bytes.Buffer
	cmd := exec.Command("kubectl", "get", resource, name, "-n="+namespace, "-o", "json")
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Cannot get resource: %s: %s", err, stderr.String())
	}
	return b, nil
}

func (c *kubectl) GetByLabel(resource string, label string, namespace string) ([]byte, error) {
	var stderr bytes.Buffer
	cmd := exec.Command("kubectl", "get", resource, "-l "+label, "-n="+namespace, "-o", "json")
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Cannot get resource: %s: %s", err, stderr.String())
	}
	return b, nil
}

func (c *kubectl) DeleteByLabel(resources []string, label string, namespace string) error {
	resource := strings.Join(resources, ",")
	var stderr bytes.Buffer
	cmd := exec.Command("kubectl", "delete", resource, "-l "+label, "-n="+namespace)
	cmd.Stderr = &stderr
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Cannot delete resource: %s: %s", err, stderr.String())
	}
	return nil
}
