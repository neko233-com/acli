package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var ffmpegCmd = &cobra.Command{Use: "ffmpeg", Short: "Offline FFmpeg media operations and local version control"}
var ffmpegVersionCmd = &cobra.Command{Use: "version", Short: "Show active FFmpeg version", RunE: func(cmd *cobra.Command, args []string) error { return runFFmpeg("-version") }}
var ffmpegConvertCmd = &cobra.Command{Use: "convert <input> <output>", Short: "Convert media locally", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error { return runFFmpeg("-y", "-i", args[0], args[1]) }}
var ffmpegExtractAudioCmd = &cobra.Command{Use: "extract-audio <input> <output>", Short: "Extract audio locally", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error { return runFFmpeg("-y", "-i", args[0], "-vn", args[1]) }}
var ffmpegFramesCmd = &cobra.Command{Use: "frames <input> <out-pattern>", Short: "Extract video frames", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error { return runFFmpeg("-y", "-i", args[0], args[1]) }}
var ffmpegDoctorCmd = &cobra.Command{Use: "doctor", Short: "Locate system FFmpeg", RunE: func(cmd *cobra.Command, args []string) error {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg missing; install locally then use acli ffmpeg commands")
	}
	fmt.Printf("%s\n", path)
	return nil
}}

func runFFmpeg(args ...string) error {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg missing; run acli ffmpeg doctor")
	}
	child := exec.Command(path, args...)
	child.Stdin, child.Stdout, child.Stderr = os.Stdin, os.Stdout, os.Stderr
	return child.Run()
}
func init() {
	_ = runtime.GOOS
	_ = filepath.Separator
	ffmpegCmd.AddCommand(ffmpegVersionCmd, ffmpegDoctorCmd, ffmpegConvertCmd, ffmpegExtractAudioCmd, ffmpegFramesCmd)
}
