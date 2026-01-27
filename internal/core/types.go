package core

import "time"

type DirEntry struct {
	Path     string
	Name     string
	Score    int
	Exists   bool
	LastUsed time.Time
}

type SessionBackend string

const (
	BackendNative SessionBackend = "native"
	BackendTmux   SessionBackend = "tmux"
	BackendZellij SessionBackend = "zellij"
)

type SessionSpec struct {
	Backend   SessionBackend
	DirPath   string
	CreateDir bool
}
