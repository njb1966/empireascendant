package sshserver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrCreateHostKeyCreatesAndReusesKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ssh_host_ed25519")

	first, err := LoadOrCreateHostKey(path)
	if err != nil {
		t.Fatalf("create host key: %v", err)
	}
	second, err := LoadOrCreateHostKey(path)
	if err != nil {
		t.Fatalf("reuse host key: %v", err)
	}

	if string(first.PublicKey().Marshal()) != string(second.PublicKey().Marshal()) {
		t.Fatal("expected reused host key to have the same public key")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat host key: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("host key permissions = %v, want 0600", info.Mode().Perm())
	}
}
