package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {
	fmt.Println("webhook started; listening on :8080")
	http.ListenAndServe(":8080", &WebhookHandler{})
}

type WebhookHandler struct {
	WebhookMessageParser
}

type WebhookMessage struct {
	Lang    string `json:"lang"`
	Code    string `json:"code"`
	branch  string
	project string
	path    string
}

func (m *WebhookMessage) clean() {
	m.Lang = strings.Trim(m.Lang, "\n ")
	m.Code = strings.Trim(m.Code, "\n ")
}

func (m *WebhookMessage) String() string {
	return fmt.Sprintf("language:%v\nproject:%v\npath:%v\nbranch:%v\ncode:%v", m.Lang, m.project, m.path, m.branch, len(m.Code))
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	msg := &WebhookMessage{}
	if err := json.Unmarshal(b, &msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	msg.clean()

	if err := h.Parse(msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(msg)
}

type WebhookMessageParser interface {
	Parse(*WebhookMessage) (map[string]interface{}, error)
}

func (h *WebhookHandler) Parse(msg *WebhookMessage) error {
	lines := strings.Split(msg.Code, "\n")
	if len(lines) >= 1 {
		sections := strings.Split(lines[0], "$$")
		if len(sections) >= 2 {
			fields := strings.Split(sections[1], ";")
			if len(fields) >= 3 {
				msg.project = fields[0]
				msg.path = fields[1]
				msg.branch = fields[2]
			}
		}
	}
	if len(msg.project)*len(msg.path)*len(msg.branch) == 0 {
		return fmt.Errorf("Unrecognized header")
	}
	return nil
}
