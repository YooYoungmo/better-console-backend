package controllers

import (
	"better-console-backend/domain"
	"better-console-backend/domain/auth"
	"better-console-backend/dtos"
	"better-console-backend/security"
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"time"
)

type AuthController struct {
}

func (controller AuthController) Init(g *echo.Group) {
	g.POST("", controller.AuthWithSignIdPassword)
	g.POST("/dooray", controller.AuthWithDoorayIdPassword)
	g.GET("/google-workspace", controller.AuthWithGoogleWorkspaceAccount)
	g.GET("/check", controller.CheckAuth)
	g.POST("/logout", controller.Logout)
	g.POST("/token/refresh", controller.RefreshAccessToken)
}

func (AuthController) AuthWithSignIdPassword(ctx echo.Context) error {
	var memberSignIn dtos.MemberSignIn

	if err := ctx.Bind(&memberSignIn); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	if err := memberSignIn.Validate(ctx); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	jwtToken, err := auth.AuthService{}.AuthWithSignIdPassword(ctx.Request().Context(), memberSignIn)
	if err != nil {
		if err == domain.ErrNotFound || err == domain.ErrAuthentication {
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}

		if err == domain.ErrUnApproved {
			return ctx.JSON(http.StatusNotAcceptable, err.Error())
		}

		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	refreshToken, err := ctx.Cookie("refreshToken")
	if err != nil || len(refreshToken.Value) == 0 {
		cookie := new(http.Cookie)
		cookie.Name = "refreshToken"
		cookie.Value = jwtToken.RefreshToken
		cookie.HttpOnly = true
		cookie.Path = "/"
		ctx.SetCookie(cookie)
	} else {
		refreshToken.Value = jwtToken.RefreshToken
		refreshToken.HttpOnly = true
		refreshToken.Path = "/"
		ctx.SetCookie(refreshToken)
	}

	result := map[string]string{}
	result["accessToken"] = jwtToken.AccessToken
	return ctx.JSON(http.StatusOK, result)
}

func (AuthController) CheckAuth(ctx echo.Context) error {
	refreshToken, err := ctx.Cookie("refreshToken")
	if err != nil || len(refreshToken.Value) == 0 {
		return ctx.JSON(http.StatusNotAcceptable, nil)
	}

	jwtAuthentication := security.JwtAuthentication{}
	if err := jwtAuthentication.ValidateToken(refreshToken.Value); err != nil {
		return ctx.JSON(http.StatusNotAcceptable, nil)
	}

	return ctx.JSON(http.StatusOK, nil)
}

func (AuthController) Logout(ctx echo.Context) error {
	cookie, err := ctx.Cookie("refreshToken")
	if err != nil {
		ctx.JSON(http.StatusOK, nil)
	}

	cookie.Value = ""
	cookie.HttpOnly = true
	cookie.Path = "/"
	cookie.Expires = time.Unix(0, 0)
	cookie.MaxAge = -1
	ctx.SetCookie(cookie)

	return ctx.JSON(http.StatusOK, nil)
}

func (AuthController) RefreshAccessToken(ctx echo.Context) error {
	cookie, err := ctx.Cookie("refreshToken")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, nil)
	}

	refreshToken := cookie.Value

	jwtAuthentication := security.JwtAuthentication{}
	accessToken, err := jwtAuthentication.RefreshAccessToken(refreshToken)

	result := map[string]string{}
	result["accessToken"] = accessToken
	return ctx.JSON(http.StatusOK, result)
}

func (controller AuthController) AuthWithDoorayIdPassword(ctx echo.Context) error {
	var memberSignIn dtos.MemberSignIn

	if err := ctx.Bind(&memberSignIn); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	if err := memberSignIn.Validate(ctx); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	jwtToken, err := auth.AuthService{}.AuthWithDoorayIdAndPassword(ctx.Request().Context(), memberSignIn)
	if err != nil {
		if err == domain.ErrAuthentication {
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	refreshToken, err := ctx.Cookie("refreshToken")
	if err != nil || len(refreshToken.Value) == 0 {
		cookie := new(http.Cookie)
		cookie.Name = "refreshToken"
		cookie.Value = jwtToken.RefreshToken
		cookie.HttpOnly = true
		cookie.Path = "/"
		ctx.SetCookie(cookie)
	} else {
		refreshToken.Value = jwtToken.RefreshToken
		refreshToken.HttpOnly = true
		refreshToken.Path = "/"
		ctx.SetCookie(refreshToken)
	}

	result := map[string]string{}
	result["accessToken"] = jwtToken.AccessToken
	return ctx.JSON(http.StatusOK, result)
}

func (AuthController) AuthWithGoogleWorkspaceAccount(ctx echo.Context) error {
	code := ctx.QueryParam("code")
	redirect := ctx.QueryParam("state")

	jwtToken, err := auth.AuthService{}.AuthWithGoogleWorkspaceAccount(ctx.Request().Context(), code)
	if err != nil {
		if e, ok := err.(*domain.ErrInvalidGoogleWorkspaceAccount); ok {
			return ctx.Redirect(http.StatusFound, redirect+fmt.Sprintf("&error=%v 로 끝나는 메일 주소만 사용 가능 합니다.", e.Domain))
		}

		return ctx.Redirect(http.StatusFound, redirect+"&error=server-internal-error")
	}

	refreshToken, err := ctx.Cookie("refreshToken")
	if err != nil || len(refreshToken.Value) == 0 {
		cookie := new(http.Cookie)
		cookie.Name = "refreshToken"
		cookie.Value = jwtToken.RefreshToken
		cookie.HttpOnly = true
		cookie.Path = "/"
		ctx.SetCookie(cookie)
	} else {
		refreshToken.Value = jwtToken.RefreshToken
		refreshToken.HttpOnly = true
		refreshToken.Path = "/"
		ctx.SetCookie(refreshToken)
	}

	return ctx.Redirect(http.StatusFound, redirect+"&accessToken="+jwtToken.AccessToken)
}
