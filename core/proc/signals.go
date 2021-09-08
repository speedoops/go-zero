package proc

var done = make(chan struct{})

// Done returns the channel that notifies the process quitting.
func Done() <-chan struct{} {
	return done
}
