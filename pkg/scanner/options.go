package scanner

type Options struct {
	Threads int
	Lines   int
}

type FnOption func(*Options)

func buildOptions(passedOpts []FnOption) *Options {
	opts := &Options{}
	for _, o := range passedOpts {
		o(opts)
	}
	return opts
}

func WithThreads(tNum int) FnOption {
	return func(o *Options) {
		o.Threads = tNum
	}
}

func WithNumLines(lines int) FnOption {
	return func(o *Options) {
		o.Lines = lines
	}
}
