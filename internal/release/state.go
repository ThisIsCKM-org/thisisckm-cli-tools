package release

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Channel string

const (
	ChannelNone   Channel = "none"
	ChannelAlpha  Channel = "alpha"
	ChannelBeta   Channel = "beta"
	ChannelRC     Channel = "rc"
	ChannelStable Channel = "stable"
)

type StateName string

const (
	StatePlanned    StateName = "planned"
	StateInProgress StateName = "in-progress"
	StateReleased   StateName = "released"
)

type State struct {
	BaseVersion     string    `json:"baseVersion"`
	Channel         Channel   `json:"channel"`
	Counter         int       `json:"counter"`
	ReleasedVersion string    `json:"releasedVersion"`
	LastTag         string    `json:"lastTag"`
	State           StateName `json:"state"`
}

var baseVersionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

func Seed(version string) State {
	return State{
		BaseVersion:     version,
		Channel:         ChannelNone,
		Counter:         0,
		ReleasedVersion: "",
		LastTag:         "",
		State:           StatePlanned,
	}
}

func Load(path string) (State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, err
	}
	return s, s.Validate()
}

func Save(path string, state State) error {
	if err := state.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func (s State) Validate() error {
	if !baseVersionPattern.MatchString(s.BaseVersion) {
		return fmt.Errorf("invalid baseVersion %q", s.BaseVersion)
	}
	switch s.Channel {
	case ChannelNone, ChannelAlpha, ChannelBeta, ChannelRC, ChannelStable:
	default:
		return fmt.Errorf("invalid channel %q", s.Channel)
	}
	switch s.State {
	case StatePlanned, StateInProgress, StateReleased:
	default:
		return fmt.Errorf("invalid state %q", s.State)
	}
	if s.Counter < 0 {
		return errors.New("counter cannot be negative")
	}
	if s.State == StateReleased && s.Channel != ChannelStable {
		return errors.New("released state requires stable channel")
	}
	if s.Channel == ChannelStable && s.State != StateReleased {
		return errors.New("stable channel requires released state")
	}
	if s.Channel == ChannelNone && s.Counter != 0 {
		return errors.New("counter must be zero when channel is none")
	}
	return nil
}

func (s State) InProgress() bool {
	return s.State == StateInProgress || s.Channel == ChannelAlpha || s.Channel == ChannelBeta || s.Channel == ChannelRC
}

func (s State) ReleaseVersion() (string, error) {
	switch s.Channel {
	case ChannelNone:
		return s.BaseVersion, nil
	case ChannelAlpha, ChannelBeta, ChannelRC:
		if s.Counter <= 0 {
			return "", errors.New("prerelease counter must be positive")
		}
		return fmt.Sprintf("%s-%s.%d", s.BaseVersion, s.Channel, s.Counter), nil
	case ChannelStable:
		return s.BaseVersion, nil
	default:
		return "", fmt.Errorf("unknown channel %q", s.Channel)
	}
}

func (s State) ReleaseTag() (string, error) {
	version, err := s.ReleaseVersion()
	if err != nil {
		return "", err
	}
	return "v" + version, nil
}

func (s State) WithInit(version string) State {
	return Seed(version)
}

func (s State) WithNew(version string) (State, error) {
	if s.InProgress() && s.Channel != ChannelStable {
		return State{}, errors.New("current release line is still in progress")
	}
	if version == "" {
		return State{}, errors.New("base version is required")
	}
	if !baseVersionPattern.MatchString(version) {
		return State{}, fmt.Errorf("invalid version %q", version)
	}
	if s.BaseVersion != "" && compareVersion(version, s.BaseVersion) <= 0 && s.State != StatePlanned {
		return State{}, fmt.Errorf("release version %s must move forward from %s", version, s.BaseVersion)
	}
	s.BaseVersion = version
	s.Channel = ChannelNone
	s.Counter = 0
	s.State = StateInProgress
	return s, nil
}

func (s State) Advance(channel Channel) (State, error) {
	if s.BaseVersion == "" {
		return State{}, errors.New("release line has not been initialized")
	}
	switch channel {
	case ChannelAlpha:
		return advancePrerelease(s, ChannelAlpha, []Channel{ChannelNone, ChannelAlpha})
	case ChannelBeta:
		return advancePrerelease(s, ChannelBeta, []Channel{ChannelAlpha, ChannelBeta})
	case ChannelRC:
		return advancePrerelease(s, ChannelRC, []Channel{ChannelBeta, ChannelRC})
	default:
		return State{}, fmt.Errorf("unsupported channel %q", channel)
	}
}

func (s State) Finalize() (State, error) {
	if s.BaseVersion == "" {
		return State{}, errors.New("release line has not been initialized")
	}
	if s.Channel != ChannelRC {
		return State{}, errors.New("final requires rc state")
	}
	s.Channel = ChannelStable
	s.Counter = 0
	s.State = StateReleased
	s.ReleasedVersion = s.BaseVersion
	tag, err := s.ReleaseTag()
	if err != nil {
		return State{}, err
	}
	s.LastTag = tag
	return s, nil
}

func advancePrerelease(s State, target Channel, allowed []Channel) (State, error) {
	if !containsChannel(allowed, s.Channel) {
		if s.State == StatePlanned && target == ChannelAlpha {
			s.Channel = ChannelAlpha
			s.Counter = 1
			s.State = StateInProgress
			return s, nil
		}
		return State{}, fmt.Errorf("%s requires previous prerelease stage", target)
	}
	switch target {
	case ChannelAlpha:
		if s.Channel == ChannelNone {
			s.Channel = ChannelAlpha
			s.Counter = 1
		} else {
			s.Counter++
		}
	case ChannelBeta:
		if s.Channel == ChannelAlpha {
			s.Channel = ChannelBeta
			s.Counter = 1
		} else {
			s.Counter++
		}
	case ChannelRC:
		if s.Channel == ChannelBeta {
			s.Channel = ChannelRC
			s.Counter = 1
		} else {
			s.Counter++
		}
	}
	s.State = StateInProgress
	return s, nil
}

func containsChannel(items []Channel, want Channel) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func compareVersion(a, b string) int {
	amaj, amin, apatch := parseBaseVersion(a)
	bmaj, bmin, bpatch := parseBaseVersion(b)
	switch {
	case amaj != bmaj:
		return compareInts(amaj, bmaj)
	case amin != bmin:
		return compareInts(amin, bmin)
	case apatch != bpatch:
		return compareInts(apatch, bpatch)
	default:
		return 0
	}
}

func parseBaseVersion(version string) (int, int, int) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return 0, 0, 0
	}
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])
	return major, minor, patch
}

func compareInts(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func StateFile(root string) string {
	return filepath.Join(root, "version.json")
}

func ChangelogFile(root string) string {
	return filepath.Join(root, "CHANGELOG.md")
}
