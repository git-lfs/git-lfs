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
	apiUrl   string
	cloneUrl string

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
		fmt.Printf("Creating test data (will modify server contents)\n")
		oidsExist, oidsMissing = constructTestOids()
		// Run a 'test' which is really just a setup task, but because it has to
		// use the same APIs it's a test in its own right too
		err := runTest(ServerTest{"Set up test data", setupTestData}, oidsExist, oidsMissing)
		if err != nil {
			exit("Failed to set up test data, aborting")
		}
	}

	runTests(oidsExist, oidsMissing)
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

func constructTestOids() (oidsExist, oidsMissing []TestObject) {
	const oidCount = 50
	oidsExist = make([]TestObject, 0, oidCount)
	oidsMissing = make([]TestObject, 0, oidCount)

	// Generate SHAs, not random so repeatable
	rand.Seed(int64(oidCount))
	runningSha := sha256.New()
	for i := 0; i < oidCount; i++ {
		runningSha.Write([]byte{byte(rand.Intn(256))})
		oid := hex.EncodeToString(runningSha.Sum(nil))
		sz := int64(rand.Intn(200)) + 50
		oidsExist = append(oidsExist, TestObject{Oid: oid, Size: sz})

		runningSha.Write([]byte{byte(rand.Intn(256))})
		oid = hex.EncodeToString(runningSha.Sum(nil))
		sz = int64(rand.Intn(200)) + 50
		oidsMissing = append(oidsMissing, TestObject{Oid: oid, Size: sz})
	}
	return
}

func runTests(oidsExist, oidsMissing []TestObject) {

	fmt.Printf("Running %d tests...\n", len(tests))
	for _, t := range tests {
		runTest(t, oidsExist, oidsMissing)
	}

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

func setupTestData(oidsExist, oidsMissing []TestObject) error {
	// TODO
	return nil
}

// Exit prints a formatted message and exits.
func exit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(2)
}

func init() {
	RootCmd.Flags().StringVarP(&apiUrl, "url", "u", "", "URL of the API (must supply this or --clone)")
	RootCmd.Flags().StringVarP(&cloneUrl, "clone", "c", "", "Clone URL from which to find API (must supply this or --url)")
}
