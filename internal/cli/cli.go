package cli

import (
	"fmt"

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
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printHelp() {
	fmt.Println("thisisckm - CLI tools for release workflows")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  thisisckm [--help] [--version]")
	fmt.Println("  thisisckm release <init|new|alpha|beta|rc|final> [version]")
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
	default:
		return fmt.Errorf("unknown release subcommand %q", args[0])
	}
}
