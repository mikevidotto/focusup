package tasks

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	s, err := newServiceAt(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestService_CaptureAndClarify(t *testing.T) {
	s := newTestService(t)

	task, err := s.Add("  Call mom ", []string{"@Phone", "errands", "phone", " "})
	if err != nil {
		t.Fatal(err)
	}
	if task.Title != "Call mom" || task.List != ListInbox {
		t.Fatalf("got title %q list %q, want trimmed title in inbox", task.Title, task.List)
	}
	if want := []string{"errands", "phone"}; !slices.Equal(task.Contexts, want) {
		t.Fatalf("contexts = %v, want %v", task.Contexts, want)
	}

	if _, err := s.Add("  ", nil); err != ErrEmptyTitle {
		t.Fatalf("err = %v, want ErrEmptyTitle", err)
	}

	if task, err = s.Move(task.ID, ListSomeday); err != nil || task.List != ListSomeday {
		t.Fatalf("move: list = %q, err = %v", task.List, err)
	}
	if _, err := s.Move(task.ID, "later"); err != ErrInvalidList {
		t.Fatalf("err = %v, want ErrInvalidList", err)
	}

	if task, err = s.SetContexts(task.ID, nil); err != nil || len(task.Contexts) != 0 {
		t.Fatalf("set contexts: %v, err = %v", task.Contexts, err)
	}
	if _, err := s.Move("missing", ListNext); err != ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestService_Projects(t *testing.T) {
	s := newTestService(t)

	project, err := s.AddProject("Plan trip")
	if err != nil {
		t.Fatal(err)
	}

	inboxed, _ := s.Add("Book flights", nil)
	if inboxed, err = s.SetProject(inboxed.ID, project.ID); err != nil {
		t.Fatal(err)
	}
	if inboxed.ProjectID != project.ID || inboxed.List != ListNext {
		t.Fatalf("assigning a project should link it and move to next, got %+v", inboxed)
	}
	if _, err := s.SetProject(inboxed.ID, "missing"); err != ErrProjectNotFound {
		t.Fatalf("err = %v, want ErrProjectNotFound", err)
	}

	action, err := s.AddProjectTask(project.ID, "Renew passport", []string{"errands"})
	if err != nil {
		t.Fatal(err)
	}
	if action.List != ListNext || action.ProjectID != project.ID {
		t.Fatalf("project task = %+v, want next action in project", action)
	}
	if _, err := s.AddProjectTask("missing", "x", nil); err != ErrProjectNotFound {
		t.Fatalf("err = %v, want ErrProjectNotFound", err)
	}

	if project, err = s.ToggleProject(project.ID); err != nil || !project.Done || project.CompletedAt == nil {
		t.Fatalf("toggle project = %+v, err = %v", project, err)
	}

	if err := s.DeleteProject(project.ID); err != nil {
		t.Fatal(err)
	}
	if got := s.List(); len(got) != 2 || got[0].ProjectID != "" || got[1].ProjectID != "" {
		t.Fatalf("deleting a project should keep and unlink its tasks, got %+v", got)
	}
	if len(s.ListProjects()) != 0 {
		t.Fatal("project not deleted")
	}
}

func TestService_PersistsAcrossReload(t *testing.T) {
	dir := t.TempDir()
	s, err := newServiceAt(dir)
	if err != nil {
		t.Fatal(err)
	}

	project, _ := s.AddProject("Garden")
	task, _ := s.AddProjectTask(project.ID, "Buy seeds", []string{"errands"})
	s.Toggle(task.ID)

	reloaded, err := newServiceAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := reloaded.List()
	if len(got) != 1 || !got[0].Done || got[0].ProjectID != project.ID || !slices.Equal(got[0].Contexts, []string{"errands"}) {
		t.Fatalf("reloaded tasks = %+v", got)
	}
	if p := reloaded.ListProjects(); len(p) != 1 || p[0].Title != "Garden" {
		t.Fatalf("reloaded projects = %+v", p)
	}
}

func TestService_LoadsPreGTDTasksAsNextActions(t *testing.T) {
	dir := t.TempDir()
	legacy := `[{"id":"1","title":"Old todo","done":false,"priority":"high","createdAt":"2026-10-01T00:00:00Z","completedAt":null}]`
	if err := os.WriteFile(filepath.Join(dir, "tasks.json"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := newServiceAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := s.List()
	if len(got) != 1 || got[0].List != ListNext || got[0].Contexts == nil {
		t.Fatalf("legacy task = %+v, want next action with empty contexts", got)
	}
}
