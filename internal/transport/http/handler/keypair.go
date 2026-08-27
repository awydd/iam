package handler

import (
	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/pkg/response"
	"github.com/awydd/iam/pkg/response/code"
	"github.com/gin-gonic/gin"
)

type KeypairHandler struct {
	keypairBiz *biz.KeypairBiz
}

func NewKeypairHandler(keypairBiz *biz.KeypairBiz) *KeypairHandler {
	return &KeypairHandler{keypairBiz: keypairBiz}
}

func (h *KeypairHandler) JWKS(c *gin.Context) {
	set, err := h.keypairBiz.JWKS(c.Request.Context())
	if err != nil {
		response.Err(c, code.InternalError)
		return
	}
	response.Success(c, set)
}

type KeypairListReq struct {
	Pagination
}

func (h *KeypairHandler) List(c *gin.Context) {
	var params KeypairListReq
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	list, count, err := h.keypairBiz.List(c.Request.Context(), params.Page, params.PerPage)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.Success(c, ListResp{
		Content: list,
		Count:   count,
	})
}

func (h *KeypairHandler) Rotate(c *gin.Context) {
	res, err := h.keypairBiz.Rotate(c.Request.Context())
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.Success(c, res)
}

type KeypairKidPathParams struct {
	Kid string `uri:"kid" binding:"required"`
}

func (h *KeypairHandler) Downgrade(c *gin.Context) {
	var params KeypairKidPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	if err := h.keypairBiz.Downgrade(c.Request.Context(), params.Kid); err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.OK(c)
}

func (h *KeypairHandler) Retire(c *gin.Context) {
	var params KeypairKidPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	if err := h.keypairBiz.Retire(c.Request.Context(), params.Kid); err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.OK(c)
}
