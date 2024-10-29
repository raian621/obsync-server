package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/raian621/obsync-server/api"
	"github.com/raian621/obsync-server/database"
)

// Delete a user
// (DELETE /user)
func (o *ObsyncServer) DeleteUser(ctx echo.Context) error {
	sessionCookie, err := ctx.Cookie("OBSYNC_SESSION_ID")
	if err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	session, err := database.GetSessionBySessionKey(o.db, sessionCookie.Value)
	if err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	if err := database.DeleteUser(o.db, session.UserId); err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	return sendApiMessage(ctx, http.StatusOK, "user deleted")
}

// Create a user
// (POST /user)
func (o *ObsyncServer) PostUser(ctx echo.Context) error {
	var user api.User

	err := json.NewDecoder(ctx.Request().Body).Decode(&user)
	if err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	// create user
	_, err = database.CreateUser(o.db, user.Username, string(user.Email), user.Password)
	if err != nil {
		ctx.Logger().Print(err)
		switch err {
		case database.ErrUsernameFormat:
			return sendApiMessage(ctx, http.StatusBadRequest, "invalid username")
		case database.ErrEmailFormat:
			return sendApiMessage(ctx, http.StatusBadRequest, "invalid email")
		case database.ErrPasswordLength:
			return sendApiMessage(ctx, http.StatusBadRequest, "password too short")
		default:
			return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
		}
	}

	return ctx.JSON(
		http.StatusOK,
		map[string]any{
			"username": user.Username,
			"email":    user.Email,
			"id":       user.Id,
		},
	)
}

// Log in a user
// (POST /user/login)
func (o *ObsyncServer) PostUserLogin(ctx echo.Context) error {
	var credentials api.PostUserLoginJSONBody
	encoder := json.NewDecoder(ctx.Request().Body)
	if err := encoder.Decode(&credentials); err != nil {
		ctx.Logger().Error(err)
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	// check if the user exists
	user, err := database.GetUserByUsername(o.db, credentials.Username)
	if err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusNotFound, "username or password invalid")
	}

	// check if the password is correct
	if err := database.ValidateHash(credentials.Password, user.Passhash); err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusNotFound, "username or password invalid")
	}

	// create user session if user is authenticated
	session, err := database.CreateSession(o.db, user.Id)
	if err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	ctx.SetCookie(&http.Cookie{
		Name:     "OBSYNC_SESSION_ID",
		Value:    session.SessionKey,
		Expires:  session.Expires,
		HttpOnly: true,
	})

	return ctx.NoContent(http.StatusOK)
}

// Log out a user
// (POST /user/logout)
func (o *ObsyncServer) PostUserLogout(ctx echo.Context) error {
	cookie, err := ctx.Cookie("OBSYNC_SESSION_ID")
	if err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	if err := database.DeleteSession(o.db, cookie.Value); err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	return nil
}

// Let users update their email
// (PUT /user/email)
func (o *ObsyncServer) PutUserEmail(ctx echo.Context) error {
	data, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}
	email := string(data)

	sessionCookie, err := ctx.Cookie("OBSYNC_SESSION_ID")
	if err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusUnauthorized, "not authenticated")
	}

	session, err := database.GetSessionBySessionKey(o.db, sessionCookie.Value)
	if err != nil {
		ctx.Logger().Print(err)
		if err == database.ErrExpiredSession {
			return sendApiMessage(ctx, http.StatusUnauthorized, "not authenticated")
		}
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	err = database.UpdateUserEmail(o.db, session.UserId, email)
	if err != nil {
		ctx.Logger().Print(err)
		if err == database.ErrEmailFormat {
			return sendApiMessage(ctx, http.StatusBadRequest, "invalid email")
		}
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	return sendApiMessage(ctx, http.StatusOK, "email updated")
}

// Let users update their password
// (PUT /user/password)
func (o *ObsyncServer) PutUserPassword(ctx echo.Context) error {
	data, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}
	password := string(data)

	sessionCookie, err := ctx.Cookie("OBSYNC_SESSION_ID")
	if err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusUnauthorized, "not authenticated")
	}

	session, err := database.GetSessionBySessionKey(o.db, sessionCookie.Value)
	if err != nil {
		ctx.Logger().Print(err)
		if err == database.ErrExpiredSession {
			return sendApiMessage(ctx, http.StatusUnauthorized, "not authenticated")
		}
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	err = database.UpdateUserPassword(o.db, session.UserId, password)
	if err != nil {
		ctx.Logger().Print(err)
		if err == database.ErrPasswordLength {
			return sendApiMessage(ctx, http.StatusBadRequest, "password too short")
		}
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	return sendApiMessage(ctx, http.StatusOK, "password updated")
}

// Let users update their username
// (PUT /user/username)
func (o *ObsyncServer) PutUserUsername(ctx echo.Context) error {
	data, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}
	username := string(data)

	sessionCookie, err := ctx.Cookie("OBSYNC_SESSION_ID")
	if err != nil {
		ctx.Logger().Print(err)
		return sendApiMessage(ctx, http.StatusUnauthorized, "not authenticated")
	}

	session, err := database.GetSessionBySessionKey(o.db, sessionCookie.Value)
	if err != nil {
		ctx.Logger().Print(err)
		if err == database.ErrExpiredSession {
			return sendApiMessage(ctx, http.StatusUnauthorized, "not authenticated")
		}
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	err = database.UpdateUserUsername(o.db, session.UserId, username)
	if err != nil {
		ctx.Logger().Print(err)
		if err == database.ErrUsernameFormat {
			return sendApiMessage(ctx, http.StatusBadRequest, "invalid username")
		}
		return sendApiMessage(ctx, http.StatusInternalServerError, "unexpected error occurred")
	}

	return sendApiMessage(ctx, http.StatusOK, "username updated")
}
