package handler

import (
	"fmt"

	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/pkg/response"
	"github.com/awydd/iam/pkg/response/code"
	"github.com/gin-gonic/gin"
)

type ApplicationHandler struct {
	applicationBiz *biz.ApplicationBiz
}

func NewApplicationHandler(applicationBiz *biz.ApplicationBiz) *ApplicationHandler {
	return &ApplicationHandler{applicationBiz: applicationBiz}
}

type ApplicationListReq struct {
	Keyword string `form:"keyword,omitempty" binding:"max=26"`
	Pagination
}

func (h *ApplicationHandler) List(c *gin.Context) {
	var params ApplicationListReq
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	list, count, err := h.applicationBiz.List(c.Request.Context(), params.Keyword, params.Page, params.PerPage)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.Success(c, ListResp{
		Content: list,
		Count:   count,
	})
}

func (h *ApplicationHandler) Info(c *gin.Context) {
	var params InfoPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	info, err := h.applicationBiz.Info(c.Request.Context(), params.ID)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.Success(c, info)
}

type ApplicationCreateReq struct {
	Name            string                     `json:"name" binding:"required,max=64"`
	ClientID        string                     `json:"client_id" binding:"required,max=36"`
	RedirectUris    []string                   `json:"redirect_uris" binding:"required,min=1,dive,required,url"`
	Type            enum.ApplicationClientType `json:"type" binding:"required"`
	Status          enum.ApplicationStatus     `json:"status" binding:"required"`
	AccessTokenTTL  int                        `json:"access_token_ttl" binding:"required,min=60,max=900"`
	RefreshTokenTTL int                        `json:"refresh_token_ttl" binding:"required,min=3600,max=604800"`
}

func (h *ApplicationHandler) Create(c *gin.Context) {
	var body ApplicationCreateReq
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println(err)
		response.Err(c, code.InvalidParams)
		return
	}

	if !body.Status.Valid() {
		response.ErrMessage(c, code.BadRequest, "invalid status")
		return
	}

	res, err := h.applicationBiz.Create(
		c.Request.Context(),
		body.Name, body.ClientID,
		body.RedirectUris,
		body.Type,
		body.Status,
		body.AccessTokenTTL, body.RefreshTokenTTL,
	)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.Success(c, res)
}

type ApplicationUpdatePathParams struct {
	ApplicationID int `uri:"id" binding:"required"`
}

type ApplicationUpdateInfoReq struct {
	Name         string   `json:"name" binding:"required,max=64"`
	ClientID     string   `json:"client_id" binding:"required,max=36"`
	RedirectUris []string `json:"redirect_uris" binding:"required,min=1,dive,required,url"`
}

func (h *ApplicationHandler) Update(c *gin.Context) {
	var params ApplicationUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	var body ApplicationUpdateInfoReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	if err := h.applicationBiz.UpdateInfo(c.Request.Context(), params.ApplicationID, body.Name, body.ClientID, body.RedirectUris); err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.OK(c)
}

type ApplicationUpdateTTLReq struct {
	AccessTokenTTL  int `json:"access_token_ttl" binding:"required,min=60,max=900"`
	RefreshTokenTTL int `json:"refresh_token_ttl" binding:"required,min=3600,max=604800"`
}

func (h *ApplicationHandler) UpdateTTL(c *gin.Context) {
	var params ApplicationUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	var body ApplicationUpdateTTLReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	if err := h.applicationBiz.UpdateTTL(c.Request.Context(), params.ApplicationID, body.AccessTokenTTL, body.RefreshTokenTTL); err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.OK(c)
}

type ApplicationUpdateStatusReq struct {
	Status enum.ApplicationStatus `json:"status" binding:"required"`
}

func (h *ApplicationHandler) UpdateStatus(c *gin.Context) {
	var params ApplicationUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	var body ApplicationUpdateStatusReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}
	if !body.Status.Valid() {
		response.ErrMessage(c, code.BadRequest, "invalid status")
		return
	}

	if err := h.applicationBiz.UpdateStatus(c.Request.Context(), params.ApplicationID, body.Status); err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.OK(c)
}

func (h *ApplicationHandler) UpdateSecret(c *gin.Context) {
	var params ApplicationUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	res, err := h.applicationBiz.UpdateSecret(c.Request.Context(), params.ApplicationID)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.Success(c, res)
}

func (h *ApplicationHandler) Delete(c *gin.Context) {
	var params DeletePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	if err := h.applicationBiz.Delete(c.Request.Context(), params.ID); err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.OK(c)
}
