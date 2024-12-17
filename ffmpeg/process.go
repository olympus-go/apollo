package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/olympus-go/apollo"
)

type Process struct {
	r      io.ReadCloser
	codec  apollo.Codec
	opts   Options
	err    error
	cancel context.CancelFunc
}

func Version() string {
	var version string

	out, err := exec.Command("ffmpeg", "-version").Output()
	if err != nil {
		return fmt.Sprintf("version unknown: %s", err)
	}

	if _, err = fmt.Sscanf(string(out), "ffmpeg version %s Copyright", &version); err != nil {
		return fmt.Sprintf("could not parse version\n%s", string(out))
	}

	return version
}

func New(opts Options) *Process {
	return &Process{
		opts: opts,
	}
}

// WithCodec sets an additional codec for processed data to be passed through before returning.
func (p *Process) WithCodec(codec apollo.Codec) *Process {
	p.codec = codec
	return p
}

func (p *Process) Open(r io.Reader) error {
	var stderr bytes.Buffer
	var ctx context.Context
	var err error

	ctx, p.cancel = context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "ffmpeg", p.opts.Args()...)

	p.r, err = cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if p.codec != nil {
		if err = p.codec.Open(p.r); err != nil {
			return err
		}
	}

	cmd.Stdin = r
	cmd.Stderr = &stderr

	// Start the process
	if err = cmd.Start(); err != nil {
		return err
	}

	// Wait for the process to end naturally. If it doesn't exit 0, send the error + stderr on the errors channel.
	go func() {
		state, _ := cmd.Process.Wait()
		if state.ExitCode() != 0 && state.String() != "signal: killed" {
			p.err = fmt.Errorf("%s: %s", state.String(), stderr.Bytes())
		}
	}()

	return nil
}

func (p *Process) Read(b []byte) (int, error) {
	if p.err != nil {
		return 0, p.err
	}

	if p.codec != nil {
		return p.codec.Read(b)
	}

	return p.r.Read(b)
}

func (p *Process) Close() error {
	if p.codec != nil {
		p.codec.Close()
	}
	p.cancel()
	p.err = nil
	return p.r.Close()
}
