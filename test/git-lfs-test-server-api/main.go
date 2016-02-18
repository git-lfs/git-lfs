package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/github/git-lfs/test"

	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

type TestObject struct {
	Oid  string
	Size int64
}

type ServerTest struct {
	Name string
	F    func(oidsExist, oidsMissing []TestObject) error
}

var (
	RootCmd = &cobra.Command{
		Use:   "git-lfs-test-server-api [--url=<apiurl> | --clone=<cloneurl>] [<oid-exists-file> <oid-missing-file>]",
		Short: "Test a Git LFS API server for compliance",
		Run:   testServerApi,
	}
	apiUrl     string
	cloneUrl   string
	savePrefix string

	tests []ServerTest
)

func main() {
	RootCmd.Execute()
}

func testServerApi(cmd *cobra.Command, args []string) {

	if (len(apiUrl) == 0 && len(cloneUrl) == 0) ||
		(len(apiUrl) != 0 && len(cloneUrl) != 0) {
		exit("Must supply either --url or --clone (and not both)")
	}

	if len(args) != 0 && len(args) != 2 {
		exit("Must supply either no file arguments or both the exists AND missing file")
	}

	if len(args) != 0 && len(savePrefix) > 0 {
		exit("Cannot combine input files and --save option")
	}

	// Force loading of config before we alter it
	lfs.Config.AllGitConfig()

	// Configure the endpoint manually
	var endp lfs.Endpoint
	if len(cloneUrl) > 0 {
		endp = lfs.NewEndpointFromCloneURL(cloneUrl)
	} else {
		endp = lfs.NewEndpoint(apiUrl)
	}
	lfs.Config.SetManualEndpoint(endp)

	var oidsExist, oidsMissing []TestObject
	if len(args) >= 2 {
		fmt.Printf("Reading test data from files (no server content changes)\n")
		oidsExist = readTestOids(args[0])
		oidsMissing = readTestOids(args[1])
	} else {
		fmt.Printf("Creating test data (will upload to server)\n")
		var err error
		oidsExist, oidsMissing, err = buildTestData()
		if err != nil {
			exit("Failed to set up test data, aborting")
		}
		if len(savePrefix) > 0 {
			existFile := savePrefix + "_exists"
			missingFile := savePrefix + "_missing"
			saveTestOids(existFile, oidsExist)
			saveTestOids(missingFile, oidsMissing)
			fmt.Printf("Wrote test to %s, %s for future use\n", existFile, missingFile)
		}

	}

	ok := runTests(oidsExist, oidsMissing)
	if !ok {
		exit("One or more tests failed, see above")
	}
	fmt.Println("All tests passed")
}

func readTestOids(filename string) []TestObject {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		exit("Error opening file %s", filename)
	}
	defer f.Close()

	var ret []TestObject
	rdr := bufio.NewReader(f)
	line, err := rdr.ReadString('\n')
	for err == nil {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 2 {
			sz, _ := strconv.ParseInt(fields[1], 10, 64)
			ret = append(ret, TestObject{Oid: fields[0], Size: sz})
		}

		line, err = rdr.ReadString('\n')
	}

	return ret
}

type testDataCallback struct{}

func (*testDataCallback) Fatalf(format string, args ...interface{}) {
	exit(format, args...)
}
func (*testDataCallback) Errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func buildTestData() (oidsExist, oidsMissing []TestObject, err error) {
	const oidCount = 50
	oidsExist = make([]TestObject, 0, oidCount)
	oidsMissing = make([]TestObject, 0, oidCount)

	// Build test data for existing files & upload
	// Use test repo for this to simplify the process of making sure data matches oid
	// We're not performing a real test at this point (although an upload fail will break it)
	var callback testDataCallback
	repo := test.NewRepo(&callback)
	repo.Pushd()
	defer repo.Cleanup()
	// just one commit
	commit := test.CommitInput{CommitterName: "A N Other", CommitterEmail: "noone@somewhere.com"}
	var totalSize int64
	for i := 0; i < oidCount; i++ {
		filename := fmt.Sprintf("file%d.dat", i)
		sz := int64(rand.Intn(200)) + 50
		commit.Files = append(commit.Files, &test.FileInput{Filename: filename, Size: sz})
		totalSize += sz
	}
	outputs := repo.AddCommits([]*test.CommitInput{&commit})

	// create metadata
	metadata := lfs.NewTransferMetadata("refs/heads/master")
	// now upload
	uploadQueue := lfs.NewUploadQueue(len(oidsExist), totalSize, false, metadata)
	for _, f := range outputs[0].Files {
		oidsExist = append(oidsExist, TestObject{Oid: f.Oid, Size: f.Size})

		u, err := lfs.NewUploadable(f.Oid, "Test file")
		if err != nil {
			return nil, nil, err
		}
		uploadQueue.Add(u)
	}
	uploadQueue.Wait()

	for _, err := range uploadQueue.Errors() {
		if lfs.IsFatalError(err) {
			exit("Fatal error setting up test data: %s", err)
		}
	}

	// Generate SHAs for missing files, random but repeatable
	// No actual file content needed for these
	rand.Seed(int64(oidCount))
	runningSha := sha256.New()
	for i := 0; i < oidCount; i++ {
		runningSha.Write([]byte{byte(rand.Intn(256))})
		oid := hex.EncodeToString(runningSha.Sum(nil))
		sz := int64(rand.Intn(200)) + 50
		oidsMissing = append(oidsMissing, TestObject{Oid: oid, Size: sz})
	}
	return oidsExist, oidsMissing, nil
}

