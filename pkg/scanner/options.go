package scanner

type Options struct {
	Threads int
	Lines   int
}

type FnOption func(*Options)

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
