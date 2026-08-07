//go:build integration

package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
)

const (
	kaprodiEmail     = "kaprodi@filkom.unida.ac.id"
	supervisorEmail  = "supervisor@filkom.unida.ac.id"
	examiner1Email   = "examiner1@filkom.unida.ac.id"
	examiner2Email   = "examiner2@filkom.unida.ac.id"
	sharedPassword   = "Password@123"
	thesisTitle      = "Sistem Informasi Manajemen Tugas Akhir Berbasis Web Untuk Fakultas Ilmu Komputer"
	lifecycleRoom    = "Ruang Sidang A"
	lifecycleRoom2   = "Ruang Sidang B"
)

// fullScoreRequest builds a passing score set for one examiner: all four
// components with generous marks so the weighted final score clears the pass
// threshold (seminar 60, defense 60–74 → passed/revision paths avoided).
func fullScoreRequest() map[string]any {
	return map[string]any{
		"scores": []map[string]any{
			{"component_name": "Presentasi", "score": 85},
			{"component_name": "Penguasaan Materi", "score": 85},
			{"component_name": "Kualitas Naskah", "score": 85},
			{"component_name": "Kemampuan Menjawab", "score": 85},
		},
	}
}

// lifecycleApp seeds every role needed to walk a thesis from submission to
// graduation: admin, kaprodi, student, a supervisor, and two examiners.
func lifecycleApp(t *testing.T) *testutil.Client {
	t.Helper()
	db := testutil.SetupTestDB(t)
	router := testutil.SetupTestRouter(t, db)
	client := testutil.NewClient(router)

	testutil.SeedUser(t, db, adminEmail, adminPassword, "Admin IT", "admin_fakultas")
	testutil.SeedUser(t, db, kaprodiEmail, sharedPassword, "Kaprodi", "kaprodi")
	testutil.SeedUser(t, db, studentEmail, studentPassword, "Student IT", "mahasiswa")
	testutil.SeedUser(t, db, supervisorEmail, sharedPassword, "Supervisor", "dosen_pembimbing")
	testutil.SeedUser(t, db, examiner1Email, sharedPassword, "Examiner Satu", "dosen_penguji")
	testutil.SeedUser(t, db, examiner2Email, sharedPassword, "Examiner Dua", "dosen_penguji")
	testutil.SeedActiveAcademicYear(t, db)
	return client
}

// login returns the bearer token for a user via the given cookie-aware client.
func login(t *testing.T, c *testutil.Client, email, pass string) string {
	t.Helper()
	w := c.Do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": pass,
	}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("login %s failed: %d %s", email, w.Code, w.Body.String())
	}
	var body struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	testutil.DecodeBody(t, w, &body)
	if body.Data.AccessToken == "" {
		t.Fatalf("login %s returned no access token", email)
	}
	return body.Data.AccessToken
}

// extractID pulls data.id (UUID string) out of a success response.
func extractID(t *testing.T, rec *httptest.ResponseRecorder, name string) string {
	t.Helper()
	var body struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	testutil.DecodeBody(t, rec, &body)
	if body.Data.ID == "" {
		t.Fatalf("%s response missing data.id: %s", name, rec.Body.String())
	}
	return body.Data.ID
}

// pdfBytes is a minimal-but-valid PDF (the validator only sniffs the %PDF- magic).
func pdfBytes() []byte {
	return []byte("%PDF-1.4\n% test document\n1 0 obj\n<<>>\nendobj\n%%EOF\n")
}

// uploadDoc uploads a document of the given type as the student and returns the
// document ID.
func uploadDoc(t *testing.T, c *testutil.Client, studentToken, thesisID, docType string) string {
	t.Helper()
	w := testutil.DoMultipart(c.Router(), http.MethodPost, "/api/v1/theses/"+thesisID+"/documents",
		map[string]string{"document_type": docType},
		"file", "dokumen.pdf", pdfBytes(), studentToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("upload %s failed: %d %s", docType, w.Code, w.Body.String())
	}
	return extractID(t, w, "upload "+docType)
}

