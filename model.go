package main

import (
	"log"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ChatMessage struct {
	Content string
	Role    MessageRole
}
type chatResponseMsg struct {
	message ChatMessage
}
type chatErrorMsg struct {
	err error
}

type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

type ChatRequest struct {
	Messages []ChatMessage
}

type ChatResponse struct {
	Response string
}

type Model struct {
	spinner          spinner.Model
	viewport         viewport.Model
	textinput        textinput.Model
	messages         []ChatMessage
	input            string
	pending          bool
	err              error
	width            int
	height           int
	cursor           int
	client           ChatClient
	chatClientConfig ChatClientConfig
	chatrequest      ChatRequest
	chatresponse     ChatResponse
}

type layoutSections struct {
	header   string
	status   string
	composer string
	footer   string
}

func newModel(config ChatClientConfig) Model {
	vp := viewport.New(
		viewport.WithWidth(80),
		viewport.WithHeight(20),
	)

	s := spinner.New()
	s.Spinner = spinner.Points

	ti := textinput.New()
	ti.Placeholder = "Ask anything"
	ti.SetVirtualCursor(false)
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(20)

	return Model{
		spinner:          s,
		viewport:         vp,
		textinput:        ti,
		pending:          false,
		messages:         []ChatMessage{},
		chatClientConfig: config,
		client:           newChatClient(config),
	}
}

func (m Model) renderMessages() string {
	log.Println("renderMessages().enter")
	defer log.Println("renderMessages().exit")
	var renderedResult []string

	appInnerWidth := m.width - appStyle.GetHorizontalFrameSize()
	if appInnerWidth < 1 {
		appInnerWidth = 1
	}

	for _, msg := range m.messages {
		var rendered string

		if msg.Role == MessageRoleUser {
			rendered = userStyle.
				Width(appInnerWidth).
				Render(msg.Content)
		} else {
			rendered = botStyle.
				Width(appInnerWidth).
				Render(msg.Content)
		}

		renderedResult = append(renderedResult, rendered)
	}

	content := strings.Join(func() []string {

		var result []string
		for _, m := range renderedResult {
			result = append(result, m)
		}

		return result
	}(), "\n")

	return content
}

func sendMessages(m Model) tea.Cmd {
	log.Println("sendMessages().enter")
	defer log.Println("sendMessages().exit")

	return func() tea.Msg {

		request := ChatRequest{
			Messages: m.messages,
		}
		m.chatrequest = request
		response, err := m.client.SendMessage(request)
		m.chatresponse = response
		if err != nil {
			errorMessage := chatErrorMsg{err: err}
			log.Println("m.sendMessages() - error sending message: " + err.Error())
			return errorMessage
		}
		chatMessage := chatResponseMsg{
			message: ChatMessage{
				Content: response.Response,
				Role:    MessageRoleAssistant,
			},
		}
		log.Println("m.sendMessages() - received message: " + chatMessage.message.Content)
		return chatMessage
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:

		m.width = msg.Width
		m.height = msg.Height

		appInnerWidth := m.width - appStyle.GetHorizontalFrameSize()
		appInnerHeight := m.height - appStyle.GetVerticalFrameSize()

		if appInnerWidth < 1 {
			appInnerWidth = 1
		}
		if appInnerHeight < 1 {
			appInnerHeight = 1
		}

		sectionWidth := appInnerWidth

		sections := m.renderLayoutSections(sectionWidth)
		headerHeight := lipgloss.Height(sections.header)
		statusHeight := lipgloss.Height(sections.status)
		footerHeight := lipgloss.Height(sections.footer)
		composerHeight := lipgloss.Height(sections.composer)

		paneTotalHeight := appInnerHeight - headerHeight - statusHeight - composerHeight - footerHeight
		if paneTotalHeight < 1 {
			paneTotalHeight = 1
		}

		paneInnerWidth := sectionWidth - paneStyle.GetHorizontalFrameSize()
		paneInnerHeight := paneTotalHeight - paneStyle.GetVerticalFrameSize()
		composerInnerWidth := sectionWidth - composerStyle.GetHorizontalFrameSize()

		if paneInnerWidth < 1 {
			paneInnerWidth = 1
		}
		if paneInnerHeight < 1 {
			paneInnerHeight = 1
		}
		if composerInnerWidth < 1 {
			composerInnerWidth = 1
		}

		m.viewport.SetWidth(paneInnerWidth)
		m.viewport.SetHeight(paneInnerHeight)
		m.textinput.SetWidth(composerInnerWidth)

	case tea.KeyPressMsg:

		switch msg.String() {

		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			if m.pending {
				return m, nil
			}

			log.Println("Update().msg.enter")
			if m.textinput.Value() == "" {
				return m, nil
			}
			m.pending = true

			msg := ChatMessage{
				Content: m.textinput.Value(),
				Role:    MessageRoleUser,
			}
			log.Println("Update().msg.enter - added user message: " + msg.Content)
			m.messages = append(m.messages, msg)

			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			m.textinput.SetValue("")

			return m, tea.Batch(
				m.spinner.Tick,
				sendMessages(m),
			)
		}

	case chatErrorMsg:
		log.Println("Update().msg.chatErrorMsg.Content: " + msg.err.Error())
		m.pending = false
		return m, nil

	case chatResponseMsg:
		log.Println("Update().msg.chatResponseMsg.Content: " + msg.message.Content)
		m.pending = false
		m.messages = append(m.messages, msg.message)
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		log.Println("Update().msg.chatResponseMsg message added")
		return m, nil

	case spinner.TickMsg:
		if m.pending {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	}
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m Model) View() tea.View {
	appInnerWidth := m.width - appStyle.GetHorizontalFrameSize()
	if appInnerWidth < 1 {
		appInnerWidth = 1
	}

	sections := m.renderLayoutSections(appInnerWidth)
	pane := m.renderPane(appInnerWidth)

	var c *tea.Cursor
	if !m.textinput.VirtualCursor() {
		c = m.textinput.Cursor()

		aboveComposer := lipgloss.Height(
			lipgloss.JoinVertical(
				lipgloss.Left,
				sections.header,
				pane,
				sections.status,
			),
		)

		c.Y += 1 + aboveComposer + 1
		c.X += 4
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		sections.header,
		pane,
		sections.status,
		sections.composer,
		sections.footer,
	)

	str := appStyle.Render(content)
	v := tea.NewView(str)
	v.Cursor = c
	return v
}

func (m Model) renderHeader(width int) string {
	return headerStyle.Render(m.chatClientConfig.Model)
}

func (m Model) renderStatus(width int) string {
	statusText := ""
	if m.pending {
		statusText = "Thinking " + m.spinner.View()
	}
	status := statusStyle.Render(statusText)
	return status
}

func (m Model) renderPane(width int) string {
	return paneStyle.Render(m.viewport.View())
}

func (m Model) renderComposer(width int) string {
	return composerStyle.Width(width).Render(m.textinput.View())
}

func (m Model) renderFooter(width int) string {
	return footerStyle.Render("Enter send | Ctrl+C exit")
}

func (m Model) renderLayoutSections(width int) layoutSections {
	return layoutSections{
		header:   m.renderHeader(width),
		status:   m.renderStatus(width),
		footer:   m.renderFooter(width),
		composer: m.renderComposer(width),
	}
}
