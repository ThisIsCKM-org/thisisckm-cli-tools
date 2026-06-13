package release

import "testing"

func TestStateAdvance(t *testing.T) {
	s := Seed("0.1.0")

	var err error
	s, err = s.WithNew("0.1.0")
	if err != nil {
		t.Fatalf("new failed: %v", err)
	}
	if got, want := s.State, StateInProgress; got != want {
		t.Fatalf("state = %s, want %s", got, want)
	}

	s, err = s.Advance(ChannelAlpha)
	if err != nil {
		t.Fatalf("alpha failed: %v", err)
	}
	if got, want := s.Channel, ChannelAlpha; got != want {
		t.Fatalf("channel = %s, want %s", got, want)
	}
	if got, want := s.Counter, 1; got != want {
		t.Fatalf("counter = %d, want %d", got, want)
	}

	s, err = s.Advance(ChannelAlpha)
	if err != nil {
		t.Fatalf("alpha increment failed: %v", err)
	}
	if got, want := s.Counter, 2; got != want {
		t.Fatalf("counter = %d, want %d", got, want)
	}

	s, err = s.Advance(ChannelBeta)
	if err != nil {
		t.Fatalf("beta failed: %v", err)
	}
	if got, want := s.Channel, ChannelBeta; got != want {
		t.Fatalf("channel = %s, want %s", got, want)
	}
	if got, want := s.Counter, 1; got != want {
		t.Fatalf("counter = %d, want %d", got, want)
	}

	s, err = s.Advance(ChannelRC)
	if err != nil {
		t.Fatalf("rc failed: %v", err)
	}
	if got, want := s.Channel, ChannelRC; got != want {
		t.Fatalf("channel = %s, want %s", got, want)
	}

	s, err = s.Finalize()
	if err != nil {
		t.Fatalf("finalize failed: %v", err)
	}
	if got, want := s.Channel, ChannelStable; got != want {
		t.Fatalf("channel = %s, want %s", got, want)
	}
	if got, want := s.State, StateReleased; got != want {
		t.Fatalf("state = %s, want %s", got, want)
	}
}

func TestNewRejectsInProgress(t *testing.T) {
	s := Seed("0.1.0")
	s.State = StateInProgress
	s.Channel = ChannelAlpha
	if _, err := s.WithNew("0.2.0"); err == nil {
		t.Fatal("expected new to fail while release is in progress")
	}
}