func saveTestOids(filename string, objs []TestObject) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		exit("Error opening file %s", filename)
	}
	defer f.Close()

	for _, o := range objs {
		f.WriteString(fmt.Sprintf("%s %d\n", o.Oid, o.Size))
	}

}

func runTests(oidsExist, oidsMissing []TestObject) bool {

	ok := true
	fmt.Printf("Running %d tests...\n", len(tests))
	for _, t := range tests {
		err := runTest(t, oidsExist, oidsMissing)
		if err != nil {
			ok = false
		}
	}
	return ok
}

func runTest(t ServerTest, oidsExist, oidsMissing []TestObject) error {
	const linelen = 70
	line := t.Name
	if len(line) > linelen {
		line = line[:linelen]
	} else if len(line) < linelen {
		line = fmt.Sprintf("%s%s", line, strings.Repeat(" ", linelen-len(line)))
	}
	fmt.Printf("%s...\r", line)

	err := t.F(oidsExist, oidsMissing)
	if err != nil {
		fmt.Printf("%s FAILED\n", line)
		fmt.Println(err.Error())
	} else {
		fmt.Printf("%s OK\n", line)
	}
	return err
}

// Exit prints a formatted message and exits.
func exit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(2)
}

func addTest(name string, f func(oidsExist, oidsMissing []TestObject) error) {
	tests = append(tests, ServerTest{Name: name, F: f})
}

func callBatchApi(op string, objs []TestObject) ([]*lfs.ObjectResource, error) {

	apiobjs := make([]*lfs.ObjectResource, 0, len(objs))
	for _, o := range objs {
		apiobjs = append(apiobjs, &lfs.ObjectResource{Oid: o.Oid, Size: o.Size})
	}
	metadata := lfs.NewTransferMetadata("refs/heads/master")
	return lfs.Batch(apiobjs, op, metadata)
}

// Combine 2 slices into one by "randomly" interleaving
// Not actually random, same sequence each time so repeatable
func interleaveTestData(slice1, slice2 []TestObject) []TestObject {
	// Predictable sequence, mixin existing & missing semi-randomly
	rand.Seed(21)
	count := len(slice1) + len(slice2)
	ret := make([]TestObject, 0, count)
	slice1Idx := 0
	slice2Idx := 0
	for left := count; left > 0; {
		for i := rand.Intn(3) + 1; slice1Idx < len(slice1) && i > 0; i-- {
			obj := slice1[slice1Idx]
			ret = append(ret, obj)
			slice1Idx++
			left--
		}
		for i := rand.Intn(3) + 1; slice2Idx < len(slice2) && i > 0; i-- {
			obj := slice2[slice2Idx]
			ret = append(ret, obj)
			slice2Idx++
			left--
		}
	}
	return ret
}

func init() {
	RootCmd.Flags().StringVarP(&apiUrl, "url", "u", "", "URL of the API (must supply this or --clone)")
	RootCmd.Flags().StringVarP(&cloneUrl, "clone", "c", "", "Clone URL from which to find API (must supply this or --url)")
	RootCmd.Flags().StringVarP(&savePrefix, "save", "s", "", "Saves generated data to <prefix>_exists|missing for subsequent use")
}
