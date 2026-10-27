package mesh

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDeduper(t *testing.T) {
	d := NewDeduper(time.Minute)
	if !d.First("a") || d.First("a") {
		t.Fatal("expected first delivery only")
	}
}
func TestRetryEventuallySucceeds(t *testing.T) {
	calls := 0
	p := RetryPolicy{Attempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}
	err := p.Run(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errors.New("temporary")
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}
