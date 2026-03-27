package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"youtube-downloader/internal/config"
)

// View is the root render function required by the BubbleTea runtime.
func (m Model) View() string {
	content := m.renderScreen()

	if m.width == 0 {
		return content
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
		lipgloss.WithWhitespaceBackground(colorBgDark),
	)
}

// renderScreen dispatches to the correct screen renderer.
func (m Model) renderScreen() string {
	switch m.screen {
	case screenHome:
		return m.viewHome()
	case screenFormat:
		return m.viewFormat()
	case screenQuality:
		return m.viewQuality()
	case screenDownloading:
		return m.viewDownloading()
	case screenDone:
		return m.viewDone()
	case screenError:
		return m.viewError()
	case screenYtdlpMissing:
		return m.viewYtdlpMissing()
	}
	return ""
}

// ─── Screens ─────────────────────────────────────────────────────────────────

func (m Model) viewHome() string {
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPurple).
		Padding(0, 1).
		Render(m.urlInput.View())

	card := styleCard.Render(lipgloss.JoinVertical(lipgloss.Left,
		styleTitle.Render("🔗  URL do YouTube"),
		"",
		inputBox,
		"",
		hintBar("Enter", "continuar", "q", "sair"),
	))

	return lipgloss.JoinVertical(lipgloss.Center,
		renderLogo(),
		"",
		styleSubtitle.Render("Baixe vídeos e músicas do YouTube com facilidade"),
		"",
		card,
	)
}

func (m Model) viewFormat() string {
	card := styleCard.Render(lipgloss.JoinVertical(lipgloss.Left,
		styleMuted.Render("URL: "+truncate(m.urlInput.Value(), config.URLDisplayLen)),
		"",
		styleTitle.Render("Escolha o formato"),
		"",
		renderList(availableFormats, m.selectedFormat, func(f Format) string { return f.Label }),
		"",
		hintBar("↑↓", "navegar", "Enter", "selecionar", "Esc", "voltar"),
	))

	return lipgloss.JoinVertical(lipgloss.Center, renderLogo(), "", card)
}

func (m Model) viewQuality() string {
	quals := m.currentQualities()
	fmtName := strings.ToUpper(availableFormats[m.selectedFormat].Value)

	card := styleCard.Render(lipgloss.JoinVertical(lipgloss.Left,
		styleMuted.Render("URL: "+truncate(m.urlInput.Value(), config.URLDisplayLen)),
		"",
		styleTitle.Render(fmt.Sprintf("Qualidade do %s", fmtName)),
		"",
		renderList(quals, m.selectedQuality, func(q Quality) string { return q.Label }),
		"",
		hintBar("↑↓", "navegar", "Enter", "baixar", "Esc", "voltar"),
	))

	return lipgloss.JoinVertical(lipgloss.Center, renderLogo(), "", card)
}

func (m Model) viewDownloading() string {
	quals := m.currentQualities()
	fmtLabel := strings.ToUpper(availableFormats[m.selectedFormat].Value)
	qualLabel := quals[m.selectedQuality].Label

	info := styleBadge.Render(fmtLabel) + "  " + styleSubtitle.Render(qualLabel)

	logLines := m.logs
	if len(logLines) == 0 {
		logLines = []string{"Iniciando download..."}
	}

	rendered := make([]string, len(logLines))
	for i, l := range logLines {
		rendered[i] = styleMuted.Render(l)
	}

	card := styleCard.Render(lipgloss.JoinVertical(lipgloss.Left,
		styleMuted.Render(truncate(m.urlInput.Value(), config.URLTruncLen)),
		"",
		info,
		"",
		m.spinner.View()+"  "+styleTitle.Render("Baixando..."),
		"",
		styleLogBox.Width(config.LogBoxWidth).Render(strings.Join(rendered, "\n")),
	))

	return lipgloss.JoinVertical(lipgloss.Center, renderLogo(), "", card)
}

