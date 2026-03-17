package main

import "charm.land/lipgloss/v2"

var botStyle = lipgloss.NewStyle().
	Align(lipgloss.Left).
	Padding(0, 1).
	Background(lipgloss.Color("238")).
	Foreground(lipgloss.Color("255")).
	Padding(0, 1).
	Margin(0, 10, 0, 0)

var userStyle = lipgloss.NewStyle().
	Align(lipgloss.Right).
	Padding(0, 1).
	Background(lipgloss.Color("62")).
	Foreground(lipgloss.Color("230")).
	Padding(0, 1).
	Margin(0, 0, 0, 10)

var appStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(0, 1)

var headerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("230")).
	Background(lipgloss.Color("62")).
	Padding(0, 1)

var paneStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240"))

var composerStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")).
	Padding(0, 1)

var statusStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("244"))

var footerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241"))