// examinerIDs resolves the two seeded examiner UUIDs via the admin user list.
func examinerIDs(t *testing.T, c *testutil.Client, adminToken string) []string {
	t.Helper()
	var users struct {
		Data []struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"data"`
	}
	w := testutil.DoJSON(c.Router(), http.MethodGet, "/api/v1/admin/users", nil, adminToken)
	testutil.DecodeBody(t, w, &users)
	ids := make([]string, 0, 2)
	for _, u := range users.Data {
		if u.Email == examiner1Email || u.Email == examiner2Email {
			ids = append(ids, u.ID)
		}
	}
	assert.Len(t, ids, 2, "both examiners must be seeded")
	return ids
}

// lifecycleTokens bundles the bearer tokens needed across a lifecycle test.
type lifecycleTokens struct {
	student    string
	admin      string
	kaprodi    string
	supervisor string
	examiner1  string
	examiner2  string
}

// loginLifecycle logs in every seeded role and returns the tokens.
func loginLifecycle(t *testing.T, c *testutil.Client) lifecycleTokens {
	t.Helper()
	return lifecycleTokens{
		student:    login(t, c, studentEmail, studentPassword),
		admin:      login(t, c, adminEmail, adminPassword),
		kaprodi:    login(t, c, kaprodiEmail, sharedPassword),
		supervisor: login(t, c, supervisorEmail, sharedPassword),
		examiner1:  login(t, c, examiner1Email, sharedPassword),
		examiner2:  login(t, c, examiner2Email, sharedPassword),
	}
}

// advanceToScheduledDefense walks a thesis from submission through to a
// scheduled defense (with both examiners assigned), returning the defense ID.
func advanceToScheduledDefense(t *testing.T, c *testutil.Client, tok lifecycleTokens) string {
	t.Helper()

	// 1) Student submits a thesis → submitted.
	wSubmit := c.Do(http.MethodPost, "/api/v1/theses", thesisPayload(), tok.student)
	assert.Equal(t, http.StatusCreated, wSubmit.Code, "create thesis: %s", wSubmit.Body.String())
	thesisID := extractID(t, wSubmit, "thesis")

	// 2) Kaprodi approves → approved.
	wApprove := c.Do(http.MethodPut, "/api/v1/theses/"+thesisID+"/review",
		map[string]string{"decision": "approved", "notes": "OK"}, tok.kaprodi)
	assert.Equal(t, http.StatusOK, wApprove.Code, "approve: %s", wApprove.Body.String())

	// 3) Kaprodi assigns a supervisor → in_progress.
	supervisorID := userIDByEmail(t, c, tok.admin, supervisorEmail)
	assert.NotEmpty(t, supervisorID, "supervisor must be seeded")
	wAssign := c.Do(http.MethodPut, "/api/v1/theses/"+thesisID+"/assign-supervisor",
		map[string]any{"supervisor_ids": []string{supervisorID}}, tok.kaprodi)
	assert.Equal(t, http.StatusOK, wAssign.Code, "assign supervisor: %s", wAssign.Body.String())

	// 4) Student uploads seminar_doc; supervisor approves it.
	docID := uploadDoc(t, c, tok.student, thesisID, entity.DocTypeSeminarDoc)
	wDocReview := c.Do(http.MethodPatch, "/api/v1/documents/"+docID+"/review",
		map[string]string{"decision": "approved"}, tok.supervisor)
	assert.Equal(t, http.StatusOK, wDocReview.Code, "review seminar doc: %s", wDocReview.Body.String())

	// 5) Student submits seminar → seminar_ready.
	wSemSubmit := c.Do(http.MethodPost, "/api/v1/theses/"+thesisID+"/seminars", nil, tok.student)
	assert.Equal(t, http.StatusCreated, wSemSubmit.Code, "submit seminar: %s", wSemSubmit.Body.String())
	seminarID := extractID(t, wSemSubmit, "seminar")

	// 6) Admin schedules the seminar with two examiners.
	examiners := examinerIDs(t, c, tok.admin)
	wSemSchedule := c.Do(http.MethodPut, "/api/v1/seminars/"+seminarID+"/schedule",
		map[string]any{
			"scheduled_at": time.Now().Add(4 * 24 * time.Hour).Format(time.RFC3339),
			"room":         lifecycleRoom,
			"examiner_ids": examiners,
		}, tok.admin)
	assert.Equal(t, http.StatusOK, wSemSchedule.Code, "schedule seminar: %s", wSemSchedule.Body.String())

	// 7) Both examiners submit seminar scores → auto-finalize → seminar_done.
	wS1 := c.Do(http.MethodPost, "/api/v1/seminars/"+seminarID+"/scores", fullScoreRequest(), tok.examiner1)
	assert.Equal(t, http.StatusOK, wS1.Code, "seminar scores 1: %s", wS1.Body.String())
	wS2 := c.Do(http.MethodPost, "/api/v1/seminars/"+seminarID+"/scores", fullScoreRequest(), tok.examiner2)
	assert.Equal(t, http.StatusOK, wS2.Code, "seminar scores 2: %s", wS2.Body.String())

	// 8) Defense gate: upload defense_doc and approve it.
	defenseDocID := uploadDoc(t, c, tok.student, thesisID, entity.DocTypeDefenseDoc)
	wDefDocReview := c.Do(http.MethodPatch, "/api/v1/documents/"+defenseDocID+"/review",
		map[string]string{"decision": "approved"}, tok.supervisor)
	assert.Equal(t, http.StatusOK, wDefDocReview.Code, "review defense doc: %s", wDefDocReview.Body.String())

	// 9) Student submits defense → defense_ready.
	wDefSubmit := c.Do(http.MethodPost, "/api/v1/theses/"+thesisID+"/defenses", nil, tok.student)
	assert.Equal(t, http.StatusCreated, wDefSubmit.Code, "submit defense: %s", wDefSubmit.Body.String())
	defenseID := extractID(t, wDefSubmit, "defense")

	// 10) Admin schedules the defense (7+ day lead time, two examiners).
	wDefSchedule := c.Do(http.MethodPut, "/api/v1/defenses/"+defenseID+"/schedule",
		map[string]any{
			"scheduled_at": time.Now().Add(8 * 24 * time.Hour).Format(time.RFC3339),
			"room":         lifecycleRoom2,
			"examiner_ids": examiners,
		}, tok.admin)
	assert.Equal(t, http.StatusOK, wDefSchedule.Code, "schedule defense: %s", wDefSchedule.Body.String())

	return defenseID
}

// userIDByEmail resolves a user's UUID via the admin user list.
func userIDByEmail(t *testing.T, c *testutil.Client, adminToken, email string) string {
	t.Helper()
	var users struct {
		Data []struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"data"`
	}
	w := testutil.DoJSON(c.Router(), http.MethodGet, "/api/v1/admin/users", nil, adminToken)
	testutil.DecodeBody(t, w, &users)
	for _, u := range users.Data {
		if u.Email == email {
			return u.ID
		}
	}
	return ""
}

