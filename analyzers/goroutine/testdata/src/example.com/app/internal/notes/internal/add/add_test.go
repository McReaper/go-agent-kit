package add

import "testing"

func TestRun(t *testing.T) {
	done := make(chan struct{})
	go close(done)
	<-done
}
