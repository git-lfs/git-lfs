package lfs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmizerany/assert"
)

func TestSSHGetExeAndArgsSsh(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSH := os.Getenv("GIT_SSH")
	os.Setenv("GIT_SSH", "")
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsSshCustomPort(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSH := os.Getenv("GIT_SSH")
	os.Setenv("GIT_SSH", "")
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, "ssh", exe)
	assert.Equal(t, []string{"-p", "8888", "user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsPlink(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSH := os.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink.exe")
	os.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsPlinkCustomPort(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSH := os.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "plink")
	os.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-P", "8888", "user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsTortoisePlink(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	oldGITSSH := os.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink.exe")
	os.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

func TestSSHGetExeAndArgsTortoisePlinkCustomPort(t *testing.T) {
	endpoint := Config.Endpoint()
	endpoint.SshUserAndHost = "user@foo.com"
	endpoint.SshPort = "8888"
	oldGITSSH := os.Getenv("GIT_SSH")
	// this will run on non-Windows platforms too but no biggie
	plink := filepath.Join("Users", "joebloggs", "bin", "tortoiseplink")
	os.Setenv("GIT_SSH", plink)
	exe, args := sshGetExeAndArgs(endpoint)
	assert.Equal(t, plink, exe)
	assert.Equal(t, []string{"-batch", "-P", "8888", "user@foo.com"}, args)

	os.Setenv("GIT_SSH", oldGITSSH)
}

type TestStruct struct {
	Name      string
	Something int
}

func TestSSHEncodeJSONRequest(t *testing.T) {
	ctx := &SshApiContext{}

	params := &TestStruct{Name: "Fred", Something: 99}
	req, err := ctx.NewJsonRequest("TestMethod", params)
	assert.Equal(t, nil, err)
	reqbytes, err := json.Marshal(req)
	assert.Equal(t, nil, err)
	assert.Equal(t, `{"id":1,"method":"TestMethod","params":{"Name":"Fred","Something":99}}`, string(reqbytes))

}

func TestSSHDecodeJSONResponse(t *testing.T) {
	ctx := &SshApiContext{}
	inputstruct := TestStruct{Name: "Fred", Something: 99}
	resp, err := ctx.NewJsonResponse(1, inputstruct)
	assert.Equal(t, nil, err)
	outputstruct := TestStruct{}
	// Now unmarshal nested result; need to extract json first
	innerbytes, err := resp.Result.MarshalJSON()
	assert.Equal(t, nil, err)
	err = json.Unmarshal(innerbytes, &outputstruct)
	assert.Equal(t, inputstruct, outputstruct)
}
