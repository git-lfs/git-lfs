package commands

import (
	"io/ioutil"
	"strings"
	"testing"
)

//New commands must be added here
var helpText = map[string]string {
	"git-lfs-checkout.1.ronn":git_lfs_checkout_HelpText,
    "git-lfs-clean.1.ronn":git_lfs_clean_HelpText,
	"git-lfs-env.1.ronn":git_lfs_env_HelpText,
	"git-lfs-fetch.1.ronn":git_lfs_fetch_HelpText,
	"git-lfs-fsck.1.ronn":git_lfs_fsck_HelpText,
	"git-lfs-init.1.ronn":git_lfs_init_HelpText,
	"git-lfs-logs.1.ronn":git_lfs_logs_HelpText,
	"git-lfs-ls-files.1.ronn":git_lfs_ls_files_HelpText,
	"git-lfs-pointer.1.ronn":git_lfs_pointer_HelpText,
	"git-lfs-pre-push.1.ronn":git_lfs_pre_push_HelpText,
	"git-lfs-pull.1.ronn":git_lfs_pull_HelpText,
	"git-lfs-push.1.ronn":git_lfs_push_HelpText,
	"git-lfs-smudge.1.ronn":git_lfs_smudge_HelpText,
	"git-lfs-status.1.ronn":git_lfs_status_HelpText,
	"git-lfs-track.1.ronn":git_lfs_track_HelpText,
	"git-lfs-uninit.1.ronn":git_lfs_uninit_HelpText,
	"git-lfs-untrack.1.ronn":git_lfs_untrack_HelpText,
	"git-lfs-update.1.ronn":git_lfs_update_HelpText,
	"git-lfs.1.ronn":git_lfs_HelpText,
}

var(
	suffix 		string 	= ".1.ronn"
)

func TestHelpTextMatchesManFiles(t *testing.T) {
		
	var suffix string = ".1.ronn"
	var documentationRoot string = "../docs/man"
	files, _ := ioutil.ReadDir(documentationRoot)
	
	
    for _, file := range files {		
        if strings.HasSuffix(file.Name(), suffix){
		
			var fileName = file.Name()			
		
			if _, ok := helpText[fileName]; ok != true {
				
				t.Errorf("Man file %s does not have corresponding help text.", fileName)
			}else{
				
				validateManFileSlice(t, documentationRoot, fileName, helpText[fileName])
			}
        }
    }
}

func validateManFileSlice(t *testing.T, root, fileName, helpText string){
	
	fileContent, err := ioutil.ReadFile(root + "/" + fileName)
	
	if(err != nil){
		t.Errorf(err.Error())
	}
	
	var fileText string = string(fileContent)
	
	fileText = strings.Trim(fileText, " ")

 	helpText = strings.Trim(helpText, " ")
	helpText = strings.Replace(helpText, "'", "`", -1)

	var fileLength int = len(fileText)
	var helpLength int = len(helpText)
	
	if(fileText[fileLength-10:] == helpText[helpLength - 10:]) {
		t.Errorf("Man file slice '%s' does not match help text slice '%s'.", fileText[fileLength-10:], helpText[helpLength - 10:])
	}
}