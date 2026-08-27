package handler

import (
	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/consts"
	"github.com/awydd/iam/internal/logger"
	"github.com/awydd/iam/internal/transport/http/middleware"
	"github.com/awydd/iam/pkg/response"
	"github.com/awydd/iam/pkg/response/code"
	"github.com/awydd/iam/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userBiz *biz.UserBiz
}

func NewUserHandler(userBiz *biz.UserBiz) *UserHandler {
	return &UserHandler{userBiz: userBiz}
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

func (h *UserHandler) ListSessions(c *gin.Context) {
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

func (h *UserHandler) RevokeSession(c *gin.Context) {
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
