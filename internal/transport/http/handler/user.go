package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/consts"
	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/internal/logger"
	"github.com/awydd/iam/internal/transport/http/middleware"
	"github.com/awydd/iam/pkg/response"
	"github.com/awydd/iam/pkg/response/code"
	"github.com/awydd/iam/pkg/utils"
	"github.com/awydd/iam/templates"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userBiz *biz.UserBiz
}

func NewUserHandler(userBiz *biz.UserBiz) *UserHandler {
	return &UserHandler{userBiz: userBiz}
}

var loginTpl = template.Must(template.ParseFS(templates.FS, "login.html"))

type ShowLoginReq struct {
	Redirect string `form:"redirect"`
}

func (h *UserHandler) validateRedirect(redirect string) (string, error) {
	if redirect == "" {
		return "/", nil
	}

	inputURL, err := url.Parse(redirect)
	if err != nil {
		return "/", nil
	}

	if inputURL.IsAbs() || inputURL.Host != "" {
		return "", errors.New("prohibited external redirect")
	}

	if !strings.HasPrefix(redirect, "/") {
		return "/", nil
	}

	return redirect, nil
}

func (h *UserHandler) ShowLogin(c *gin.Context) {
	var params ShowLoginReq
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	finalRedirect, err := h.validateRedirect(params.Redirect)
	if err != nil {
		logger.Error("validate redirect failed: %s", err)
		finalRedirect = ""
	}

	if params.Redirect != "" && finalRedirect == "" {
		logger.Warn("redirect url was blocked or invalid, input=%s", params.Redirect)
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)

	if err := loginTpl.Execute(c.Writer, gin.H{
		"Redirect":    finalRedirect,
		"LoginApiUrl": conf.Get().HTTP.AuthLoginPath(),
	}); err != nil {
		logger.Error("render login template failed: %s", err)
	}
}

type LoginReq struct {
	Username    string      `json:"username" binding:"required,max=26"`
	Password    string      `json:"password" binding:"required,min=6,max=18"`
	RequestMeta RequestMeta `json:"-"`
}

func (h *UserHandler) Login(c *gin.Context) {
	var body LoginReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	body.RequestMeta = GetRequestMeta(c.Request)

	ctx := c.Request.Context()

	result, err := h.userBiz.Login(ctx, body.Username, body.Password, body.RequestMeta.IP, body.RequestMeta.UserAgent)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	jwtCfg := conf.Get().JWT

	cookie := utils.Cookie()
	cookie.Set(c.Writer, consts.CookieAccessToken, result.AccessToken, int(jwtCfg.AccessTokenTTL.Seconds()))
	cookie.Set(c.Writer, consts.CookieRefreshToken, result.RefreshToken, int(jwtCfg.RefreshTokenTTL.Seconds()))
	cookie.Set(c.Writer, consts.CookieSessionID, result.SessionID, int(consts.SessionTTL.Seconds()))
	response.OK(c)
}

func (h *UserHandler) Logout(c *gin.Context) {
	cookie := utils.Cookie()
	ctx := c.Request.Context()

	sid, _ := cookie.Get(c.Request, consts.CookieSessionID)

	if sessionID, ok := middleware.SessionIDFromContext(c); ok {
		if err := h.userBiz.Logout(ctx, sessionID, sid); err != nil {
			logger.Error("logout failed: session_id=%s err=%v", sessionID, err)
		}
	}

	cookie.Delete(c.Writer, consts.CookieAccessToken)
	cookie.Delete(c.Writer, consts.CookieRefreshToken)
	cookie.Delete(c.Writer, consts.CookieSessionID)
	response.OK(c)
}

func (h *UserHandler) Me(c *gin.Context) {
	userUUID, ok := middleware.UserUUIDFromContext(c)
	if !ok {
		response.Err(c, code.Unauthorized)
		return
	}

	res, err := h.userBiz.Me(c.Request.Context(), userUUID)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.Success(c, res)
}

