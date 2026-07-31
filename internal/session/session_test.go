package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewAndSaveLoad(t *testing.T) {
	s := New("groq", "llama-3.3-70b")
	if s.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	s.AddMessage("system", "hello", nil)
	s.AddMessage("assistant", "hi there", nil)

	d, err := sessionDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filepath.Join(d, s.ID+".json"))

	if err := Save(s); err != nil {
		t.Fatal(err)
	}

	s2, err := Load(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if s2.Provider != s.Provider {
		t.Fatalf("provider mismatch: %s vs %s", s2.Provider, s.Provider)
	}
	if len(s2.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(s2.Messages))
	}
}

func TestList(t *testing.T) {
	s1 := New("groq", "m1")
	s2 := New("openai", "m2")

	defer func() {
		Delete(s1.ID)
		Delete(s2.ID)
	}()

	if err := Save(s1); err != nil {
		t.Fatal(err)
	}
	if err := Save(s2); err != nil {
		t.Fatal(err)
	}

	sessions, err := List()
	if err != nil {
		t.Fatal(err)
	}

	found1 := false
	found2 := false
	for _, s := range sessions {
		if s.ID == s1.ID {
			found1 = true
		}
		if s.ID == s2.ID {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Fatalf("expected s1=%v and s2=%v both in list, got %d sessions", found1, found2, len(sessions))
	}
}

func TestToProviderMessages(t *testing.T) {
	s := New("groq", "m")
	s.AddMessage("system", "hello", nil)
	s.AddMessage("assistant", "reply", nil)

	msgs := s.ToProviderMessages()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "system" || msgs[0].Content != "hello" {
		t.Fatal("first message mismatch")
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	s := New("openai", "gpt-4o")
	s.AddMessage("role", "hello", nil)

	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}

	var s2 Session
	if err := json.Unmarshal(b, &s2); err != nil {
		t.Fatal(err)
	}
	if s2.ID != s.ID {
		t.Fatal("id mismatch")
	}
}

func TestDelete(t *testing.T) {
	s := New("groq", "m1")
	if err := Save(s); err != nil {
		t.Fatal(err)
	}
	if err := Delete(s.ID); err != nil {
		t.Fatal(err)
	}
	_, err := Load(s.ID)
	if err == nil {
		t.Fatal("expected error loading deleted session")
	}
}