package testapp

import (
	_ "embed"
)

type ClipboardTool string

const (
	RecordingTool ClipboardTool = "recording"
	FailingTool   ClipboardTool = "failing"
	NoTool        ClipboardTool = ""
)

var (
	//go:embed testdata/recording-tool.sh
	recordingScript []byte
	//go:embed testdata/reading-tool.sh
	readingScript []byte
	//go:embed testdata/failing-tool.sh
	failingScript []byte
)

func (c ClipboardTool) copyScript() []byte {
	switch c {
	case RecordingTool:
		return recordingScript
	case FailingTool:
		return failingScript
	case NoTool:
	}

	return nil
}

func (c ClipboardTool) readScript() []byte {
	switch c {
	case RecordingTool:
		return readingScript
	case FailingTool:
		return failingScript
	case NoTool:
	}

	return nil
}
