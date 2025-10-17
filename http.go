package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
)

/*
This is the expected message format.
*/
type Message struct {
	Lang    string `json:"lang"`
	Code    string `json:"code"`
	Project string `json:"-"`
	File    string `json:"-"`
	Branch  string `json:"-"`
}

func (m *Message) clean() {
	m.Lang = strings.Trim(m.Lang, "\n ")
	m.Code = strings.Trim(m.Code, "\n ")
}

func (m *Message) parse() error {
	var (
		lines     = strings.Split(m.Code, "\n")
		start     = "webhook$$"
		end       = "$$"
		delimiter = ";"
		l         = lines[0]
		i         = strings.Index(l, start) + len(start)
		j         = strings.LastIndex(l, end)
	)
	if i == -1 || j == -1 || i > j {
		return nil
	}
	fields := strings.Split(l[i:j], delimiter)
	if len(fields) != 3 {
		return fmt.Errorf("bad header")
	}
	m.Project, m.File, m.Branch = fields[0], fields[1], fields[2]
	return nil
}

type Handler struct {
	Projects map[string]string // project name to project dir
	Messages chan *Message
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Body.Close()
	m := &Message{}
	if err := json.Unmarshal(b, &m); err != nil {
		fmt.Println(err)
		return
	}
	m.clean()
	if err := m.parse(); err != nil {
		fmt.Println(err)
		return
	}
	projectDir, ok := h.Projects[m.Project]
	if !ok {
		// log.Printf("no matching project: %v", m.project)
		return
	}
	if headBranch, err := getHeadBranch(projectDir); err != nil {
		log.Println("git error:", err)
		return
	} else if headBranch != m.Branch {
		log.Printf("head branch mismatch: expected %v got %v", m.Branch, headBranch)
		return
	}
	fullPath := path.Join(projectDir, m.File)
	dir := path.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Println("mkdir error:", err)
		return
	}
	if err := os.WriteFile(path.Join(projectDir, m.File), []byte(m.Code), 0644); err != nil {
		log.Println("write error:", err)
		return
	}
	h.Messages <- m
}

func getHeadBranch(projectDir string) (string, error) {
	repo, err := git.PlainOpen(projectDir)
	if err != nil {
		return "", err
	}
	head, err := repo.Head()
	if err != nil {
		return "", err
	}
	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}
	return "", nil
}
