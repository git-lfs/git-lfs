package lfs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	v1Aliases = []string{
		"http://git-media.io/v/2",            // alpha
		"https://hawser.github.com/spec/v1",  // pre-release
		"https://git-lfs.github.com/spec/v1", // public launch
	}
	latest      = "https://git-lfs.github.com/spec/v1"
	oidType     = "sha256"
	oidRE       = regexp.MustCompile(`\A[[:alnum:]]{64}`)
	matcherRE   = regexp.MustCompile("git-media|hawser|git-lfs")
	extRE       = regexp.MustCompile(`\Aext-\d{1}-\w+`)
	pointerKeys = []string{"version", "oid", "size"}
)

type Pointer struct {
	Version    string
	Oid        string
	Size       int64
	OidType    string
	Extensions []*PointerExtension
}

// A PointerExtension is parsed from the Git LFS Pointer file.
type PointerExtension struct {
	Name     string
	Priority int
	Oid      string
	OidType  string
}

type ByPriority []*PointerExtension

func (p ByPriority) Len() int           { return len(p) }
func (p ByPriority) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ByPriority) Less(i, j int) bool { return p[i].Priority < p[j].Priority }

func NewPointer(oid string, size int64, exts []*PointerExtension) *Pointer {
	return &Pointer{latest, oid, size, oidType, exts}
}

func NewPointerExtension(name string, priority int, oid string) *PointerExtension {
	return &PointerExtension{name, priority, oid, oidType}
}

func (p *Pointer) Smudge(writer io.Writer, workingfile string, download bool, cb CopyCallback) error {
	return PointerSmudge(writer, p, workingfile, download, cb)
}

func (p *Pointer) Encode(writer io.Writer) (int, error) {
	return EncodePointer(writer, p)
}

func (p *Pointer) Encoded() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("version %s\n", latest))
	for _, ext := range p.Extensions {
		buffer.WriteString(fmt.Sprintf("ext-%d-%s %s:%s\n", ext.Priority, ext.Name, ext.OidType, ext.Oid))
	}
	buffer.WriteString(fmt.Sprintf("oid %s:%s\n", p.OidType, p.Oid))
	buffer.WriteString(fmt.Sprintf("size %d\n", p.Size))
	return buffer.String()
}

func EncodePointer(writer io.Writer, pointer *Pointer) (int, error) {
	return writer.Write([]byte(pointer.Encoded()))
}

func DecodePointerFromFile(file string) (*Pointer, error) {
	// Check size before reading
	stat, err := os.Stat(file)
	if err != nil {
		return nil, err
	}
	if stat.Size() > blobSizeCutoff {
		return nil, newNotAPointerError(nil)
	}
	f, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return DecodePointer(f)
}
func DecodePointer(reader io.Reader) (*Pointer, error) {
	_, p, err := DecodeFrom(reader)
	return p, err
}

func DecodeFrom(reader io.Reader) ([]byte, *Pointer, error) {
	buf := make([]byte, blobSizeCutoff)
	written, err := reader.Read(buf)
	output := buf[0:written]

	if err != nil {
		return output, nil, err
	}

	p, err := decodeKV(bytes.TrimSpace(output))
	return output, p, err
}

func verifyVersion(version string) error {
	if len(version) == 0 {
		return newNotAPointerError(errors.New("Missing version"))
	}

	for _, v := range v1Aliases {
		if v == version {
			return nil
		}
	}

	return errors.New("Invalid version: " + version)
}

func decodeKV(data []byte) (*Pointer, error) {
	kvps, exts, err := decodeKVData(data)
	if err != nil {
		if IsBadPointerKeyError(err) {
			badErr := err.(badPointerKeyError)
			if badErr.Expected == "version" {
				return nil, newNotAPointerError(err)
			}
		}
		return nil, err
	}

	if err := verifyVersion(kvps["version"]); err != nil {
		return nil, err
	}

	value, ok := kvps["oid"]
	if !ok {
		return nil, errors.New("Invalid Oid")
	}

	oid, err := parseOid(value)
	if err != nil {
		return nil, err
	}

	value, ok = kvps["size"]
	size, err := strconv.ParseInt(value, 10, 0)
	if err != nil || size < 0 {
		return nil, fmt.Errorf("Invalid size: %q", value)
	}

	var extensions []*PointerExtension
	if exts != nil {
		for key, value := range exts {
			ext, err := parsePointerExtension(key, value)
			if err != nil {
				return nil, err
			}
			extensions = append(extensions, ext)
		}
		if err = validatePointerExtensions(extensions); err != nil {
			return nil, err
		}
		sort.Sort(ByPriority(extensions))
	}

	return NewPointer(oid, size, extensions), nil
}

func parseOid(value string) (string, error) {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return "", errors.New("Invalid Oid value: " + value)
	}
	if parts[0] != oidType {
		return "", errors.New("Invalid Oid type: " + parts[0])
	}
	oid := parts[1]
	if !oidRE.Match([]byte(oid)) {
		return "", errors.New("Invalid Oid: " + oid)
	}
	return oid, nil
}

func parsePointerExtension(key string, value string) (*PointerExtension, error) {
	keyParts := strings.SplitN(key, "-", 3)
	if len(keyParts) != 3 || keyParts[0] != "ext" {
		return nil, errors.New("Invalid extension value: " + value)
	}

	p, err := strconv.Atoi(keyParts[1])
	if err != nil || p < 0 {
		return nil, errors.New("Invalid priority: " + keyParts[1])
	}

	name := keyParts[2]

	oid, err := parseOid(value)
	if err != nil {
		return nil, err
	}

	return NewPointerExtension(name, p, oid), nil
}

func validatePointerExtensions(exts []*PointerExtension) error {
	m := make(map[int]struct{})
	for _, ext := range exts {
		if _, exist := m[ext.Priority]; exist {
			return fmt.Errorf("Duplicate priority found: %d", ext.Priority)
		}
		m[ext.Priority] = struct{}{}
	}
	return nil
}

func decodeKVData(data []byte) (kvps map[string]string, exts map[string]string, err error) {
	kvps = make(map[string]string)

	if !matcherRE.Match(data) {
		err = newNotAPointerError(err)
		return
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	line := 0
	numKeys := len(pointerKeys)
	for scanner.Scan() {
		text := scanner.Text()
		if len(text) == 0 {
			continue
		}

		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			err = fmt.Errorf("Error reading line %d: %s", line, text)
			return
		}

		key := parts[0]
		value := parts[1]

		if numKeys <= line {
			err = fmt.Errorf("Extra line: %s", text)
			return
		}

		if expected := pointerKeys[line]; key != expected {
			if !extRE.Match([]byte(key)) {
				err = newBadPointerKeyError(expected, key)
				return
			}
			if exts == nil {
				exts = make(map[string]string)
			}
			exts[key] = value
			continue
		}

		line += 1
		kvps[key] = value
	}

	err = scanner.Err()
	return
}
