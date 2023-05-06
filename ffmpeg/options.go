package ffmpeg

const Stdin = "pipe:0"
const Stdout = "pipe:1"

type Coder interface {
	Name() []string
	Format() string
	Args() []string
}

type Options struct {
	Decoder          Coder
	Encoder          Coder
	Input            string // Input stream to read from (-i)
	Output           string // Output stream to write to
	Channels         string // Number of audio channels (-ac)
	Bitrate          string // Bitrate (-b:a)
	Quality          string // Quality of bitrate conversion 0-9 (-q:a)
	FrameRate        string // Audio sampling rate (-ar)
	StartTime        string // Time after 0 to start at in seconds (-ss)
	CompressionLevel string // Compression level between 0 and 10 (-compression_level)
	Threads          string // Number of threads to use (-threads)
}

func (o Options) Args() []string {
	args := make([]string, 0, 10)

	args = append(args, "-hide_banner", "-loglevel", "error")

	if o.Decoder != nil {
		args = append(args, o.Decoder.Name()...)
		args = append(args, o.Decoder.Args()...)
	}
	args = append(args, "-i", o.Input)

	if o.Input == Stdin {
		args = append(args, "-f", o.Encoder.Format())
	}

	if o.Encoder != nil {
		args = append(args, o.Encoder.Name()...)
		args = append(args, o.Encoder.Args()...)
	}

	if o.Bitrate != "" {
		args = append(args, "-b:a", o.Bitrate)
	}

	if o.Quality != "" {
		args = append(args, "-q:a", o.Quality)
	}

	if o.FrameRate != "" {
		args = append(args, "-ar", o.FrameRate)
	}

	if o.StartTime != "" {
		args = append(args, "-ss", o.StartTime)
	}

	if o.CompressionLevel != "" {
		args = append(args, "-compression_level", o.CompressionLevel)
	}

	if o.Threads != "" {
		args = append(args, "-threads", o.Threads)
	}

	args = append(args, o.Output)

	return args
}
