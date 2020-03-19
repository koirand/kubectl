package kubectl

import (
	"bytes"
	"fmt"
	"os/exec"
	"text/template"
)

func Apply(manifest string, params map[string]string) error {
	tpl, err := template.New("template").Parse(manifest)
	if err != nil {
		return fmt.Errorf("Cannot perse template: %s", err)
	}

	var stderr bytes.Buffer
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stderr = &stderr
	stdin, _ := cmd.StdinPipe()
	tpl.Execute(stdin, params)
	stdin.Close()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Cannot apply from manifest: %s: %s", err, stderr.String())
	}
	return nil
}

func Delete(manifest string, params map[string]string) error {
	tpl, err := template.New("template").Parse(manifest)
	if err != nil {
		return fmt.Errorf("Cannot perse template: %s", err)
	}

	var stderr bytes.Buffer
	cmd := exec.Command("kubectl", "delete", "-f", "-")
	cmd.Stderr = &stderr
	stdin, _ := cmd.StdinPipe()
	tpl.Execute(stdin, params)
	stdin.Close()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Cannot delete from manifest: %s: %s", err, stderr.String())
	}
	return nil
}

func Exec(name string, namespace string, commands ...string) ([]byte, error) {
	args := []string{
		"exec",
		name,
		"-n=" + namespace,
	}
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

func Get(resource string, name string, namespace string) ([]byte, error) {
	var stderr bytes.Buffer
	cmd := exec.Command("kubectl", "get", resource, name, "-n="+namespace, "-o", "json")
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Cannot get resource: %s: %s", err, stderr.String())
	}
	return b, nil
}
