package sshserver

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func TestServerRunsShellSession(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := Server{
		HostKeyPath: filepath.Join(t.TempDir(), "ssh_host_ed25519"),
		Runner:      scriptedRunner{},
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(ctx, listener)
	}()

	client, err := ssh.Dial("tcp", listener.Addr().String(), &ssh.ClientConfig{
		User:            "review",
		Auth:            []ssh.AuthMethod{ssh.Password("anything")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
	if err != nil {
		t.Fatalf("dial ssh server: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("new ssh session: %v", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stdin, err := session.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	if err := session.RequestPty("xterm", 24, 80, ssh.TerminalModes{}); err != nil {
		t.Fatalf("request pty: %v", err)
	}
	if err := session.Shell(); err != nil {
		t.Fatalf("start shell: %v", err)
	}
	if _, err := stdin.Write([]byte("Q\r")); err != nil {
		t.Fatalf("write command: %v", err)
	}
	_ = stdin.Close()

	output, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	_ = session.Wait()

	text := string(output)
	for _, want := range []string{"EMPIRE ASCENDANT", "Goodbye."} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q:\n%s", want, text)
		}
	}

	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("serve returned error: %v", err)
	}
}

func TestTerminalReaderEchoesAndEditsLine(t *testing.T) {
	var echo bytes.Buffer
	reader := &terminalReader{
		r:    strings.NewReader("Qx\x7f\r"),
		echo: &echo,
	}

	got, err := bufio.NewReader(reader).ReadString('\n')
	if err != nil {
		t.Fatalf("read line: %v", err)
	}
	if got != "Q\n" {
		t.Fatalf("line = %q, want %q", got, "Q\n")
	}
	if echo.String() != "Qx\b \b\r\n" {
		t.Fatalf("echo = %q, want %q", echo.String(), "Qx\b \b\r\n")
	}
}

func TestTerminalWriterConvertsFramesToCP437(t *testing.T) {
	var output bytes.Buffer
	writer := &terminalWriter{w: &output, encoding: terminalEncodingCP437}

	n, err := writer.Write([]byte("┌─┐\n│X│\n└─┘"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if n != len("┌─┐\n│X│\n└─┘") {
		t.Fatalf("write count = %d", n)
	}

	want := []byte{
		0xda, 0xc4, 0xbf, '\r', '\n',
		0xb3, 'X', 0xb3, '\r', '\n',
		0xc0, 0xc4, 0xd9,
	}
	if !bytes.Equal(output.Bytes(), want) {
		t.Fatalf("output bytes = % x, want % x", output.Bytes(), want)
	}
}

func TestTerminalWriterPreservesUTF8Frames(t *testing.T) {
	var output bytes.Buffer
	writer := &terminalWriter{w: &output, encoding: terminalEncodingUTF8}

	n, err := writer.Write([]byte("┌─┐\n│X│\n└─┘"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if n != len("┌─┐\n│X│\n└─┘") {
		t.Fatalf("write count = %d", n)
	}

	want := "┌─┐\r\n│X│\r\n└─┘"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestResolveTerminalEncoding(t *testing.T) {
	tests := []struct {
		configured string
		term       string
		want       string
	}{
		{configured: "cp437", term: "xterm-256color", want: terminalEncodingCP437},
		{configured: "utf8", term: "ansi", want: terminalEncodingUTF8},
		{configured: "auto", term: "xterm-256color", want: terminalEncodingUTF8},
		{configured: "auto", term: "ansi", want: terminalEncodingCP437},
		{configured: "auto", term: "syncterm", want: terminalEncodingCP437},
		{configured: "", term: "", want: terminalEncodingCP437},
	}
	for _, tt := range tests {
		if got := resolveTerminalEncoding(tt.configured, tt.term); got != tt.want {
			t.Fatalf("resolveTerminalEncoding(%q, %q) = %q, want %q", tt.configured, tt.term, got, tt.want)
		}
	}
}

type scriptedRunner struct{}

func (scriptedRunner) Run(_ context.Context, input io.Reader, output io.Writer) error {
	fmt.Fprintln(output, "EMPIRE ASCENDANT")
	fmt.Fprint(output, "Command: ")
	line, err := bufio.NewReader(input).ReadString('\n')
	if err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(line), "Q") {
		fmt.Fprintln(output, "Goodbye.")
	}
	return nil
}
