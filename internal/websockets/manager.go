package websockets

import (
	"github.com/sirupsen/logrus"
)


func SendToUser(userID string, messageType int, data []byte) {
	Manager.Mu.RLock()
	defer Manager.Mu.RUnlock()

	for _, conn := range Manager.Connections {
		if conn.UserID == userID {
			conn.WriteMu.Lock()
			err := conn.Conn.WriteMessage(messageType, data)
			conn.WriteMu.Unlock()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"component": "WebSocket",
					"action":    "send_to_user",
					"user_id":   userID,
				}).Error("Failed to send message to user: ", err)
			}
		}
	}
}

func BroadcastToAll(messageType int, data []byte) {
	Manager.Mu.RLock()
	var failedSessions []string

	for sessionID, conn := range Manager.Connections {
		if conn != nil && conn.Conn != nil {
			conn.WriteMu.Lock()
			err := conn.Conn.WriteMessage(messageType, data)
			conn.WriteMu.Unlock()
			if err != nil {
				failedSessions = append(failedSessions, sessionID)
			}
		}
	}
	Manager.Mu.RUnlock()

	if len(failedSessions) > 0 {
		Manager.Mu.Lock()
		for _, sessionID := range failedSessions {
			delete(Manager.Connections, sessionID)
		}
		Manager.Mu.Unlock()
	}
}