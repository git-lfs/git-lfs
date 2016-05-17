package lfs

import "github.com/github/git-lfs/auth"

var (
	TestCredentialsFunc auth.CredentialFunc
	origCredentialsFunc auth.CredentialFunc
)

func init() {
	TestCredentialsFunc = func(input auth.Creds, subCommand string) (auth.Creds, error) {
		output := make(auth.Creds)
		for key, value := range input {
			output[key] = value
		}
		if _, ok := output["username"]; !ok {
			output["username"] = input["host"]
		}
		output["password"] = "monkey"
		return output, nil
	}
}

// Override the credentials func for testing
func SetupTestCredentialsFunc() {
	origCredentialsFunc = auth.SetCredentialsFunc(TestCredentialsFunc)
}

// Put the original credentials func back
func RestoreCredentialsFunc() {
	auth.SetCredentialsFunc(origCredentialsFunc)
}
