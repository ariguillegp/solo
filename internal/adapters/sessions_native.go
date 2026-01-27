package adapters

import (
	"os"
	"syscall"

	"github.com/ariguillegp/solo/internal/core"
)

type NativeSession struct{}

func NewNativeSession() *NativeSession {
	return &NativeSession{}
}

func (n *NativeSession) OpenSession(spec core.SessionSpec) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	if err := os.Chdir(spec.DirPath); err != nil {
		return err
	}

	return syscall.Exec(shell, []string{shell}, os.Environ())
}
