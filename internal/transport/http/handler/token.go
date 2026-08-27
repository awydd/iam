package handler

import (
	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/pkg/response"
	"github.com/awydd/iam/pkg/response/code"
	"github.com/gin-gonic/gin"
)

type TokenHandler struct {
	tokenBiz *biz.TokenBiz
}

func NewTokenHandler(tokenBiz *biz.TokenBiz) *TokenHandler {
	return &TokenHandler{tokenBiz: tokenBiz}
}

type TokenListReq struct {
	UserID        int `form:"user_id"`
	ApplicationID int `form:"application_id" binding:"min=0"`
	Pagination
}

func (h *TokenHandler) List(c *gin.Context) {
	var params TokenListReq
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	list, count, err := h.tokenBiz.List(c.Request.Context(), params.UserID, params.ApplicationID, params.Page, params.PerPage)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.Success(c, ListResp{
		Content: list,
		Count:   count,
	})
}

func (h *TokenHandler) Info(c *gin.Context) {
	var params InfoPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	info, err := h.tokenBiz.Info(c.Request.Context(), params.ID)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.Success(c, info)
}

func (h *TokenHandler) Revoke(c *gin.Context) {
	var params InfoPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	if err := h.tokenBiz.Revoke(c.Request.Context(), params.ID); err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.OK(c)
}
