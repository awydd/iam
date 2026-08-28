// 用于测试 IAM OAuth2 授权码和令牌流程的本地测试程序。
//
// 使用方法：
//
//	# 场景 1：机密客户端，通过 HTTP Basic Auth 传递凭证（推荐）
//	go run main.go -mode=confidential-basic \
//	  -client-id=internal-app-001 -client-secret=your-secret
//
//	# 场景 2：机密客户端，通过请求体传递凭证
//	go run main.go -mode=confidential-body \
//	  -client-id=internal-app-001 -client-secret=your-secret
//
//	# 场景 3：机密客户端且无密钥 - 负向测试，IAM 必须在 /token 阶段拒绝
//	go run main.go -mode=confidential-no-secret \
//	  -client-id=internal-app-001
//
//	# 场景 4：公共客户端（SPA/移动端），必须使用 PKCE，无密钥
//	go run main.go -mode=public \
//	  -client-id=spa-app-001
//
//	# 场景 5：公共客户端但未使用 PKCE - 负向测试，IAM 必须在 /authorize 阶段拒绝
//	go run main.go -mode=public-no-pkce \
//	  -client-id=spa-app-001
//
// 启动后在浏览器中打开 http://127.0.0.1:9999/start，登录后将在 /callback 页面看到可直接复制的 curl 命令。
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
	sessions = map[string]*session{}
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
		authorizeURL = flag.String("authorize-url", "http://127.0.0.1:26824/api/v1/oauth/authorize", "IAM 授权端点完整 URL")
		tokenURL     = flag.String("token-url", "http://127.0.0.1:26824/api/v1/oauth/token", "IAM 令牌端点完整 URL")
		clientID     = flag.String("client-id", "test-client", "已在 IAM 注册的 client_id")
		clientSecret = flag.String("client-secret", "", "客户端密钥；公共客户端及无密钥模式下会被忽略")
		redirectURI  = flag.String("redirect-uri", "http://127.0.0.1:9999/callback", "必须与 IAM 中注册的回调地址完全一致")
		modeFlag     = flag.String("mode", string(modeConfidentialBody), "运行模式选择")
		addr         = flag.String("addr", "127.0.0.1:9999", "本地监听地址")
	)
	flag.Parse()

	m := mode(*modeFlag)
	switch m {
	case modeConfidentialBasic, modeConfidentialBody:
		if *clientSecret == "" {
			log.Fatalf("当前模式 %s 需要提供 -client-secret", m)
		}
	case modeConfidentialNoSecret:
		if *clientSecret != "" {
			log.Printf("提示：模式 %s 会忽略 -client-secret，此处已自动清空", m)
			*clientSecret = ""
		}
	case modePublic, modePublicNoPKCE:
		if *clientSecret != "" {
			log.Printf("提示：模式 %s 会忽略 -client-secret，公共客户端不能持有密钥", m)
			*clientSecret = ""
		}
	default:
		log.Fatalf("未知的 -mode=%s", m)
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

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if errCode := r.URL.Query().Get("error"); errCode != "" {
			desc := r.URL.Query().Get("error_description")
			w.WriteHeader(http.StatusBadRequest)

			expected := ""
			if m == modePublicNoPKCE {
				expected = `<p style="color:green">符合预期：IAM 已成功拒绝未提供 PKCE 的公共客户端。</p>`
			}

			fmt.Fprintf(w, `<h2>IAM 返回了错误</h2><p>错误码: %s</p><p>描述: %s</p>%s`,
				html.EscapeString(errCode), html.EscapeString(desc), expected)
			return
		}

		if m == modePublicNoPKCE {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, `<h2>异常：IAM 在公共客户端未提供 PKCE 的情况下签发了授权码。</h2>`)
			return
		}

		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "<p>回调中未找到授权码。</p>")
			return
		}

		mu.Lock()
		sess, ok := sessions[state]
		if ok {
			delete(sessions, state)
		}
		mu.Unlock()

		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "<h2>State 校验失败</h2><p>可能存在 CSRF 风险，客户端应予以拒绝。</p>")
			return
		}

		var curl, label, note string
		switch m {
		case modeConfidentialBasic:
			label = "机密客户端 - HTTP Basic Auth"
			curl = buildBasicAuthCurl(*tokenURL, *clientID, *clientSecret, *redirectURI, code, sess.codeVerifier)
			note = "凭证放在 Authorization 请求头中。"
		case modeConfidentialBody:
			label = "机密客户端 - 请求体传参"
			curl = buildBodyAuthCurl(*tokenURL, *clientID, *clientSecret, *redirectURI, code, sess.codeVerifier)
			note = "通过表单字段传递 client_id 和 client_secret。"
		case modeConfidentialNoSecret:
			label = "机密客户端 - 无密钥（负向测试，应被拒绝）"
			curl = buildBodyAuthCurl(*tokenURL, *clientID, "", *redirectURI, code, sess.codeVerifier)
			note = "未提供 client_secret，请执行该 curl 确认令牌端点返回 invalid_client 错误。"
		case modePublic:
			label = "公共客户端 - 仅 PKCE，无密钥"
			curl = buildBodyAuthCurl(*tokenURL, *clientID, "", *redirectURI, code, sess.codeVerifier)
			note = "安全性完全依赖下方的 code_verifier。"
		}

		fmt.Fprintf(w, `
<h2>已获取授权码</h2>
<p><b>模式:</b> %s</p>
<p><b>授权码:</b> %s</p>
<p><b>State 校验:</b> 通过</p>
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
	log.Printf("测试服务已启动 [模式=%s]，请在浏览器中打开: %s\n", m, startURL)
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
