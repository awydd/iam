// Quick local test harness for the IAM OAuth2 authorize/token flow.
//
// Usage:
//
//	# Scenario 1: Confidential Client, credentials via HTTP Basic Auth (RFC 6749 recommended)
//	go run main.go -mode=confidential-basic \
//	  -client-id=internal-app-001 -client-secret=your-secret
//
//	# Scenario 2: Confidential Client, credentials via request body (legacy but valid)
//	go run main.go -mode=confidential-body \
//	  -client-id=internal-app-001 -client-secret=your-secret
//
//	# Scenario 3: Confidential Client WITHOUT secret — negative test, IAM must reject
//	# this at /token exchange with "invalid client credentials" (ErrClientSecretWrong)
//	go run main.go -mode=confidential-no-secret \
//	  -client-id=internal-app-001
//
//	# Scenario 4: Public Client (SPA/mobile), PKCE mandatory, no secret
//	go run main.go -mode=public \
//	  -client-id=spa-app-001
//
//	# Scenario 5: Public Client WITHOUT PKCE — negative test, IAM must reject
//	# this at /authorize with "PKCE (code_challenge) is required for public clients"
//	go run main.go -mode=public-no-pkce \
//	  -client-id=spa-app-001
//
// Then open http://127.0.0.1:9999/start in your browser, log in,
// and you'll land on /callback with ready-to-copy curl commands
// (or, for negative-test scenarios, IAM's rejection reason).
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"sync"
)

type mode string

const (
	modeConfidentialBasic    mode = "confidential-basic"
	modeConfidentialBody     mode = "confidential-body"
	modeConfidentialNoSecret mode = "confidential-no-secret"
	modePublic               mode = "public"
	modePublicNoPKCE         mode = "public-no-pkce"
)

type session struct {
	state        string
	codeVerifier string
}

var (
	mu       sync.Mutex
	sessions = map[string]*session{} // key: state
)

