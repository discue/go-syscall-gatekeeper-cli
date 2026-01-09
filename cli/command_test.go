package cli

import (
	"reflect"
	"testing"
)

func TestParseAllowFileSystemPaths(t *testing.T) {
	c := NewCommand()
	args := []string{"--allow-file-system-path=/a", "--allow-file-system-path=/b", "--allow-file-system-path=/c"}
	if err := c.Parse(args); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	exp := []string{"/a", "/b", "/c"}
	if !reflect.DeepEqual(c.AllowFileSystemPathsList, exp) {
		t.Fatalf("expected %v got %v", exp, c.AllowFileSystemPathsList)
	}
}
