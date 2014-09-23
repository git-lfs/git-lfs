package git

type gitConfig struct {
}

var Config = &gitConfig{}

func (c *gitConfig) Find(val string) string {
	output, _ := simpleExec(nil, "git", "config", val)
	return output
}

func (c *gitConfig) SetGlobal(key, val string) {
	simpleExec(nil, "git", "config", "--global", "--add", key, val)
}

func (c *gitConfig) UnsetGlobal(key string) {
	simpleExec(nil, "git", "config", "--global", "--unset", key)
}

func (c *gitConfig) List() (string, error) {
	return simpleExec(nil, "git", "config", "-l")
}

func (c *gitConfig) ListFromFile() (string, error) {
	return simpleExec(nil, "git", "config", "-l", "-f", ".gitconfig")
}

func (c *gitConfig) Version() (string, error) {
	return simpleExec(nil, "git", "version")
}
