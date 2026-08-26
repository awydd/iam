package handler

import (
	"net/http"

	"github.com/awydd/iam/pkg/utils"
)

type RequestMeta struct {
	IP        string `json:"-"`
	UserAgent string `json:"-"`
}

func GetRequestMeta(r *http.Request) RequestMeta {
	return RequestMeta{
		IP:        utils.ClientIP(r),
		UserAgent: utils.UserAgent(r),
	}
}

type Pagination struct {
	Page    int `form:"page" binding:"required,min=1"`
	PerPage int `form:"per_page" binding:"required,oneof=10 20 50"`
}

type InfoPathParams struct {
	ID int `uri:"id" binding:"required"`
}

type DeletePathParams struct {
	ID int `uri:"id" binding:"required"`
}
