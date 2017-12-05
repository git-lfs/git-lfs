package githistory

import (
	"bufio"
	"bytes"
	"strings"
	"text/template"

	"github.com/git-lfs/git-lfs/git/odb"
)

type MessageRewriter struct {
	t *template.Template
}

func NewMessageRewriter(t *template.Template) *MessageRewriter {
	return &MessageRewriter{
		t: t,
	}
}

type MessageTemplateArgument struct {
	SHA1 string

	Author       string
	Committer    string
	ExtraHeaders []*ExtraHeader
	Message      string
}

func (m *MessageRewriter) Rewrite(sha1 []byte, c *odb.Commit) (*odb.Commit, error) {
	if m == nil {
		return c, nil
	}

	var msg bytes.Buffer
	if err := m.t.Execute(&msg, m.argument(sha1, c)); err != nil {
		return nil, err
	}

	var parsingHeaders bool
	var headers []*odb.ExtraHeader
	var messages []string

	scanner := bufio.NewScanner(&msg)
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			parsingHeaders = false
			continue
		}

		if parsingHeaders {
			fields := strings.SplitN(scanner.Text(), " ", 2)
			headers = append(headers, &odb.ExtraHeader{
				K: fields[0],
				V: fields[1],
			})
		} else {
			messages = append(messages, scanner.Text())
		}
	}

	if err := scanner.Error(); err != nil {
		return nil, err
	}

	return &odb.Commit{
		Author:       c.Author,
		Committer:    c.Committer,
		ExtraHeaders: headers,
		Message:      strings.Join(messages, "\n"),

		ParentIDs: c.ParentIDs,
		TreeID:    c.TreeID,
	}, nil
}

func (m *MessageRewriter) argument(sha1 []byte, c *odb.Commit) {
	return &MessageTemplateArgument{
		SHA1: sha1,

		Author:       c.Author,
		Committer:    c.Committer,
		ExtraHeaders: c.ExtraHeaders,
		Message:      c.Message,
	}
}
