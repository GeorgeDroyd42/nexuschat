package api

import (
	"encoding/json"
	"fmt"
	"time"

	"auth.com/v4/utils"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func HandleAuthWebSocket(c echo.Context) error {
	userID := utils.GetUserID(c)

	wsConn, sessionID, err := utils.UpgradeAndRegister(c, userID) // Capture sessionID
	if err != nil {
		return err
	}
	if wsConn == nil {
		return nil
	}
	defer wsConn.Conn.Close()
	defer utils.RemoveConnection(sessionID)
	messageCount := 0
	resetTime := time.Now().Add(time.Minute)
	messageProcessCount := 0

	for {
		wsConn.Conn.SetReadDeadline(time.Now().Add(90 * time.Second))

		// Rate limiting check
		now := time.Now()
		if now.After(resetTime) {
			messageCount = 0
			resetTime = now.Add(time.Minute)
		}

		if messageCount >= utils.AppConfig.MaxWSMessagesPerMinute {
			wsConn.Conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"rate_limit_exceeded"}`))
			time.Sleep(time.Second)
			continue
		}

		messageType, message, err := wsConn.Conn.ReadMessage()
		if err != nil {
			break
		}

		sessionValid, err := utils.ValidateWebSocketSession(userID, sessionID)
		if err != nil || !sessionValid {
			wsConn.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, utils.ErrorMessages[utils.ErrUnauthorized]))
			wsConn.Conn.Close()
			return nil
		}
		currentUserID := userID

		isBanned, err := utils.IsUserBanned(currentUserID)
		if err == nil && isBanned {
			wsConn.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, utils.ErrorMessages[utils.ErrAccountSuspended]))
			wsConn.Conn.Close()
			return nil
		}

		if messageType == websocket.TextMessage {

			var jsonMsg map[string]interface{}
			if json.Unmarshal(message, &jsonMsg) != nil {
				wsConn.Conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid_message_format"}`))
				continue
			}

			messageProcessCount++

			err := utils.HandleWebSocketMessage(userID, message)
			if err != nil {
				errorMsg := fmt.Sprintf(`{"error":"%s"}`, err.Error())
				wsConn.Conn.WriteMessage(websocket.TextMessage, []byte(errorMsg))
				continue
			}
		}
		messageCount++
	}

	return nil
}

func TerminateSessionHandler(c echo.Context) error {
	sessionID := c.Param("sessionid")
	if sessionID == "" {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}

	utils.PerformLogoutBySessionID(sessionID)
	return utils.SendSuccessResponse(c, "Session terminated successfully")

}

func TerminateUserSessionsHandler(c echo.Context) error {
	userID := c.Param("userid")
	if userID == "" {
		return utils.SendErrorResponse(c, utils.ErrInvalidCredentials)
	}

	success, err := utils.TerminateAllUserSessions(userID)
	if !success || err != nil {
		return utils.SendErrorResponse(c, utils.ErrDatabaseError)
	}

	utils.ClearAuthCookie(c)
	return utils.SendSuccessResponse(c, "All sessions terminated for user: "+userID)
}