func (h *UserHandler) Refresh(c *gin.Context) {
	refreshToken, err := utils.Cookie().Get(c.Request, consts.CookieRefreshToken)
	if err != nil || refreshToken == "" {
		response.Err(c, code.RefreshTokenInvalid)
		return
	}

	meta := GetRequestMeta(c.Request)
	ctx := c.Request.Context()

	result, err := h.userBiz.Refresh(ctx, refreshToken, meta.IP, meta.UserAgent)
	if err != nil {
		cookie := utils.Cookie()
		cookie.Delete(c.Writer, consts.CookieAccessToken)
		cookie.Delete(c.Writer, consts.CookieRefreshToken)

		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	jwtCfg := conf.Get().JWT
	cookie := utils.Cookie()
	cookie.Set(c.Writer, consts.CookieAccessToken, result.AccessToken, int(jwtCfg.AccessTokenTTL.Seconds()))
	cookie.Set(c.Writer, consts.CookieRefreshToken, result.RefreshToken, int(jwtCfg.RefreshTokenTTL.Seconds()))

	response.OK(c)
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=18"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=18"`
}

func (h *UserHandler) Password(c *gin.Context) {
	userUUID, ok := middleware.UserUUIDFromContext(c)
	if !ok {
		response.Err(c, code.Unauthorized)
		return
	}

	var body ChangePasswordReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	ctx := c.Request.Context()

	info, err := h.userBiz.GetByUUID(ctx, userUUID)
	if err != nil {
		response.Err(c, code.Unauthorized)
		return
	}

	if err := h.userBiz.Password(ctx, info.ID, body.OldPassword, body.NewPassword); err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	cookie := utils.Cookie()
	cookie.Delete(c.Writer, consts.CookieAccessToken)
	cookie.Delete(c.Writer, consts.CookieRefreshToken)
	cookie.Delete(c.Writer, consts.CookieSessionID)

	response.OK(c)
}

type SessionListReq struct {
	Pagination
}

func (h *UserHandler) MySessions(c *gin.Context) {
	userUUID, ok := middleware.UserUUIDFromContext(c)
	if !ok {
		response.Err(c, code.Unauthorized)
		return
	}
	currentSessionID, _ := middleware.SessionIDFromContext(c)

	var req SessionListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	ctx := c.Request.Context()

	u, err := h.userBiz.GetByUUID(ctx, userUUID)
	if err != nil {
		response.Err(c, code.Unauthorized)
		return
	}

	list, total, err := h.userBiz.ListSessions(ctx, currentSessionID, u.ID, req.Page, req.PerPage)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.Success(c, ListResp{
		Content: list,
		Count:   total,
	})
}

type RevokeSessionPathParams struct {
	SessionID UUIDParam `uri:"session_id" binding:"required"`
}

func (h *UserHandler) RevokeMySession(c *gin.Context) {
	userUUID, ok := middleware.UserUUIDFromContext(c)
	if !ok {
		response.Err(c, code.Unauthorized)
		return
	}

	var params RevokeSessionPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	ctx := c.Request.Context()

	u, err := h.userBiz.GetByUUID(ctx, userUUID)
	if err != nil {
		response.Err(c, code.Unauthorized)
		return
	}

	if err := h.userBiz.RevokeSession(ctx, u.ID, uuid.UUID(params.SessionID)); err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.OK(c)
}

type UserListReq struct {
	Keyword string `form:"keyword,omitempty" binding:"max=26"`
	Pagination
}

func (h *UserHandler) List(c *gin.Context) {
	var params UserListReq
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	list, count, err := h.userBiz.List(c.Request.Context(), params.Keyword, params.Page, params.PerPage)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.Success(c, ListResp{
		Content: list,
		Count:   count,
	})
}

func (h *UserHandler) StatusOptions(c *gin.Context) {
	response.Success(c, enum.UserStatusOptions())
}

func (h *UserHandler) Info(c *gin.Context) {
	var params InfoPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	info, err := h.userBiz.Info(c.Request.Context(), params.ID)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.Success(c, info)
}

type UserCreateReq struct {
	Username string          `json:"username" binding:"required,max=26"`
	Email    string          `json:"email" binding:"required,email"`
	Status   enum.UserStatus `json:"status" binding:"required"`
	Password string          `json:"password" binding:"required,min=6,max=18"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var body UserCreateReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	if !body.Status.Valid() {
		response.ErrMessage(c, code.BadRequest, "invalid status")
		return
	}

	err := h.userBiz.Create(c.Request.Context(), body.Username, body.Email, body.Password, body.Status)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.OK(c)
}

type UserUpdatePathParams struct {
	UserID int `uri:"id" binding:"required"`
}

type UserUpdateReq struct {
	Email    string          `json:"email" binding:"required,email"`
	Username string          `json:"username" binding:"required,max=26"`
	Status   enum.UserStatus `json:"status" binding:"required"`
	Password string          `json:"password" binding:"omitempty,min=6,max=18"`
}

func (h *UserHandler) Update(c *gin.Context) {
	var params UserUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	var body UserUpdateReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	if !body.Status.Valid() {
		response.ErrMessage(c, code.BadRequest, "invalid status")
		return
	}

	err := h.userBiz.Update(c.Request.Context(), params.UserID, body.Username, body.Email, body.Password, body.Status)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.OK(c)
}

func (h *UserHandler) Delete(c *gin.Context) {
	var params DeletePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	if err := h.userBiz.Delete(c.Request.Context(), params.ID); err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.OK(c)
}

type AdminUserPathParams struct {
	ID int `uri:"id" binding:"required"`
}

func (h *UserHandler) RevokeSessions(c *gin.Context) {
	var pathParams AdminUserPathParams
	if err := c.ShouldBindUri(&pathParams); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	if err := h.userBiz.InvalidateAllSessions(c.Request.Context(), pathParams.ID); err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.OK(c)
}
