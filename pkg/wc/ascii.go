package wc

var asciiSpace = func() [256]bool {
	var t [256]bool
	t['\t'] = true
	t['\n'] = true
	t['\v'] = true
	t['\f'] = true
	t['\r'] = true
	t[' '] = true
	return t
}()

