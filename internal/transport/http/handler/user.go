package handler

import (
	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/consts"
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
