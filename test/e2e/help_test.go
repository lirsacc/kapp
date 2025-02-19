package e2e

import (
	"strings"
	"testing"
)

func TestHelpCommandGroup(t *testing.T) {
	env := BuildEnv(t)
	kapp := Kapp{t, env.Namespace, Logger{}}

	_, err := kapp.RunWithOpts([]string{"app-group"}, RunOpts{NoNamespace: true, AllowError: true})
	if err == nil {
		t.Fatalf("Expected error")
	}
	if !strings.Contains(err.Error(), "Error: Use one of available subcommands: delete, deploy") {
		t.Fatalf("Expected helpful error message, but was '%s'", err.Error())
	}
}