func (m Model) viewDone() string {
	var filePath string
	if m.downloadResult != nil {
		filePath = m.downloadResult.FilePath
	}

	card := styleCard.Render(lipgloss.JoinVertical(lipgloss.Left,
		styleSuccess.Render("✔  Download concluído!"),
		"",
		styleMuted.Render("Arquivo salvo em:\n"+filePath),
		"",
		hintBar("r", "baixar outro", "q", "sair"),
	))

	return lipgloss.JoinVertical(lipgloss.Center, renderLogo(), "", card)
}

func (m Model) viewError() string {
	var errMsg string
	if m.downloadResult != nil && m.downloadResult.Err != nil {
		errMsg = truncate(m.downloadResult.Err.Error(), 70)
	}

	rendered := make([]string, len(m.logs))
	for i, l := range m.logs {
		rendered[i] = styleMuted.Render(l)
	}

	card := styleCard.Render(lipgloss.JoinVertical(lipgloss.Left,
		styleError.Render("✖  Falha no download"),
		"",
		styleMuted.Render(errMsg),
		"",
		styleLogBox.Width(config.LogBoxWidth).Render(strings.Join(rendered, "\n")),
		"",
		hintBar("r", "tentar novamente", "q", "sair"),
	))

	return lipgloss.JoinVertical(lipgloss.Center, renderLogo(), "", card)
}

func (m Model) viewYtdlpMissing() string {
	steps := lipgloss.JoinVertical(lipgloss.Left,
		styleTitle.Render("Como instalar o yt-dlp:"),
		"",
		styleMuted.Render("  Windows (winget):"),
		styleKeyHint.Render("    winget install yt-dlp"),
		"",
		styleMuted.Render("  Windows (pip):"),
		styleKeyHint.Render("    pip install yt-dlp"),
		"",
		styleMuted.Render("  Ou baixe em:"),
		styleKeyHint.Render("    https://github.com/yt-dlp/yt-dlp/releases"),
		"",
		styleMuted.Render("  💡 ffmpeg recomendado para MP3 e merge de streams:"),
		styleKeyHint.Render("    winget install ffmpeg"),
	)

	card := styleCard.Render(lipgloss.JoinVertical(lipgloss.Left,
		styleWarning.Render("⚠  yt-dlp não encontrado"),
		"",
		steps,
		"",
		hintBar("r", "verificar novamente", "q", "sair"),
	))

	return lipgloss.JoinVertical(lipgloss.Center, renderLogo(), "", card)
}

// ─── Components ───────────────────────────────────────────────────────────────

// renderLogo returns the YOUTUBE DOWNLOADER ASCII-art header.
func renderLogo() string {
	lines := []string{
		lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render("╦ ╦╔═╗╦ ╦╔╦╗╦ ╦╔╗ ╔═╗"),
		lipgloss.NewStyle().Foreground(colorPink).Bold(true).Render("╚╦╝║ ║║ ║ ║ ║ ║╠╩╗║╣ "),
		lipgloss.NewStyle().Foreground(colorCyan).Bold(true).Render(" ╩ ╚═╝╚═╝ ╩ ╚═╝╚═╝╚═╝"),
	}
	logo := lipgloss.JoinVertical(lipgloss.Center, lines...)
	badge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0F0F1A")).
		Background(colorPurple).
		Bold(true).
		Padding(0, 1).
		Render(" DOWNLOADER ")
	return lipgloss.JoinVertical(lipgloss.Center, logo, badge)
}

// renderList renders a selectable list of items using a label extractor func.
func renderList[T any](items []T, selected int, label func(T) string) string {
	rows := make([]string, len(items))
	for i, item := range items {
		if i == selected {
			rows[i] = styleItemSelected.Render("▶  " + label(item))
		} else {
			rows[i] = styleItemNormal.Render("   " + label(item))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// hintBar renders a horizontal row of key hints from (key, label) pairs.
func hintBar(pairs ...string) string {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts,
			styleKeyHint.Render("["+pairs[i]+"]")+" "+styleKeyLabel.Render(pairs[i+1]),
		)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top,
		strings.Join(parts, "   "),
	)
}

// truncate shortens s to max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
