package clipboard

const (
	sshTTYVariable        = "SSH_TTY"
	sshConnectionVariable = "SSH_CONNECTION"
)

func isRemoteSession(env environment) bool {
	return env.get(sshTTYVariable) != "" || env.get(sshConnectionVariable) != ""
}
