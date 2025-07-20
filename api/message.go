package api

import (
	"auth.com/v4/utils"
	"github.com/labstack/echo/v4"
)

func GetChannelMessagesHandler(c echo.Context) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return err
	}

	channelID, err := utils.RequireParam(c, "channelid")
	if err != nil {
		return err
	}

	valid, errCode := utils.ValidateChannelAccess(userID, channelID)
	if !valid {
		return utils.SendErrorResponse(c, errCode)
	}

	limit := 25
	beforeMessageID := c.QueryParam("before")

	messages, err := utils.GetChannelMessages(channelID, limit, beforeMessageID)
	if err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	return c.JSON(200, map[string]interface{}{
		"success":  true,
		"messages": messages,
		"has_more": len(messages) == limit,
	})
}
