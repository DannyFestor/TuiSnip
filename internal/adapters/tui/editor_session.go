package tui

import "io"

type editorSession struct {
	run    EditorRun
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	edited string
}

func newEditorSession(run EditorRun) *editorSession {
	return &editorSession{run: run, stdin: nil, stdout: nil, stderr: nil, edited: ""}
}

func (s *editorSession) SetStdin(stdin io.Reader) {
	s.stdin = stdin
}

func (s *editorSession) SetStdout(stdout io.Writer) {
	s.stdout = stdout
}

func (s *editorSession) SetStderr(stderr io.Writer) {
	s.stderr = stderr
}

func (s *editorSession) Run() error {
	edited, err := s.run(s.stdin, s.stdout, s.stderr)
	s.edited = edited

	return err
}
