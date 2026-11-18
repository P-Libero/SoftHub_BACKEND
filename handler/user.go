package handler

import (
	"errors"
	"softhub/middleware"
	"softhub/pkg/response"
	"softhub/request"
	"softhub/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	Service *service.UserService
}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{Service: s}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req request.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := h.Service.Register(c.Request.Context(), req); err != nil {
		if errors.Is(err, service.ErrUsernameTaken) {
			response.BadRequest(c, err.Error())
			return
		}
		_ = c.Error(err)
		response.ServerError(c, "服务器内部错误")
		return
	}
	response.Success(c, nil)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req request.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误:"+err.Error())
		return
	}
	token, err := h.Service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(c, err.Error())
			return
		}
		_ = c.Error(err)
		response.ServerError(c, "服务器内部错误")
		return
	}

	res := map[string]any{
		"token": token,
	}
	response.Success(c, res)
}

func (h *UserHandler) FindMe(c *gin.Context) {
	id := c.GetUint64(middleware.CtxUserID)
	user, err := h.Service.GetProfile(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.Unauthorized(c, err.Error())
			return
		}
		_ = c.Error(err)
		response.ServerError(c, "服务器内部错误")
		return
	}
	response.Success(c, user)
}

func (h *UserHandler) Logout(c *gin.Context) {
	token := c.GetString(middleware.CtxToken)
	expiresAt := c.GetTime(middleware.CtxExpiresAt)
	if token == "" || expiresAt.IsZero() {
		response.Unauthorized(c, "未登录")
		return
	}
	if err := h.Service.Logout(c.Request.Context(), token, expiresAt); err != nil {
		_ = c.Error(err)
		response.ServerError(c, "服务器内部错误")
		return
	}
	response.Success(c, nil)
}
