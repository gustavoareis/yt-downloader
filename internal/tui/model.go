package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"youtube-downloader/internal/config"
	"youtube-downloader/internal/downloader"
)

// ─── Screen IDs ──────────────────────────────────────────────────────────────

type screen int

const (
	screenHome screen = iota
	screenFormat
	screenQuality
	screenDownloading
	screenDone
	screenError
	screenYtdlpMissing
)

// ─── Format / Quality descriptors ────────────────────────────────────────────

// Format holds the user-facing label and the value passed to yt-dlp.
type Format struct {
	Label string
	Value string
}

// Quality holds the user-facing label and the value passed to yt-dlp.
type Quality struct {
	Label string
	Value string
}

var availableFormats = []Format{
	{Label: "🎵  MP3  — Áudio (somente som)", Value: "mp3"},
	{Label: "🎬  MP4  — Vídeo (vídeo + áudio)", Value: "mp4"},
}

var mp3Qualities = []Quality{
	{Label: "320 kbps  — Máxima qualidade", Value: "320"},
	{Label: "256 kbps  — Alta qualidade", Value: "256"},
	{Label: "192 kbps  — Qualidade padrão", Value: "192"},
	{Label: "128 kbps  — Qualidade básica", Value: "128"},
}

var mp4Qualities = []Quality{
	{Label: "4K  (2160p)  — Ultra HD", Value: "2160"},
	{Label: "2K  (1440p)  — Quad HD", Value: "1440"},
	{Label: "FHD (1080p)  — Full HD", Value: "1080"},
	{Label: "HD  (720p)   — HD", Value: "720"},
	{Label: "SD  (480p)   — Padrão", Value: "480"},
	{Label: "LD  (360p)   — Baixa qualidade", Value: "360"},
}

// ─── Tea messages ────────────────────────────────────────────────────────────

type ytdlpAvailableMsg bool
type logLineMsg string
type downloadDoneMsg struct{ result *downloader.Result }

// ─── Model ───────────────────────────────────────────────────────────────────

// Model holds the complete TUI state.
type Model struct {
	screen          screen
	urlInput        textinput.Model
	selectedFormat  int
	selectedQuality int
	spinner         spinner.Model
	width           int
	height          int
	downloadResult  *downloader.Result
	logs            []string
	logChan         chan string
	doneChan        chan *downloader.Result
}

// NewModel creates and returns an initialized Model ready to run.
func NewModel() Model {
	input := textinput.New()
	input.Placeholder = "https://www.youtube.com/watch?v=..."
	input.Focus()
	input.CharLimit = 500
	input.Width = 60

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = styleSpinner

	return Model{
		screen:   screenHome,
		urlInput: input,
		spinner:  sp,
		logChan:  make(chan string, 100),
		doneChan: make(chan *downloader.Result, 1),
	}
}

// Init returns the initial set of commands to run when the program starts.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		cmdCheckYtdlp(),
	)
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case ytdlpAvailableMsg:
		if !bool(msg) {
			m.screen = screenYtdlpMissing
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case logLineMsg:
		m.appendLog(string(msg))
		return m, cmdWaitForLog(m.logChan)

	case downloadDoneMsg:
		m.downloadResult = msg.result
		if msg.result.Err != nil {
			m.screen = screenError
		} else {
			m.screen = screenDone
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if m.screen == screenHome {
		var cmd tea.Cmd
		m.urlInput, cmd = m.urlInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleKey routes key events to the appropriate screen handler.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenHome:
		return m.handleHomeKey(msg)
	case screenFormat:
		return m.handleFormatKey(msg)
	case screenQuality:
		return m.handleQualityKey(msg)
	case screenDone, screenError, screenYtdlpMissing:
		return m.handleResultKey(msg)
	case screenDownloading:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) handleHomeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "enter":
		if strings.TrimSpace(m.urlInput.Value()) != "" {
			m.screen = screenFormat
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.urlInput, cmd = m.urlInput.Update(msg)
	return m, cmd
}

func (m Model) handleFormatKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.screen = screenHome
	case "up", "k":
		if m.selectedFormat > 0 {
			m.selectedFormat--
		}
	case "down", "j":
		if m.selectedFormat < len(availableFormats)-1 {
			m.selectedFormat++
		}
	case "enter", " ":
		m.screen = screenQuality
		m.selectedQuality = 0
	}
	return m, nil
}

func (m Model) handleQualityKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	quals := m.currentQualities()
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.screen = screenFormat
	case "up", "k":
		if m.selectedQuality > 0 {
			m.selectedQuality--
		}
	case "down", "j":
		if m.selectedQuality < len(quals)-1 {
			m.selectedQuality++
		}
	case "enter", " ":
		m.screen = screenDownloading
		m.logs = nil
		opts := downloader.Options{
			URL:     strings.TrimSpace(m.urlInput.Value()),
			Format:  availableFormats[m.selectedFormat].Value,
			Quality: quals[m.selectedQuality].Value,
		}
		return m, tea.Batch(
			m.spinner.Tick,
			cmdDownload(opts, m.logChan, m.doneChan),
			cmdWaitForLog(m.logChan),
		)
	}
	return m, nil
}

func (m Model) handleResultKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r", "R":
		fresh := NewModel()
		fresh.width, fresh.height = m.width, m.height
		return fresh, tea.Batch(textinput.Blink, cmdCheckYtdlp())
	case "q", "Q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// currentQualities returns the quality list for the currently selected format.
func (m Model) currentQualities() []Quality {
	if availableFormats[m.selectedFormat].Value == "mp3" {
		return mp3Qualities
	}
	return mp4Qualities
}

// appendLog adds a line to the log buffer, evicting the oldest when full.
func (m *Model) appendLog(line string) {
	m.logs = append(m.logs, line)
	if len(m.logs) > config.MaxLogLines {
		m.logs = m.logs[len(m.logs)-config.MaxLogLines:]
	}
}

// ─── Tea Commands ─────────────────────────────────────────────────────────────

// cmdCheckYtdlp returns a command that checks whether yt-dlp is installed.
func cmdCheckYtdlp() tea.Cmd {
	return func() tea.Msg {
		return ytdlpAvailableMsg(downloader.IsAvailable())
	}
}

// cmdWaitForLog blocks on ch and returns the next log line as a logLineMsg.
func cmdWaitForLog(ch chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return nil
		}
		return logLineMsg(line)
	}
}

// cmdDownload starts yt-dlp in a goroutine and returns a downloadDoneMsg when finished.
func cmdDownload(opts downloader.Options, logChan chan string, doneChan chan *downloader.Result) tea.Cmd {
	return func() tea.Msg {
		go func() {
			doneChan <- downloader.Run(opts, logChan)
		}()
		return downloadDoneMsg{result: <-doneChan}
	}
}
