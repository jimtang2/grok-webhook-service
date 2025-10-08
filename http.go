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
	project string
	file    string
	branch  string
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
	m.project, m.file, m.branch = fields[0], fields[1], fields[2]
	return nil
}

type Handler struct {
	Projects map[string]string // project name to project dir
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("webhook request received")
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
	projectDir, ok := h.Projects[m.project]
	if !ok {
		log.Printf("no matching project: %v", m.project)
		return
	}
	if headBranch, err := getHeadBranch(projectDir); err != nil {
		log.Println("git error:", err)
		return
	} else if headBranch != m.branch {
		log.Printf("head branch mismatch: expected %v got %v", m.branch, headBranch)
		return
	}
	if err := os.WriteFile(path.Join(projectDir, m.file), []byte(m.Code), 0644); err != nil {
		log.Println("write error:", err)
		return
	}
	log.Printf("project:  %v", m.project)
	log.Printf("branch:   %v", m.branch)
	log.Printf("file:     %v", m.file)
	log.Printf("lang:     %v", m.Lang)
	log.Printf("size:     %v", len(m.Code))
	log.Printf("status:   ok")
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
