package lfs

import (
	"bufio"
	"strings"
	"sync"
)

// ScanIndex returns a slice of WrappedPointer objects for all Git LFS pointers
// it finds in the index.
//
// Ref is the ref at which to scan, which may be "HEAD" if there is at least one
// commit.
func scanIndex(cb GitScannerFoundPointer, ref string) error {
	indexMap := &indexFileMap{
		nameMap:      make(map[string][]*indexFile),
		nameShaPairs: make(map[string]bool),
		mutex:        &sync.Mutex{},
	}

	revs, err := revListIndex(ref, false, indexMap)
	if err != nil {
		return err
	}

	cachedRevs, err := revListIndex(ref, true, indexMap)
	if err != nil {
		return err
	}

	allRevsErr := make(chan error, 5) // can be multiple errors below
	allRevsChan := make(chan string, 1)
	allRevs := NewStringChannelWrapper(allRevsChan, allRevsErr)
	go func() {
		seenRevs := make(map[string]bool, 0)

		for rev := range revs.Results {
			if !seenRevs[rev] {
				allRevsChan <- rev
				seenRevs[rev] = true
			}
		}
		err := revs.Wait()
		if err != nil {
			allRevsErr <- err
		}

		for rev := range cachedRevs.Results {
			if !seenRevs[rev] {
				allRevsChan <- rev
				seenRevs[rev] = true
			}
		}
		err = cachedRevs.Wait()
		if err != nil {
			allRevsErr <- err
		}
		close(allRevsChan)
		close(allRevsErr)
	}()

	smallShas, err := catFileBatchCheck(allRevs)
	if err != nil {
		return err
	}

	ch := make(chan gitscannerResult, chanBufSize)

	barePointerCh, err := catFileBatch(smallShas)
	if err != nil {
		return err
	}

	go func() {
		for p := range barePointerCh.Results {
			for _, file := range indexMap.FilesFor(p.Sha1) {
				// Append a new *WrappedPointer that combines the data
				// from the index file, and the pointer "p".
				ch <- gitscannerResult{
					Pointer: &WrappedPointer{
						Sha1:    p.Sha1,
						Name:    file.Name,
						SrcName: file.SrcName,
						Status:  file.Status,
						Pointer: p.Pointer,
					},
				}
			}
		}

		if err := barePointerCh.Wait(); err != nil {
			ch <- gitscannerResult{Err: err}
		}

		close(ch)
	}()

	for result := range ch {
		cb(result.Pointer, result.Err)
	}

	return nil
}

// revListIndex uses git diff-index to return the list of object sha1s
// for in the indexf. It returns a channel from which sha1 strings can be read.
// The namMap will be filled indexFile pointers mapping sha1s to indexFiles.
func revListIndex(atRef string, cache bool, indexMap *indexFileMap) (*StringChannelWrapper, error) {
	cmdArgs := []string{"diff-index", "-M"}
	if cache {
		cmdArgs = append(cmdArgs, "--cached")
	}
	cmdArgs = append(cmdArgs, atRef)

	cmd, err := startCommand("git", cmdArgs...)
	if err != nil {
		return nil, err
	}

	cmd.Stdin.Close()

	revs := make(chan string, chanBufSize)
	errchan := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(cmd.Stdout)
		for scanner.Scan() {
			// Format is:
			// :100644 100644 c5b3d83a7542255ec7856487baa5e83d65b1624c 9e82ac1b514be060945392291b5b3108c22f6fe3 M foo.gif
			// :<old mode> <new mode> <old sha1> <new sha1> <status>\t<file name>[\t<file name>]
			line := scanner.Text()
			parts := strings.Split(line, "\t")
			if len(parts) < 2 {
				continue
			}

			description := strings.Split(parts[0], " ")
			files := parts[1:len(parts)]

			if len(description) >= 5 {
				status := description[4][0:1]
				sha1 := description[3]
				if status == "M" {
					sha1 = description[2] // This one is modified but not added
				}

				indexMap.Add(sha1, &indexFile{
					Name:    files[len(files)-1],
					SrcName: files[0],
					Status:  status,
				})
				revs <- sha1
			}
		}

		// Note: deliberately not checking result code here, because doing that
		// can fail fsck process too early since clean filter will detect errors
		// and set this to non-zero. How to cope with this better?
		// stderr, _ := ioutil.ReadAll(cmd.Stderr)
		// err := cmd.Wait()
		// if err != nil {
		// 	errchan <- fmt.Errorf("Error in git diff-index: %v %v", err, string(stderr))
		// }
		cmd.Wait()
		close(revs)
		close(errchan)
	}()

	return NewStringChannelWrapper(revs, errchan), nil
}

// indexFile is used when scanning the index. It stores the name of
// the file, the status of the file in the index, and, in the case of
// a moved or copied file, the original name of the file.
type indexFile struct {
	Name    string
	SrcName string
	Status  string
}

type indexFileMap struct {
	// mutex guards nameMap and nameShaPairs
	mutex *sync.Mutex
	// nameMap maps SHA1s to a slice of `*indexFile`s
	nameMap map[string][]*indexFile
	// nameShaPairs maps "sha1:name" -> bool
	nameShaPairs map[string]bool
}

// FilesFor returns all `*indexFile`s that match the given `sha`.
func (m *indexFileMap) FilesFor(sha string) []*indexFile {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.nameMap[sha]
}

// Add appends unique index files to the given SHA, "sha". A file is considered
// unique if its combination of SHA and current filename have not yet been seen
// by this instance "m" of *indexFileMap.
func (m *indexFileMap) Add(sha string, index *indexFile) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	pairKey := strings.Join([]string{sha, index.Name}, ":")
	if m.nameShaPairs[pairKey] {
		return
	}

	m.nameMap[sha] = append(m.nameMap[sha], index)
	m.nameShaPairs[pairKey] = true
}