// TestFullThesisLifecycle walks the entire workflow end-to-end:
// Submit → Approve → Assign → Upload → Review → Seminar → Score → Defense → Graduate.
func TestFullThesisLifecycle(t *testing.T) {
	c := lifecycleApp(t)
	tok := loginLifecycle(t, c)

	defenseID := advanceToScheduledDefense(t, c, tok)
	thesisID := thesisIDForDefense(t, c, tok.student, defenseID)

	// 11) Both examiners submit defense scores → auto-finalize → defense_done.
	wD1 := c.Do(http.MethodPost, "/api/v1/defenses/"+defenseID+"/scores", fullScoreRequest(), tok.examiner1)
	assert.Equal(t, http.StatusOK, wD1.Code, "defense scores 1: %s", wD1.Body.String())
	wD2 := c.Do(http.MethodPost, "/api/v1/defenses/"+defenseID+"/scores", fullScoreRequest(), tok.examiner2)
	assert.Equal(t, http.StatusOK, wD2.Code, "defense scores 2: %s", wD2.Body.String())

	// 12) Graduation gate: final_thesis must be uploaded and approved.
	finalDocID := uploadDoc(t, c, tok.student, thesisID, entity.DocTypeFinalThesis)
	wFinalReview := c.Do(http.MethodPatch, "/api/v1/documents/"+finalDocID+"/review",
		map[string]string{"decision": "approved"}, tok.supervisor)
	assert.Equal(t, http.StatusOK, wFinalReview.Code, "review final thesis: %s", wFinalReview.Body.String())

	// 13) Kaprodi graduates the thesis.
	wGrad := c.Do(http.MethodPut, "/api/v1/theses/"+thesisID+"/graduation",
		map[string]string{"notes": "Lulus dengan pujian"}, tok.kaprodi)
	assert.Equal(t, http.StatusOK, wGrad.Code, "graduate: %s", wGrad.Body.String())

	// 14) Verify the thesis reached the terminal graduated state.
	var thesisDetail struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	wGet := testutil.DoJSON(c.Router(), http.MethodGet, "/api/v1/theses/"+thesisID, nil, tok.student)
	testutil.DecodeBody(t, wGet, &thesisDetail)
	assert.Equal(t, "graduated", thesisDetail.Data.Status)
}

