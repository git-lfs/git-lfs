package main

import(
	"os"
	"io/ioutil"
	"strings"
)

var(
	prefix 		string 	= "git-lfs-"
	suffix 		string 	= ".1.ronn"
	backQuote 	byte 	= byte("`"[0])
	singleQuote byte 	= byte("'"[0])
)

func main(){
	var documentationRoot string ="docs/man"
	
	files, _ := ioutil.ReadDir(documentationRoot)
    out, _ := os.Create("commands/help-text.go")
    out.Write([]byte("package commands \n\nconst (\n"))

    for _, file := range files {
		
        if strings.HasSuffix(file.Name(), suffix) {
			
			var newFileName string = strings.Replace(file.Name(), "-", "_", -1)
			newFileName = strings.TrimSuffix(newFileName, suffix)
			//newFileName = strings.ToUpper(newFileName)
			
            out.Write([]byte(newFileName + "_HelpText = `"))
			
            helpText, _ := ioutil.ReadFile(documentationRoot + "/" + file.Name())
		
			for i, runeVal := range helpText {
				if runeVal == backQuote {
					helpText[i] = singleQuote
				}
			}
		
            out.Write(helpText)
			out.Write([]byte("`\n"))
        }
    }
	out.Write([]byte(")\n"))
}


