package cmd

import (
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
)

func TestLoginFileScrub(t *testing.T) {
	u, _ := url.Parse("http://localhost:9090/MidVision")

	// New-style save: credentials cleared before saving must not reach the file.
	c := &RDClient{BaseUrl: u, AuthToken: "tok123", Username: "", Password: ""}
	if err := c.saveLoginFile(); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(path.Join(getHome(), loginFile))
	if strings.Contains(string(content), "param1") || strings.Contains(string(content), "param2") {
		t.Fatalf("credentials keys present in new-style login file: %s", content)
	}

	// Old-style file with plaintext credentials: load must scrub memory and rewrite the file.
	// Old binaries saved the struct with the credentials still set, so simulate exactly that.
	oldClient := &RDClient{BaseUrl: u, AuthToken: "tok123", Username: "mvadmin", Password: "secretpw"}
	if err := oldClient.saveLoginFile(); err != nil {
		t.Fatal(err)
	}
	content, _ = os.ReadFile(path.Join(getHome(), loginFile))
	if !strings.Contains(string(content), "secretpw") {
		t.Fatalf("test setup failed to write old-style file with credentials: %s", content)
	}
	c2 := &RDClient{}
	if err := c2.loadLoginFile(); err != nil {
		t.Fatal(err)
	}
	if c2.Username != "" || c2.Password != "" {
		t.Fatalf("credentials not scrubbed from memory: %q %q", c2.Username, c2.Password)
	}
	if c2.AuthToken != "tok123" || c2.BaseUrl == nil {
		t.Fatalf("token or url lost during migration: %+v", c2)
	}
	content, _ = os.ReadFile(path.Join(getHome(), loginFile))
	if strings.Contains(string(content), "secretpw") || strings.Contains(string(content), "param2") {
		t.Fatalf("old login file not rewritten, still contains credentials: %s", content)
	}
	if !strings.Contains(string(content), "tok123") {
		t.Fatalf("rewritten login file lost the token: %s", content)
	}
}
