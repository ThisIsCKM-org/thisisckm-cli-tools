package release

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateAdvance(t *testing.T) {
	s := Seed("0.1.0")

	var err error
	s, err = s.Advance(ChannelAlpha)
	if err != nil {
		t.Fatalf("alpha failed: %v", err)
	}
	if got, want := s.State, StateInProgress; got != want {
		t.Fatalf("state = %s, want %s", got, want)
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
	if got, want := s.ReleasedVersion, "0.1.0"; got != want {
		t.Fatalf("releasedVersion = %s, want %s", got, want)
	}
	if got, want := s.LastTag, "v0.1.0"; got != want {
		t.Fatalf("lastTag = %s, want %s", got, want)
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

func TestLoadStateFileFallsBackToLegacyVersionFile(t *testing.T) {
	root := t.TempDir()
	legacy := Seed("0.1.0")
	legacy.State = StateInProgress
	legacy.Channel = ChannelAlpha
	legacy.Counter = 2
	if err := Save(LegacyStateFile(root), legacy); err != nil {
		t.Fatalf("save legacy state: %v", err)
	}
	state, path, err := loadStateFile(root)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if got, want := path, LegacyStateFile(root); got != want {
		t.Fatalf("state path = %q, want %q", got, want)
	}
	if got, want := state.Channel, ChannelAlpha; got != want {
		t.Fatalf("channel = %s, want %s", got, want)
	}
	if got, want := state.Counter, 2; got != want {
		t.Fatalf("counter = %d, want %d", got, want)
	}
}

func TestEnsureReleaseFileMigratesLegacyVersionFile(t *testing.T) {
	root := t.TempDir()
	legacy := Seed("0.1.0")
	legacy.State = StateInProgress
	legacy.Channel = ChannelAlpha
	legacy.Counter = 2
	if err := Save(LegacyStateFile(root), legacy); err != nil {
		t.Fatalf("save legacy state: %v", err)
	}
	if err := EnsureReleaseFile(root, "0.1.0"); err != nil {
		t.Fatalf("ensure release file: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "release.json"))
	if err != nil {
		t.Fatalf("read migrated release.json: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected migrated release.json content")
	}
	state, err := Load(StateFile(root))
	if err != nil {
		t.Fatalf("load migrated release.json: %v", err)
	}
	if got, want := state.Channel, ChannelAlpha; got != want {
		t.Fatalf("channel = %s, want %s", got, want)
	}
	if got, want := state.Counter, 2; got != want {
		t.Fatalf("counter = %d, want %d", got, want)
	}
}