// thesisIDForDefense resolves the thesis behind a defense via the student's
// defense list (GET /defenses returns data[].thesis.id for the owner).
func thesisIDForDefense(t *testing.T, c *testutil.Client, studentToken, defenseID string) string {
	t.Helper()
	var body struct {
		Data []struct {
			ID     string `json:"id"`
			Thesis *struct {
				ID string `json:"id"`
			} `json:"thesis"`
		} `json:"data"`
	}
	w := testutil.DoJSON(c.Router(), http.MethodGet, "/api/v1/defenses", nil, studentToken)
	testutil.DecodeBody(t, w, &body)
	for _, d := range body.Data {
		if d.ID == defenseID && d.Thesis != nil {
			return d.Thesis.ID
		}
	}
	t.Fatalf("defense %s not found in student's defense list", defenseID)
	return ""
}

// TestConcurrentDefenseScoreSubmission verifies the race-condition guard in
// FinalizeDefense: when both examiners submit scores at the same time, exactly
// one finalization runs (the FOR UPDATE row lock serializes them), producing a
// single correct final score instead of double-counting.
func TestConcurrentDefenseScoreSubmission(t *testing.T) {
	c := lifecycleApp(t)
	tok := loginLifecycle(t, c)
	defenseID := advanceToScheduledDefense(t, c, tok)

	// Fire both examiners' submissions concurrently. The stateless DoJSON helper
	// is used because the cookie-aware client is not goroutine-safe.
	start := make(chan struct{})
	codes := make(chan int, 2)
	submit := func(token string) {
		<-start
		w := testutil.DoJSON(c.Router(), http.MethodPost,
			"/api/v1/defenses/"+defenseID+"/scores", fullScoreRequest(), token)
		codes <- w.Code
	}
	go submit(tok.examiner1)
	go submit(tok.examiner2)
	close(start)

	var statuses []int
	for i := 0; i < 2; i++ {
		statuses = append(statuses, <-codes)
	}
	// The race guard must not error out legitimate examiner submissions: both
	// scores land, and at most the finalize path is what gets serialized.
	for _, s := range statuses {
		assert.Equal(t, http.StatusOK, s, "concurrent score submission should succeed")
	}

	// Exactly one finalization must have occurred: the final score is the
	// average of the two examiners' 85s (85.0), and the thesis advanced to
	// defense_done — proving the second finalize observed status != scheduled
	// and no-op'd rather than recomputing/double-counting.
	var result struct {
		Data struct {
			Status      string   `json:"status"`
			FinalScore  *float64 `json:"final_score"`
			IsComplete  bool     `json:"is_complete"`
			ExaminerScores []struct {
				ExaminerScore float64 `json:"examiner_score"`
			} `json:"examiner_scores"`
		} `json:"data"`
	}
	wRes := testutil.DoJSON(c.Router(), http.MethodGet, "/api/v1/defenses/"+defenseID+"/result", nil, tok.student)
	testutil.DecodeBody(t, wRes, &result)
	assert.Equal(t, http.StatusOK, wRes.Code)
	assert.True(t, result.Data.IsComplete, "defense should be finalized")
	if result.Data.FinalScore == nil {
		t.Fatal("final score must be set after both examiners submit")
	}
	assert.InDelta(t, 85.0, *result.Data.FinalScore, 0.001, "final score = average of examiners")
	assert.Equal(t, "passed", result.Data.Status)
	// Two distinct examiner scores are averaged — a double-count bug would
	// inflate the examiner list or the score beyond 85.
	assert.Len(t, result.Data.ExaminerScores, 2, "both examiners contribute exactly once")

	thesisID := thesisIDForDefense(t, c, tok.student, defenseID)
	var thesisDetail struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	wThesis := testutil.DoJSON(c.Router(), http.MethodGet, "/api/v1/theses/"+thesisID, nil, tok.student)
	testutil.DecodeBody(t, wThesis, &thesisDetail)
	assert.Equal(t, "defense_done", thesisDetail.Data.Status)
}
