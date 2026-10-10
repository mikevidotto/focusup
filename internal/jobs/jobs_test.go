package jobs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleQueue = `# Apply Queue

| # | Done | Fit | Role | Company | Notes | URL | Outcome |
|---|------|-----|------|---------|-------|-----|---------|
| 1 | [x] | High | IT Analyst | CSIS | Closes 2026-10-30 | https://example.com/csis | applied 2026-10-07 |
| 2 | [ ] | High | Software Developer - Cloud Platform | IBM (Confluent) | Go CLI | https://example.com/ibm | |
| 3 | [ ] | Medium | Junior Developer (Finance Systems) | Harbor | Remote | https://example.com/harbor | |

Left out on purpose (low match): Nokia.`

const sampleTracker = `date,company,sector,role,role_type,channel,status,contact_person,fit_rating,notes,cv_file,cover_letter_file,source,deadline
2026-10-07,CSIS,Federal government,IT Analyst,IT analyst,online,applied,,71,Ref 1; applied 2026-10-07,cv/main_csis_it_analyst.tex,cover_letters/cover_csis_it_analyst.tex,https://example.com/csis,2026-10-30
2026-10-08,IBM (Confluent),Software / technology,Software Developer - Cloud Platform,software developer,portal,drafted,,72,"Go CLI, Terraform",cv/main_ibm_cloud.tex,cover_letters/cover_ibm_cloud.tex,https://example.com/ibm,
`

func newTestService(t *testing.T) (*Service, string) {
	t.Helper()

	dir := t.TempDir()
	write := func(rel, content string) {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write(queueFile, sampleQueue)
	write(trackerFile, sampleTracker)
	return NewService(dir), dir
}

func TestListQueue(t *testing.T) {
	s, _ := newTestService(t)

	items, err := s.ListQueue()
	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}

	if !items[0].Done || items[0].Outcome != "applied 2026-10-07" {
		t.Errorf("row 1 = %+v", items[0])
	}

	if items[1].Done || items[1].Company != "IBM (Confluent)" || items[1].URL != "https://example.com/ibm" {
		t.Errorf("row 2 = %+v", items[1])
	}
}

func TestSetQueueOutcomeChangesOneLine(t *testing.T) {
	s, dir := newTestService(t)

	if err := s.SetQueueOutcome(3, "skipped - too far | really"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, queueFile))
	before := strings.Split(sampleQueue, "\n")
	after := strings.Split(string(data), "\n")

	if len(before) != len(after) {
		t.Fatalf("line count changed: %d -> %d", len(before), len(after))
	}

	changed := 0
	for i := range before {
		if before[i] != after[i] {
			changed++
			want := "| 3 | [x] | Medium | Junior Developer (Finance Systems) | Harbor | Remote | https://example.com/harbor | skipped - too far / really |"
			if after[i] != want {
				t.Errorf("got  %q\nwant %q", after[i], want)
			}
		}
	}

	if changed != 1 {
		t.Errorf("%d lines changed, want 1", changed)
	}

	if err := s.SetQueueOutcome(99, "x"); err != ErrNotFound {
		t.Errorf("missing row: got %v, want ErrNotFound", err)
	}
}

func TestListApplications(t *testing.T) {
	s, _ := newTestService(t)

	apps, err := s.ListApplications()
	if err != nil {
		t.Fatal(err)
	}

	if len(apps) != 2 {
		t.Fatalf("got %d applications, want 2", len(apps))
	}

	ibm := apps[1]
	if ibm.Notes != "Go CLI, Terraform" || ibm.Status != "drafted" {
		t.Errorf("ibm = %+v", ibm)
	}
}

func TestMarkAppliedUpdatesTrackerAndQueue(t *testing.T) {
	s, dir := newTestService(t)

	if err := s.MarkApplied(1, "2026-10-09"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, trackerFile))
	before := strings.Split(sampleTracker, "\n")
	after := strings.Split(string(data), "\n")

	if len(before) != len(after) {
		t.Fatalf("line count changed: %d -> %d", len(before), len(after))
	}
	for i := range before {
		if i == 2 {
			continue
		}
		if before[i] != after[i] {
			t.Errorf("line %d changed: %q", i, after[i])
		}
	}

	want := `2026-10-08,IBM (Confluent),Software / technology,Software Developer - Cloud Platform,software developer,portal,applied,,72,"Go CLI, Terraform; applied 2026-10-09",cv/main_ibm_cloud.tex,cover_letters/cover_ibm_cloud.tex,https://example.com/ibm,`
	if after[2] != want {
		t.Errorf("got  %q\nwant %q", after[2], want)
	}

	items, _ := s.ListQueue()
	if !items[1].Done || items[1].Outcome != "applied 2026-10-09" {
		t.Errorf("queue row 2 = %+v", items[1])
	}
}

func TestPathsStayInsideRepo(t *testing.T) {
	s, dir := newTestService(t)

	for _, rel := range []string{"", "../x.csv", "/etc/passwd", "cv/../../x.csv"} {
		if _, err := s.resolve(rel); err != ErrOutsideDir {
			t.Errorf("resolve(%q) err = %v, want ErrOutsideDir", rel, err)
		}
	}

	got, err := s.resolve(trackerFile)
	if err != nil || got != filepath.Join(dir, trackerFile) {
		t.Errorf("resolve(tracker) = %q, %v", got, err)
	}
}