func randomURLSafe(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func main() {
	var (
		authorizeURL = flag.String("authorize-url", "http://127.0.0.1:26824/api/v1/oauth/authorize", "full URL of IAM's /oauth/authorize endpoint")
		tokenURL     = flag.String("token-url", "http://127.0.0.1:26824/api/v1/oauth/token", "full URL of IAM's /oauth/token endpoint")
		clientID     = flag.String("client-id", "test-client", "client_id already registered in IAM")
		clientSecret = flag.String("client-secret", "", "client_secret; ignored/must be empty for -mode=public* and -mode=confidential-no-secret")
		redirectURI  = flag.String("redirect-uri", "http://127.0.0.1:9999/callback", "must match the redirect_uri registered in IAM exactly")
		modeFlag     = flag.String("mode", string(modeConfidentialBody), "confidential-basic | confidential-body | confidential-no-secret | public | public-no-pkce")
		addr         = flag.String("addr", "127.0.0.1:9999", "local address this server listens on")
	)
	flag.Parse()

	m := mode(*modeFlag)
	switch m {
	case modeConfidentialBasic, modeConfidentialBody:
		if *clientSecret == "" {
			log.Fatalf("mode=%s requires -client-secret", m)
		}
	case modeConfidentialNoSecret:
		if *clientSecret != "" {
			log.Printf("⚠️  mode=%s ignores -client-secret; this test intentionally omits it", m)
			*clientSecret = ""
		}
	case modePublic, modePublicNoPKCE:
		if *clientSecret != "" {
			log.Printf("⚠️  mode=%s ignores -client-secret; a public client must not hold a secret", m)
			*clientSecret = ""
		}
	default:
		log.Fatalf("unknown -mode=%s (want confidential-basic | confidential-body | confidential-no-secret | public | public-no-pkce)", m)
	}

	usePKCE := m == modePublic

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		state := randomURLSafe(16)
		sess := &session{state: state}

		q := url.Values{}
		q.Set("response_type", "code")
		q.Set("client_id", *clientID)
		q.Set("redirect_uri", *redirectURI)
		q.Set("state", state)

		if usePKCE {
			verifier := randomURLSafe(32)
			sess.codeVerifier = verifier
			q.Set("code_challenge", pkceChallenge(verifier))
			q.Set("code_challenge_method", "S256")
		}

		mu.Lock()
		sessions[state] = sess
		mu.Unlock()

		target := *authorizeURL + "?" + q.Encode()
		http.Redirect(w, r, target, http.StatusFound)
	})

	// Callback: IAM redirects back here once the user is authorized —
	// or, for negative-test modes, redirects back here with an error instead.
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if errCode := r.URL.Query().Get("error"); errCode != "" {
			desc := r.URL.Query().Get("error_description")
			w.WriteHeader(http.StatusBadRequest)

			expected := ""
			if m == modePublicNoPKCE {
				expected = `<p style="color:green">✅ Expected result for -mode=public-no-pkce: IAM correctly rejected a public client that skipped PKCE.</p>`
			}

			fmt.Fprintf(w, `<h2>IAM sent back an error</h2><p>error: %s</p><p>description: %s</p>%s`,
				html.EscapeString(errCode), html.EscapeString(desc), expected)
			return
		}

		if m == modePublicNoPKCE {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, `<h2 style="color:red">❌ Unexpected: IAM issued a code without requiring PKCE for a public client.</h2><p>The PKCE-required guard in Authorize() is not working — check enum.ApplicationClientTypePublic comparison and app.Type value.</p>`)
			return
		}

		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "<p>No code in the callback — something's off in the IAM authorize logic.</p>")
			return
		}

		// state check: CSRF guard
		mu.Lock()
		sess, ok := sessions[state]
		if ok {
			delete(sessions, state) // one-time use, no replay
		}
		mu.Unlock()

		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "<h2 style='color:red'>state check failed</h2><p>This state wasn't issued by this app — could be a CSRF attempt. A real client must reject this.</p>")
			return
		}

		var curl, label, note string
		switch m {
		case modeConfidentialBasic:
			label = "Confidential Client — HTTP Basic Auth (RFC 6749 recommended)"
			curl = buildBasicAuthCurl(*tokenURL, *clientID, *clientSecret, *redirectURI, code, sess.codeVerifier)
			note = "Credentials go in the Authorization header; client_id/client_secret are absent from the body."
		case modeConfidentialBody:
			label = "Confidential Client — credentials in request body"
			curl = buildBodyAuthCurl(*tokenURL, *clientID, *clientSecret, *redirectURI, code, sess.codeVerifier)
			note = "Legacy but RFC-valid: client_id/client_secret sent as form fields."
		case modeConfidentialNoSecret:
			label = "Confidential Client — NO secret (negative test, IAM must reject)"
			curl = buildBodyAuthCurl(*tokenURL, *clientID, "", *redirectURI, code, sess.codeVerifier)
			note = "This client is registered as confidential but sends no client_secret at all. " +
				"POST this curl and confirm the token endpoint returns invalid_client / ClientSecretWrong — " +
				"if it succeeds instead, ValidateClient's empty-secret check is broken."
		case modePublic:
			label = "Public Client — PKCE only, no secret"
			curl = buildBodyAuthCurl(*tokenURL, *clientID, "", *redirectURI, code, sess.codeVerifier)
			note = "No client_secret anywhere. Security relies entirely on the PKCE code_verifier below."
		}

		fmt.Fprintf(w, `
<h2>Got the OAuth code</h2>
<p><b>mode:</b> %s</p>
<p><b>code:</b> %s</p>
<p><b>state check:</b> passed (verified locally against generated state)</p>
%s

<hr/>

<h3>%s</h3>
<p style="color:gray;font-size:13px;">%s</p>
<pre style="background:#eef6ff;padding:12px;border-radius:6px;white-space:pre-wrap;border:1px solid #b6d4fe;">%s</pre>
`,
			html.EscapeString(string(m)),
			html.EscapeString(code),
			pkceNote(sess.codeVerifier),
			html.EscapeString(label),
			html.EscapeString(note),
			html.EscapeString(curl),
		)
	})

	startURL := fmt.Sprintf("http://%s/start", *addr)
	log.Printf("🚀 test app is up [mode=%s], open this in your browser to start:\n   %s\n", m, startURL)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}

func pkceNote(verifier string) string {
	if verifier == "" {
		return ""
	}
	return fmt.Sprintf("<p><b>code_verifier:</b> %s</p>", html.EscapeString(verifier))
}

func buildBasicAuthCurl(tokenURL, clientID, clientSecret, redirectURI, code, codeVerifier string) string {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	if codeVerifier != "" {
		form.Set("code_verifier", codeVerifier)
	}

	authRaw := fmt.Sprintf("%s:%s", clientID, clientSecret)
	authBase64 := base64.StdEncoding.EncodeToString([]byte(authRaw))

	return fmt.Sprintf(`curl -X POST "%s" \
  -H "Authorization: Basic %s" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d '%s'`, tokenURL, authBase64, form.Encode())
}

func buildBodyAuthCurl(tokenURL, clientID, clientSecret, redirectURI, code, codeVerifier string) string {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", clientID)
	form.Set("redirect_uri", redirectURI)
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}
	if codeVerifier != "" {
		form.Set("code_verifier", codeVerifier)
	}

	return fmt.Sprintf(`curl -X POST "%s" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d '%s'`, tokenURL, form.Encode())
}
