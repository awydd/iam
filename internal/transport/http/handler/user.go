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
