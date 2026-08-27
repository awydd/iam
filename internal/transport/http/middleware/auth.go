package middleware

import (
	"github.com/awydd/iam/internal/biz"
	"github.com/awydd/iam/internal/consts"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/jwt"
	"github.com/awydd/iam/pkg/response"
	"github.com/awydd/iam/pkg/response/code"
	"github.com/awydd/iam/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Auth(jwtManager *jwt.Manager, tokenCache biz.TokenCache, tokenBiz *biz.TokenBiz) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := utils.Cookie().Get(c.Request, consts.CookieAccessToken)
		if err != nil || accessToken == "" {
			response.Err(c, code.Unauthorized)
			c.Abort()
			return
		}

		claims, err := jwtManager.Parse(c.Request.Context(), accessToken)
		if err != nil {
			response.Err(c, code.Unauthorized)
			c.Abort()
			return
		}

		online, err := tokenCache.ExistsAccess(c.Request.Context(), claims.SessionID)
		if err != nil || !online {
			response.Err(c, code.Unauthorized)
			c.Abort()
			return
		}

		c.Set(consts.CtxUserUUID, claims.UserUUID)
		c.Set(consts.CtxSessionID, claims.SessionID)
		c.Set(consts.CtxUsername, claims.Name)

		tokenBiz.Touch(c.Request.Context(), claims.SessionID)

		c.Next()
	}
}

func RequireSystem(userBiz *biz.UserBiz) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUUID, ok := UserUUIDFromContext(c)
		if !ok {
			response.Err(c, code.Unauthorized)
			c.Abort()
			return
		}

		ctx := c.Request.Context()

		u, err := userBiz.GetByUUID(ctx, userUUID)
		if err != nil {
			if db.IsNotFound(err) {
				response.Err(c, code.Unauthorized)
			} else {
				response.Err(c, code.InternalError)
			}
			c.Abort()
			return
		}

		if !u.IsSystem {
			response.Err(c, code.Forbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

func UserUUIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(consts.CtxUserUUID)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func SessionIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(consts.CtxSessionID)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}
