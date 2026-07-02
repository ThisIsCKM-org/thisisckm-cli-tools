package release

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const configFileName = "release.config.json"

type ReleaseConfig struct {
	Branches BranchConfig `json:"branches"`
}

type BranchConfig struct {
	Develop BranchName `json:"develop"`
	Main    BranchName `json:"main"`
}

type BranchName string

func (b *BranchName) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err == nil {
		*b = BranchName(strings.TrimSpace(value))
		return nil
	}
	var values []string
	if err := json.Unmarshal(data, &values); err == nil {
		for _, item := range values {
			item = strings.TrimSpace(item)
			if item != "" {
				*b = BranchName(item)
				return nil
			}
		}
		*b = ""
		return nil
	}
	return errors.New("branch name must be a string")
}

func defaultReleaseConfig() ReleaseConfig {
	return ReleaseConfig{
		Branches: BranchConfig{
			Develop: "develop",
			Main:    "main",
		},
	}
}

func ConfigFile(root string) string {
	return filepath.Join(root, configFileName)
}

func LoadConfig(root string) (ReleaseConfig, error) {
	path := ConfigFile(root)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultReleaseConfig(), nil
		}
		return ReleaseConfig{}, err
	}
	cfg := defaultReleaseConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ReleaseConfig{}, err
	}
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return ReleaseConfig{}, err
	}
	return cfg, nil
}

func EnsureConfig(root string) error {
	path := ConfigFile(root)
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return SaveConfig(path, defaultReleaseConfig())
}

func Configure(root string, in io.Reader, out io.Writer) error {
	current, err := LoadConfig(root)
	if err != nil {
		return err
	}
	cfg, err := PromptConfig(in, out, current)
	if err != nil {
		return err
	}
	return SaveConfig(ConfigFile(root), cfg)
}

func PromptConfig(in io.Reader, out io.Writer, current ReleaseConfig) (ReleaseConfig, error) {
	reader := bufio.NewReader(in)
	developDefault := current.PreferredBranch("develop")
	if developDefault == "" {
		developDefault = "develop"
	}
	mainDefault := current.PreferredBranch("main")
	if mainDefault == "" {
		mainDefault = "main"
	}
	develop, err := promptBranchName(reader, out, "What is the development branch?", developDefault)
	if err != nil {
		return ReleaseConfig{}, err
	}
	main, err := promptBranchName(reader, out, "What is the main branch?", mainDefault)
	if err != nil {
		return ReleaseConfig{}, err
	}
	cfg := ReleaseConfig{
		Branches: BranchConfig{
			Develop: BranchName(develop),
			Main:    BranchName(main),
		},
	}
	cfg.normalize()
	return cfg, cfg.Validate()
}

func promptBranchName(reader *bufio.Reader, out io.Writer, question, defaultValue string) (string, error) {
	if _, err := fmt.Fprintf(out, "%s [%s]: ", question, defaultValue); err != nil {
		return "", err
	}
	input, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	value := strings.TrimSpace(input)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func SaveConfig(path string, cfg ReleaseConfig) error {
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func (c ReleaseConfig) Validate() error {
	if strings.TrimSpace(string(c.Branches.Develop)) == "" {
		return errors.New("develop branch name is required")
	}
	if strings.TrimSpace(string(c.Branches.Main)) == "" {
		return errors.New("main branch name is required")
	}
	return nil
}

func (c ReleaseConfig) BranchName(role string) string {
	switch role {
	case "develop", "development", "dev":
		if v := strings.TrimSpace(string(c.Branches.Develop)); v != "" {
			return v
		}
		return "develop"
	case "main", "master":
		if v := strings.TrimSpace(string(c.Branches.Main)); v != "" {
			return v
		}
		return "main"
	default:
		return role
	}
}

func (c ReleaseConfig) PreferredBranch(role string) string {
	return c.BranchName(role)
}

func (c *ReleaseConfig) normalize() {
	c.Branches.Develop = BranchName(strings.TrimSpace(string(c.Branches.Develop)))
	c.Branches.Main = BranchName(strings.TrimSpace(string(c.Branches.Main)))
	if c.Branches.Develop == "" {
		c.Branches.Develop = "develop"
	}
	if c.Branches.Main == "" {
		c.Branches.Main = "main"
	}
}

func resolveBranchName(root, role string) string {
	cfg, err := LoadConfig(root)
	if err != nil {
		return role
	}
	return cfg.BranchName(role)
}

func branchCandidates(root, role string) []string {
	cfg, err := LoadConfig(root)
	if err != nil {
		return []string{role}
	}
	name := cfg.BranchName(role)
	if name == "" {
		return []string{role}
	}
	return []string{name}
}

func branchAvailable(root, branch string) bool {
	return localBranchExists(root, branch) || remoteBranchExists(root, "origin", branch)
}

func firstAvailableBranch(root string, candidates []string) string {
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if branchAvailable(root, candidate) {
			return candidate
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return ""
}

func branchNamesForRoles(root, source, target string) (string, string) {
	return resolveBranchName(root, source), resolveBranchName(root, target)
}
