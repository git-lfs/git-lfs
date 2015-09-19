package main

import(
	"io/ioutil"
	"os"
	"strings"
)

func main(){
	var suffix string = ".1.ronn"
	var documentationRoot string ="docs/man"
	files, _ := ioutil.ReadDir(documentationRoot)
    out, _ := os.Create("commands/help-text.go")
    out.Write([]byte("package commands \n\nconst (\n"))

    for _, file := range files {
        if strings.HasSuffix(file.Name(), suffix) {
			var variableName string = strings.Replace(file.Name(), "-", "_", -1)
			variableName = strings.TrimSuffix(variableName, suffix)
			
            out.Write([]byte(variableName + "_HelpText = `"))
            out.Write(convertManFileToString(file.Name(), documentationRoot))
			out.Write([]byte("`\n"))
        }
    }
	out.Write([]byte(")\n"))
}


func convertManFileToString(fileName, root string) []byte {
    helpText, _ := ioutil.ReadFile(root + "/" + fileName)
	return []byte(strings.Replace(string(helpText), "`", "'", -1))

}


