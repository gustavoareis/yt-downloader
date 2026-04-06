// Package downloader handles all yt-dlp subprocess interaction.
package downloader

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"youtube-downloader/internal/config"
)

// Options describes a download request.
type Options struct {
	URL     string
	Format  string // "mp3" or "mp4"
	Quality string // e.g. "320" (kbps) or "1080" (pixels)
}

// Result holds the outcome of a completed download.
type Result struct {
	FilePath string
	Err      error
}

// IsAvailable reports whether yt-dlp is present on the system PATH.
func IsAvailable() bool {
	_, err := exec.LookPath("yt-dlp")
	return err == nil
}

// Run performs a download as described by opts, streaming log lines into logChan.
// It returns a Result when the download is complete (success or failure).
func Run(opts Options, logChan chan<- string) *Result {
	outDir, err := prepareOutputDir()
	if err != nil {
		return &Result{Err: err}
	}

	args, err := buildArgs(opts, outDir)
	if err != nil {
		return &Result{Err: err}
	}

	logChan <- fmt.Sprintf("📁 Destino: %s", outDir)
	logChan <- "▶ Iniciando yt-dlp..."

	return execute(args, outDir, logChan)
}

// prepareOutputDir creates the downloads directory next to the executable and returns its absolute path.
func prepareOutputDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		exe = "."
	}
	dir := filepath.Join(filepath.Dir(exe), config.DownloadsDir)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("erro ao criar diretório de downloads: %w", err)
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir, nil
	}

	return abs, nil
}

// buildArgs dispatches to the format-specific argument builder.
func buildArgs(opts Options, outDir string) ([]string, error) {
	switch opts.Format {
	case "mp3":
		return buildMP3Args(opts.URL, opts.Quality, outDir), nil
	case "mp4":
		return buildMP4Args(opts.URL, opts.Quality, outDir), nil
	default:
		return nil, fmt.Errorf("formato não suportado: %q", opts.Format)
	}
}

func buildMP3Args(url, quality, outDir string) []string {
	return []string{
		"--extract-audio",
		"--audio-format", "mp3",
		"--audio-quality", quality + "K",
		"--output", filepath.Join(outDir, "%(title)s.%(ext)s"),
		"--no-playlist",
		"--progress",
		url,
	}
}

func buildMP4Args(url, quality, outDir string) []string {
	// Prefer pre-muxed mp4; fall back to merging best streams; fall back to any best.
	fmtSelector := fmt.Sprintf(
		"bestvideo[height<=%s][ext=mp4]+bestaudio[ext=m4a]/bestvideo[height<=%s]+bestaudio/best[height<=%s]",
		quality, quality, quality,
	)
	return []string{
		"--format", fmtSelector,
		"--merge-output-format", "mp4",
		"--output", filepath.Join(outDir, "%(title)s [%(height)sp].%(ext)s"),
		"--no-playlist",
		"--progress",
		url,
	}
}

// execute runs yt-dlp and streams stdout/stderr into logChan.
func execute(args []string, fallbackDir string, logChan chan<- string) *Result {
	cmd := exec.Command("yt-dlp", args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return &Result{Err: fmt.Errorf("falha ao iniciar yt-dlp: %w", err)}
	}

	var outputFile string
	pipesDone := make(chan struct{}, 2)

	go func() {
		streamLines(stdout, logChan, func(raw string) {
			if f := extractOutputFile(raw); f != "" {
				outputFile = f
			}
		})
		pipesDone <- struct{}{}
	}()

	go func() {
		streamLines(stderr, logChan, nil)
		pipesDone <- struct{}{}
	}()

	<-pipesDone
	<-pipesDone

	if err := cmd.Wait(); err != nil {
		return &Result{Err: fmt.Errorf("yt-dlp falhou: %w", err)}
	}

	if outputFile == "" {
		outputFile = fallbackDir
	}

	return &Result{FilePath: outputFile}
}

// streamLines reads lines from r, sanitizes each one, sends it to logChan,
// and optionally calls onLine with the raw line.
func streamLines(r io.Reader, logChan chan<- string, onLine func(string)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		raw := scanner.Text()
		if clean := sanitize(raw); clean != "" {
			logChan <- clean
		}
		if onLine != nil {
			onLine(raw)
		}
	}
}

// extractOutputFile attempts to parse the saved file path from a yt-dlp log line.
func extractOutputFile(line string) string {
	for _, marker := range []string{
		"[download] Destination:",
		"[ExtractAudio] Destination:",
	} {
		if strings.Contains(line, marker) {
			parts := strings.SplitN(line, "Destination:", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	if strings.Contains(line, "[Merger] Merging formats into") {
		parts := strings.SplitN(line, "into", 2)
		if len(parts) == 2 {
			return strings.Trim(strings.TrimSpace(parts[1]), `"`)
		}
	}

	return ""
}

// sanitize strips ANSI escape codes, carriage returns, and truncates long lines.
func sanitize(line string) string {
	out := strings.Builder{}
	inEscape := false

	for i := 0; i < len(line); i++ {
		switch {
		case line[i] == '\x1b':
			inEscape = true
		case inEscape && line[i] == 'm':
			inEscape = false
		case !inEscape:
			out.WriteByte(line[i])
		}
	}

	result := strings.ReplaceAll(out.String(), "\r", "")
	if len(result) > config.MaxLineLength {
		return result[:config.MaxLineLength] + "…"
	}

	return result
}
