package main

import (
	"fmt"
	"os"
	"subgo"
	"time"

	"github.com/spf13/cobra"
)

var (
	outputFile  string
	shiftAmount time.Duration
	stretch     float64
	clamp       bool
	trimFirst   int
	trimLast    int
	trimBefore  time.Duration
	trimAfter   time.Duration
	removeHI    bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "subgo <input-file>",
	Short: "A tool for working with subtitle files",
	Long: `subgo is a CLI tool for processing subtitle files.
It can perform various operations on subtitle files and conversion between formats.

Examples:
  subgo input.srt --shift 300ms --output output.srt
  subgo input.srt --shift -1s --stretch 1.05
  subgo input.ass --output output.srt
  subgo input.srt --trim-before 1m --trim-after 1h49m30s
  subgo input.srt --trim-first 5 --trim-last 2`,
	Args: cobra.ExactArgs(1),
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "output.srt", "output subtitle file")
	rootCmd.Flags().DurationVarP(&shiftAmount, "shift", "s", 0, "shift amount (e.g., 300ms, 2s, -1s500ms)")
	rootCmd.Flags().Float64VarP(&stretch, "stretch", "t", 1.0, "stretch factor (e.g., 1.05 for 5% longer)")
	rootCmd.Flags().BoolVarP(&clamp, "clamp", "c", true, "clamp negative times to zero")
	rootCmd.Flags().IntVar(&trimFirst, "trim-first", 0, "remove first n events")
	rootCmd.Flags().IntVar(&trimLast, "trim-last", 0, "remove last n events")
	rootCmd.Flags().DurationVar(&trimBefore, "trim-before", 0, "remove events before timestamp (e.g., 1m, 1h30m)")
	rootCmd.Flags().DurationVar(&trimAfter, "trim-after", 0, "remove events after timestamp (e.g., 1h49m30s)")
	rootCmd.Flags().BoolVar(&removeHI, "remove-hi", false, "remove hearing impaired annotations like (sobbing) or [loud noise]")
}

func run(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Load subtitle file
	sub, err := subgo.Load(inputFile)
	if err != nil {
		return fmt.Errorf("load: %w", err)
	}

	// Apply operations
	sub = applyOperations(sub)

	// Save subtitle file
	if err := sub.Save(outputFile); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

func applyOperations(sub subgo.Subtitle) subgo.Subtitle {
	// Apply trimming first (before other operations)
	if trimFirst > 0 {
		sub = sub.TrimFirst(trimFirst)
	}
	if trimLast > 0 {
		sub = sub.TrimLast(trimLast)
	}
	if trimBefore > 0 {
		sub = sub.TrimBefore(trimBefore)
	}
	if trimAfter > 0 {
		sub = sub.TrimAfter(trimAfter)
	}

	// Apply shift if specified
	if shiftAmount != 0 {
		sub = sub.Shift(shiftAmount, clamp)
	}

	// Apply stretch if not 1.0
	if stretch != 1.0 {
		sub = sub.Stretch(stretch, 0)
	}

	// Remove HI annotations if requested
	if removeHI {
		sub = sub.RemoveHI()
	}

	return sub
}
