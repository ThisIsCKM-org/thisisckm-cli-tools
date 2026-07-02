package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"thisisckm-cli-tools/internal/release"
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
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printHelp() {
	fmt.Println("thisisckm - CLI tools for release workflows")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  thisisckm [--help] [--version]")
	fmt.Println("  thisisckm release <init|config|new|alpha|beta|rc|final|sync-develop> [version]")
	fmt.Println("  thisisckm changelog <bug|feature|added|change|removed> -m <message>")
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
		return workspace.Advance(release.ChannelAlpha)
	case "beta":
		return workspace.Advance(release.ChannelBeta)
	case "rc":
		return workspace.Advance(release.ChannelRC)
	case "final":
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
