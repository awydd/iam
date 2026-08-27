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
