package notify

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type fakeClient struct {
	lastReq  *http.Request
	lastBody string
	status   int
	err      error
}

func (f *fakeClient) Do(req *http.Request) (*http.Response, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.lastReq = req
	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		f.lastBody = string(b)
	}
	status := f.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

func TestNotifier_Notify_Success(t *testing.T) {
	fc := &fakeClient{}
	n := Notifier{Topic: "my-topic", Client: fc}

	err := n.Notify("New relevant jobs", "You have 3 unviewed relevant postings")
	if err != nil {
		t.Fatalf("Notify: %v", err)
	}

	if fc.lastReq.URL.String() != "https://ntfy.sh/my-topic" {
		t.Fatalf("unexpected URL: %s", fc.lastReq.URL.String())
	}
	if fc.lastReq.Header.Get("Title") != "New relevant jobs" {
		t.Fatalf("unexpected Title header: %s", fc.lastReq.Header.Get("Title"))
	}
	if fc.lastBody != "You have 3 unviewed relevant postings" {
		t.Fatalf("unexpected body: %s", fc.lastBody)
	}
}

func TestNotifier_Notify_CustomBaseURL(t *testing.T) {
	fc := &fakeClient{}
	n := Notifier{BaseURL: "https://ntfy.example.com/", Topic: "my-topic", Client: fc}

	if err := n.Notify("t", "m"); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if fc.lastReq.URL.String() != "https://ntfy.example.com/my-topic" {
		t.Fatalf("unexpected URL: %s", fc.lastReq.URL.String())
	}
}

func TestNotifier_Notify_MissingTopic(t *testing.T) {
	n := Notifier{Client: &fakeClient{}}
	if err := n.Notify("t", "m"); err == nil {
		t.Fatalf("expected error for missing topic")
	}
}

func TestNotifier_Notify_ErrorStatus(t *testing.T) {
	fc := &fakeClient{status: http.StatusInternalServerError}
	n := Notifier{Topic: "my-topic", Client: fc}
	if err := n.Notify("t", "m"); err == nil {
		t.Fatalf("expected error for non-2xx status")
	}
}
