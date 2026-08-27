package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awydd/iam/conf"
)

func TestCookieUtil_BasicAndOptions(t *testing.T) {
	cfg := struct {
		Secret   string
		Path     string
		Domain   string
		HttpOnly bool
		Secure   bool
		SameSite string
	}{
		Secret:   "FVlDdqPT8yJFmuLEebpVUHnwoDrxfnWT",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: "Lax",
	}

	InitCookieUtil(conf.Cookie{
		Secret:   cfg.Secret,
		Path:     cfg.Path,
		HttpOnly: cfg.HttpOnly,
		SameSite: cfg.SameSite,
	})

	cUtil := Cookie()
	if cUtil == nil {
		t.Fatal("expected cookie util instance, got nil")
	}

	rec := httptest.NewRecorder()
	cUtil.Set(rec, "username", "eren", 3600, WithPath("/api"), WithHttpOnly(false))

	res := rec.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected cookie to be set")
	}

	found := false
	for _, ck := range cookies {
		if ck.Name == "username" {
			found = true
			if ck.Value != "eren" {
				t.Errorf("expected value 'eren', got '%s'", ck.Value)
			}
			if ck.Path != "/api" {
				t.Errorf("expected path '/api', got '%s'", ck.Path)
			}
			if ck.HttpOnly != false {
				t.Errorf("expected HttpOnly false, got true")
			}
		}
	}
	if !found {
		t.Fatal("cookie 'username' not found in response")
	}
}

func TestCookieUtil_Signed(t *testing.T) {
	InitCookieUtil(conf.Cookie{
		Secret:   "FVlDdqPT8yJFmuLEebpVUHnwoDrxfnWT", // > 32
		Path:     "/",
		HttpOnly: true,
	})
	cUtil := Cookie()

	rec := httptest.NewRecorder()
	err := cUtil.SetSigned(rec, "session_id", "user_abc_123", 7200)
	if err != nil {
		t.Fatalf("failed to set signed cookie: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	for _, ck := range rec.Result().Cookies() {
		req.AddCookie(ck)
	}

	val, err := cUtil.GetSigned(req, "session_id")
	if err != nil {
		t.Fatalf("expected valid signature, got error: %v", err)
	}
	if val != "user_abc_123" {
		t.Errorf("expected 'user_abc_123', got '%s'", val)
	}

	badRec := httptest.NewRecorder()
	http.SetCookie(badRec, &http.Cookie{
		Name:  "session_id",
		Value: "user_hacked_999.invalidsignature",
	})
	badReq := httptest.NewRequest("GET", "/", nil)
	for _, ck := range badRec.Result().Cookies() {
		badReq.AddCookie(ck)
	}

	_, err = cUtil.GetSigned(badReq, "session_id")
	if err != ErrInvalidSignature && err != ErrMalformedValue {
		t.Errorf("expected signature error or malformed error, got: %v", err)
	}
}

func TestCookieUtil_Delete(t *testing.T) {
	InitCookieUtil(conf.Cookie{
		Path: "/",
	})
	cUtil := Cookie()

	rec := httptest.NewRecorder()
	cUtil.Delete(rec, "token")

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected delete cookie to be set")
	}

	for _, ck := range cookies {
		if ck.Name == "token" {
			if ck.MaxAge != -1 {
				t.Errorf("expected MaxAge -1 for deletion, got %d", ck.MaxAge)
			}
			if ck.Value != "" {
				t.Errorf("expected empty value for deletion, got '%s'", ck.Value)
			}
		}
	}
}
