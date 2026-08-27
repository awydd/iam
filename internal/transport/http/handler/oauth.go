package handler

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/consts"
	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/pkg/random"
	"github.com/awydd/iam/pkg/response"
	"github.com/awydd/iam/pkg/response/code"
	"github.com/awydd/iam/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OAuthHandler struct {
	userBiz        *biz.UserBiz
	applicationBiz *biz.ApplicationBiz
	codeCache      biz.OAuthCodeCache
}

func NewOAuthHandler(userBiz *biz.UserBiz, applicationBiz *biz.ApplicationBiz, codeCache biz.OAuthCodeCache) *OAuthHandler {
	return &OAuthHandler{userBiz: userBiz, applicationBiz: applicationBiz, codeCache: codeCache}
}

type AuthorizeReq struct {
	ResponseType        string `form:"response_type" binding:"required"`
	ClientID            string `form:"client_id" binding:"required"`
	RedirectURI         string `form:"redirect_uri" binding:"required"`
	State               string `form:"state"`
	CodeChallenge       string `form:"code_challenge"`
	CodeChallengeMethod string `form:"code_challenge_method"`
}

type ExchangeTokenReq struct {
	GrantType    string `form:"grant_type" binding:"required"`
	Code         string `form:"code" binding:"required"`
	RedirectURI  string `form:"redirect_uri" binding:"required"`
	ClientID     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
	CodeVerifier string `form:"code_verifier"`
}

func (h *OAuthHandler) errorRedirect(c *gin.Context, redirectURI, state, errCode, desc string) {
	target, err := url.Parse(redirectURI)
	if err != nil {
		response.Err(c, code.InvalidParams)
		return
	}
	q := target.Query()
	q.Set("error", errCode)
	if desc != "" {
		q.Set("error_description", desc)
	}
	if state != "" {
		q.Set("state", state)
	}
	target.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, target.String())
}

func (h *OAuthHandler) Authorize(c *gin.Context) {
	var params AuthorizeReq
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	if params.ResponseType != "code" {
		response.ErrMessage(c, code.InvalidParams, "response_type must be \"code\"")
		return
	}

	ctx := c.Request.Context()

	app, err := h.applicationBiz.ValidateAuthorize(ctx, params.ClientID, params.RedirectURI)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	if params.CodeChallenge != "" && params.CodeChallengeMethod != "S256" {
		h.errorRedirect(c, params.RedirectURI, params.State, "invalid_request", "transform algorithm must be S256")
		return
	}

	if app.Type == enum.ApplicationClientTypePublic && params.CodeChallenge == "" {
		h.errorRedirect(c, params.RedirectURI, params.State, "invalid_request", "PKCE (code_challenge) is required for public clients")
		return
	}

	sid, _ := utils.Cookie().Get(c.Request, consts.CookieSessionID)

	userID, userUUID, ok := h.userBiz.CheckSession(ctx, sid)
	if !ok {
		v := url.Values{}
		v.Set("redirect", c.Request.URL.RequestURI())
		c.Redirect(http.StatusFound, conf.Get().HTTP.AuthLoginPath()+"?"+v.Encode())
		return
	}

	authCode, err := random.Alphanumeric(32)
	if err != nil {
		h.errorRedirect(c, params.RedirectURI, params.State, "server_error", "")
		return
	}

	appSessionID := uuid.New()

	if err := h.codeCache.SetCode(ctx, authCode, biz.OAuthCodePayload{
		UserID:              userID,
		UserUUID:            userUUID,
		SessionID:           appSessionID,
		ClientID:            params.ClientID,
		RedirectURI:         params.RedirectURI,
		State:               params.State,
		CodeChallenge:       params.CodeChallenge,
		CodeChallengeMethod: params.CodeChallengeMethod,
	}); err != nil {
		h.errorRedirect(c, params.RedirectURI, params.State, "server_error", "")
		return
	}

	target, err := url.Parse(params.RedirectURI)
	if err != nil {
		response.Err(c, code.InvalidParams)
		return
	}
	q := target.Query()
	q.Set("code", authCode)
	if params.State != "" {
		q.Set("state", params.State)
	}
	target.RawQuery = q.Encode()

	c.Redirect(http.StatusFound, target.String())
}

func verifyPKCE(verifier, challenge, method string) bool {
	if challenge == "" {
		return true
	}
	if verifier == "" || method != "S256" {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])
	return computed == challenge
}

func resolveClientCredentials(r *http.Request, bodyClientID, bodyClientSecret string) (clientID, clientSecret string, err error) {
	basicID, basicSecret, hasBasic := r.BasicAuth()

	switch {
	case hasBasic && bodyClientID != "" && basicID != bodyClientID:
		return "", "", errors.New("client_id mismatch between Authorization header and body")
	case hasBasic:
		return basicID, basicSecret, nil
	default:
		return bodyClientID, bodyClientSecret, nil
	}
}

func (h *OAuthHandler) ExchangeToken(c *gin.Context) {
	var params ExchangeTokenReq
	if err := c.ShouldBind(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	if params.GrantType != "authorization_code" {
		response.Err(c, code.UnsupportedGrantType)
		return
	}

	clientID, clientSecret, err := resolveClientCredentials(c.Request, params.ClientID, params.ClientSecret)
	if err != nil {
		response.ErrMessage(c, code.InvalidParams, err.Error())
		return
	}
	if clientID == "" {
		response.ErrMessage(c, code.InvalidParams, "client_id is required")
		return
	}

	ctx := c.Request.Context()

	payload, err := h.codeCache.GetAndDelCode(ctx, params.Code)
	if err != nil {
		response.Err(c, code.InvalidOrExpiredCode)
		return
	}

	if payload.ClientID != clientID {
		response.ErrMessage(c, code.InvalidParams, "client mismatch")
		return
	}
	if payload.RedirectURI != params.RedirectURI {
		response.Err(c, code.RedirectURIInvalid)
		return
	}
	if !verifyPKCE(params.CodeVerifier, payload.CodeChallenge, payload.CodeChallengeMethod) {
		response.Err(c, code.InvalidCodeVerifier)
		return
	}

	app, err := h.applicationBiz.ValidateClient(ctx, clientID, clientSecret)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	meta := GetRequestMeta(c.Request)

	result, err := h.userBiz.LoginWithOAuthCode(ctx, payload, app.ID, meta.IP, meta.UserAgent)
	if err != nil {
		response.ErrMessage(c, code.InternalError, "failed to issue token")
		return
	}

	response.Success(c, result)
}
