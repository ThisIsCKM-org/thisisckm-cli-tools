package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"thisisckm-cli-tools/internal/release"
	"thisisckm-cli-tools/skills"
)

const version = "0.1.0"

func Run(args []string) error {
	if len(args) == 0 {
		printHelp()
		return nil
	}

	switch args[0] {
	case "-h", "--help", "help":
		printHelp()
		return nil
	case "-v", "--version", "version":
		fmt.Println(version)
		return nil
	case "release":
		return runRelease(args[1:])
	case "changelog":
		return runChangelog(args[1:])
	case "add-skills":
		return runAddSkills(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printHelp() {
	fmt.Println("thisisckm - CLI tools for release workflows")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  thisisckm [--help] [--version]")
	fmt.Println("  thisisckm add-skills <codex|claude|cursor> [--dry-run] [--force]")
	fmt.Println("  thisisckm release <init|config|new|alpha|beta|rc|final|sync-develop> [version]")
	fmt.Println("  thisisckm changelog <bug|feature|added|change|removed> -m <message>")
}

func runAddSkills(args []string) error {
	agent, opts, err := parseAddSkillsArgs(args)
	if err != nil {
		return err
	}
	actions, err := skills.Install(agent, opts)
	if err != nil {
		return err
	}
	for _, action := range actions {
		if action.Skipped {
			fmt.Printf("skipped %s -> %s\n", action.Skill, action.Destination)
			continue
		}
		if opts.DryRun {
			fmt.Printf("would install %s -> %s\n", action.Skill, action.Destination)
			continue
		}
		fmt.Printf("installed %s -> %s\n", action.Skill, action.Destination)
	}
	return nil
}

func parseAddSkillsArgs(args []string) (skills.Agent, skills.Options, error) {
	var agent string
	opts := skills.Options{}
	for _, arg := range args {
		switch arg {
		case "--dry-run", "-n":
			opts.DryRun = true
		case "--force", "-f":
			opts.Force = true
		case "":
			continue
		default:
			if strings.HasPrefix(arg, "-") {
				return "", skills.Options{}, fmt.Errorf("unknown flag %q", arg)
			}
			if agent != "" {
				return "", skills.Options{}, fmt.Errorf("usage: thisisckm add-skills <codex|claude|cursor> [--dry-run] [--force]")
			}
			agent = arg
		}
	}
	if agent == "" {
		return "", skills.Options{}, fmt.Errorf("usage: thisisckm add-skills <codex|claude|cursor> [--dry-run] [--force]")
	}
	parsed, err := skills.ParseAgent(agent)
	if err != nil {
		return "", skills.Options{}, err
	}
	return parsed, opts, nil
}

func runRelease(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing release subcommand")
	}

	workspace, err := release.NewWorkspaceFromCWD()
	if err != nil {
		return err
	}

	switch args[0] {
	case "init":
		if len(args) != 2 {
			return fmt.Errorf("usage: thisisckm release init <version>")
		}
		return workspace.Init(args[1])
	case "config":
		if len(args) != 1 {
			return fmt.Errorf("usage: thisisckm release config")
		}
		return release.Configure(workspace.Root, os.Stdin, os.Stdout)
	case "new":
		if len(args) != 2 {
			return fmt.Errorf("usage: thisisckm release new <version>")
		}
		return workspace.New(args[1])
	case "alpha":
		if len(args) != 1 {
			return fmt.Errorf("usage: thisisckm release alpha")
		}
		return workspace.Advance(release.ChannelAlpha)
	case "beta":
		if len(args) != 1 {
			return fmt.Errorf("usage: thisisckm release beta")
		}
		return workspace.Advance(release.ChannelBeta)
	case "rc":
		if len(args) != 1 {
			return fmt.Errorf("usage: thisisckm release rc")
		}
		return workspace.Advance(release.ChannelRC)
	case "final":
		if len(args) != 1 {
			return fmt.Errorf("usage: thisisckm release final")
		}
		return workspace.Finalize()
	case "sync-develop":
		if len(args) != 1 {
			return fmt.Errorf("usage: thisisckm release sync-develop")
		}
		return workspace.SyncDevelop()
	default:
		return fmt.Errorf("unknown release subcommand %q", args[0])
	}
}

func runChangelog(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: thisisckm changelog <bug|feature|added|change|removed> -m <message>")
	}
	if len(args) < 3 || args[1] != "-m" {
		return fmt.Errorf("usage: thisisckm changelog <bug|feature|added|change|removed> -m <message>")
	}
	kind := strings.ToLower(args[0])
	message := strings.Join(args[2:], " ")
	workspace, err := release.NewWorkspaceFromCWD()
	if err != nil {
		return err
	}
	path, err := release.CreateStagedChangelogEntry(workspace.Root, kind, message)
	if err != nil {
		return err
	}
	fmt.Println(filepath.Base(path))
	return nil
}
