package sshserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/ssh"
)

type Runner interface {
	Run(context.Context, io.Reader, io.Writer) error
}

type Server struct {
	Addr             string
	HostKeyPath      string
	TerminalEncoding string
	Runner           Runner
}

func (s Server) ListenAndServe(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("listen ssh: %w", err)
	}
	return s.Serve(ctx, listener)
}

func (s Server) Serve(ctx context.Context, listener net.Listener) error {
	if s.Runner == nil {
		return errors.New("ssh server runner is required")
	}

	signer, err := LoadOrCreateHostKey(s.HostKeyPath)
	if err != nil {
		return err
	}
	config := &ssh.ServerConfig{
		PasswordCallback: func(_ ssh.ConnMetadata, _ []byte) (*ssh.Permissions, error) {
			return nil, nil
		},
	}
	config.AddHostKey(signer)

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("accept ssh connection: %w", err)
		}
		go s.handleConn(ctx, conn, config)
	}
}

func (s Server) handleConn(ctx context.Context, conn net.Conn, config *ssh.ServerConfig) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		_ = conn.Close()
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "session channels only")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go s.handleSession(ctx, channel, requests)
	}
}

func (s Server) handleSession(ctx context.Context, channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	terminalType := ""
	for req := range requests {
		switch req.Type {
		case "pty-req":
			terminalType = parsePTYTerminal(req.Payload)
			_ = req.Reply(true, nil)
		case "shell":
			_ = req.Reply(true, nil)
			encoding := resolveTerminalEncoding(s.TerminalEncoding, terminalType)
			_ = s.Runner.Run(ctx, &terminalReader{r: channel, echo: channel}, &terminalWriter{w: channel, encoding: encoding})
			return
		default:
			if req.WantReply {
				_ = req.Reply(false, nil)
			}
		}
	}
}

type terminalWriter struct {
	w        io.Writer
	encoding string
}

type ptyRequest struct {
	Term    string
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
	Modes   string
}

type terminalReader struct {
	r       io.Reader
	echo    io.Writer
	line    []byte
	pending []byte
	lastCR  bool
}

func (r *terminalReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	for {
		if len(r.pending) > 0 {
			n := copy(p, r.pending)
			r.pending = r.pending[n:]
			return n, nil
		}

		var buf [1]byte
		n, err := r.r.Read(buf[:])
		if n == 0 {
			return 0, err
		}
		b := buf[0]

		if r.lastCR {
			r.lastCR = false
			if b == '\n' {
				continue
			}
		}

		switch b {
		case '\r':
			r.lastCR = true
			r.finishLine()
		case '\n':
			r.finishLine()
		case 0x7f, 0x08:
			r.backspace()
		case 0x03:
			_, _ = r.echo.Write([]byte("^C\r\n"))
			return 0, io.EOF
		default:
			if b < 0x20 && b != '\t' {
				continue
			}
			r.line = append(r.line, b)
			_, _ = r.echo.Write([]byte{b})
		}

		if len(r.pending) == 0 {
			continue
		}
	}
}

func (r *terminalReader) finishLine() {
	r.line = append(r.line, '\n')
	r.pending = append(r.pending, r.line...)
	r.line = r.line[:0]
	_, _ = r.echo.Write([]byte("\r\n"))
}

func (r *terminalReader) backspace() {
	if len(r.line) == 0 {
		return
	}
	r.line = r.line[:len(r.line)-1]
	_, _ = r.echo.Write([]byte("\b \b"))
}

func (w *terminalWriter) Write(p []byte) (int, error) {
	originalLen := len(p)
	out := make([]byte, 0, len(p)*2)
	for len(p) > 0 {
		r, size := utf8.DecodeRune(p)
		if r == utf8.RuneError && size == 1 && p[0] >= 0x80 {
			out = append(out, '?')
			p = p[1:]
			continue
		}
		p = p[size:]
		if r == '\n' {
			out = append(out, '\r', '\n')
			continue
		}
		if w.encoding == terminalEncodingCP437 {
			out = appendTerminalRuneCP437(out, r)
		} else {
			out = utf8.AppendRune(out, r)
		}
	}
	_, err := w.w.Write(out)
	if err != nil {
		return 0, err
	}
	return originalLen, nil
}

const (
	terminalEncodingAuto  = "auto"
	terminalEncodingCP437 = "cp437"
	terminalEncodingUTF8  = "utf8"
)

func parsePTYTerminal(payload []byte) string {
	var req ptyRequest
	if err := ssh.Unmarshal(payload, &req); err != nil {
		return ""
	}
	return strings.ToLower(req.Term)
}

func resolveTerminalEncoding(configured, terminalType string) string {
	switch strings.ToLower(strings.TrimSpace(configured)) {
	case terminalEncodingCP437:
		return terminalEncodingCP437
	case terminalEncodingUTF8, "utf-8":
		return terminalEncodingUTF8
	}
	term := strings.ToLower(terminalType)
	switch {
	case strings.Contains(term, "sync"), strings.Contains(term, "ansi"), strings.Contains(term, "bbs"), strings.Contains(term, "pcansi"):
		return terminalEncodingCP437
	case strings.Contains(term, "xterm"), strings.Contains(term, "screen"), strings.Contains(term, "tmux"), strings.Contains(term, "linux"), strings.Contains(term, "vt"), strings.Contains(term, "rxvt"), strings.Contains(term, "kitty"), strings.Contains(term, "alacritty"), strings.Contains(term, "foot"):
		return terminalEncodingUTF8
	default:
		return terminalEncodingCP437
	}
}

func appendTerminalRuneCP437(out []byte, r rune) []byte {
	if r < 0x80 {
		return append(out, byte(r))
	}
	if b, ok := cp437Output[r]; ok {
		return append(out, b)
	}
	return append(out, '?')
}

var cp437Output = map[rune]byte{
	'░': 0xb0,
	'▒': 0xb1,
	'▓': 0xb2,
	'│': 0xb3,
	'┤': 0xb4,
	'╡': 0xb5,
	'╢': 0xb6,
	'╖': 0xb7,
	'╕': 0xb8,
	'╣': 0xb9,
	'║': 0xba,
	'╗': 0xbb,
	'╝': 0xbc,
	'╜': 0xbd,
	'╛': 0xbe,
	'┐': 0xbf,
	'└': 0xc0,
	'┴': 0xc1,
	'┬': 0xc2,
	'├': 0xc3,
	'─': 0xc4,
	'┼': 0xc5,
	'╞': 0xc6,
	'╟': 0xc7,
	'╚': 0xc8,
	'╔': 0xc9,
	'╩': 0xca,
	'╦': 0xcb,
	'╠': 0xcc,
	'═': 0xcd,
	'╬': 0xce,
	'╧': 0xcf,
	'╨': 0xd0,
	'╤': 0xd1,
	'╥': 0xd2,
	'╙': 0xd3,
	'╘': 0xd4,
	'╒': 0xd5,
	'╓': 0xd6,
	'╫': 0xd7,
	'╪': 0xd8,
	'┘': 0xd9,
	'┌': 0xda,
	'█': 0xdb,
	'▄': 0xdc,
	'▌': 0xdd,
	'▐': 0xde,
	'▀': 0xdf,
}
