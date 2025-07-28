package utils

// IsUserOnline checks if user has any active WebSocket connections
func IsUserOnline(userID string) bool {
	WebSocketManager.Mu.RLock()
	defer WebSocketManager.Mu.RUnlock()
	
	for _, conn := range WebSocketManager.Connections {
		if conn.UserID == userID {
			return true
		}
	}
	return false
}

// GetOnlineUsersInGuild returns list of online userIDs for a specific guild
func GetOnlineUsersInGuild(guildID string, allGuildMembers []string) []string {
	WebSocketManager.Mu.RLock()
	defer WebSocketManager.Mu.RUnlock()
	
	onlineUsers := make([]string, 0)
	onlineUserMap := make(map[string]bool)
	
	// Build map of online users
	for _, conn := range WebSocketManager.Connections {
		onlineUserMap[conn.UserID] = true
	}
	
	// Filter guild members who are online
	for _, memberID := range allGuildMembers {
		if onlineUserMap[memberID] {
			onlineUsers = append(onlineUsers, memberID)
		}
	}
	
	return onlineUsers
}