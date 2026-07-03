package skills

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed prepare-pr/** prepare-release-pr/**
var embeddedSkills embed.FS

type Agent string

const (
	AgentCodex  Agent = "codex"
	AgentClaude Agent = "claude"
	AgentCursor Agent = "cursor"

	preparePR        = "prepare-pr"
	prepareReleasePR = "prepare-release-pr"
	cursorRulesDir   = ".cursor/rules"
)

type Options struct {
	DryRun bool
	Force  bool
}

type Action struct {
	Skill       string
	Source      string
	Destination string
	Skipped     bool
}

type skillSpec struct {
	Name string
}

var skillManifest = []skillSpec{
	{Name: preparePR},
	{Name: prepareReleasePR},
}

func ParseAgent(value string) (Agent, error) {
	normalized := Agent(strings.ToLower(strings.TrimSpace(value)))
	switch normalized {
	case AgentCodex, AgentClaude, AgentCursor:
		return normalized, nil
	default:
		return "", fmt.Errorf("unknown agent %q; expected codex, claude, or cursor", value)
	}
}

func ValidAgents() []string {
	return []string{string(AgentCodex), string(AgentClaude), string(AgentCursor)}
}

func Install(agent Agent, opts Options) ([]Action, error) {
	plans, err := buildPlans(agent)
	if err != nil {
		return nil, err
	}

	actions := make([]Action, 0, len(plans))
	for _, plan := range plans {
		action := Action{
			Skill:       plan.Name,
			Source:      plan.Source,
			Destination: plan.Destination,
		}
		exists, err := pathExists(plan.Destination)
		if err != nil {
			return nil, err
		}
		if exists && !opts.Force {
			action.Skipped = true
			actions = append(actions, action)
			continue
		}
		actions = append(actions, action)
		if opts.DryRun {
			continue
		}
		if exists {
			if err := os.RemoveAll(plan.Destination); err != nil {
				return nil, err
			}
		}
		if err := installItem(plan.Source, plan.Destination); err != nil {
			return nil, err
		}
	}
	return actions, nil
}

type planItem struct {
	Name        string
	Source      string
	Destination string
}

func buildPlans(agent Agent) ([]planItem, error) {
	if len(skillManifest) == 0 {
		return nil, errors.New("no skills are configured")
	}
	switch agent {
	case AgentCodex, AgentClaude:
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		base := filepath.Join(home, "."+string(agent), "skills")
		plans := make([]planItem, 0, len(skillManifest))
		for _, spec := range skillManifest {
			plans = append(plans, planItem{
				Name:        spec.Name,
				Source:      filepath.Join(spec.Name),
				Destination: filepath.Join(base, spec.Name),
			})
		}
		return plans, nil
	case AgentCursor:
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		base := filepath.Join(home, cursorRulesDir)
		plans := make([]planItem, 0, len(skillManifest))
		for _, spec := range skillManifest {
			plans = append(plans, planItem{
				Name:        spec.Name,
				Source:      filepath.Join(spec.Name, "SKILL.md"),
				Destination: filepath.Join(base, "thisisckm-"+spec.Name+".mdc"),
			})
		}
		return plans, nil
	default:
		return nil, fmt.Errorf("unknown agent %q; expected codex, claude, or cursor", agent)
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func installItem(source, destination string) error {
	info, err := fs.Stat(embeddedSkills, source)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyEmbeddedDir(source, destination)
	}
	data, err := fs.ReadFile(embeddedSkills, source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destination, data, 0o644)
}

func copyEmbeddedDir(source, destination string) error {
	entries, err := fs.ReadDir(embeddedSkills, source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	for _, entry := range entries {
		src := filepath.Join(source, entry.Name())
		dst := filepath.Join(destination, entry.Name())
		if entry.IsDir() {
			if err := copyEmbeddedDir(src, dst); err != nil {
				return err
			}
			continue
		}
		data, err := fs.ReadFile(embeddedSkills, src)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
