package cache

import "sync"

// we want a fan-in as described in https://blog.golang.org/pipelines
//
// generics would also be great so that we wouldn't have to literally copy-paste
// this next function to manually monomorphize it, but if we're wishing for
// a better go, we may as well just wish for something like std::sync::mpsc,
// which has the important property of never panicing.

func mergeKV(done <-chan struct{}, cs ...<-chan kv) <-chan kv {
	var wg sync.WaitGroup
	out := make(chan kv)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c or done is closed, then calls
	// wg.Done.
	output := func(c <-chan kv) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func mergeErr(done <-chan struct{}, cs ...<-chan error) <-chan error {
	var wg sync.WaitGroup
	out := make(chan error)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c or done is closed, then calls
	// wg.Done.
	output := func(c <-chan error) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
