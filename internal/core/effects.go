package core

type Effect interface {
	isEffect()
}

type EffScanDirs struct {
	Roots []string
}

func (EffScanDirs) isEffect() {}

type EffMkdirAll struct {
	Path string
}

func (EffMkdirAll) isEffect() {}

type EffOpenSession struct {
	Spec SessionSpec
}

func (EffOpenSession) isEffect() {}

type EffQuit struct{}

func (EffQuit) isEffect() {}
