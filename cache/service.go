package cache

var Service = &cacheService{}

type cacheService struct{}

func (s *cacheService) CheckUserBan(userID string) (bool, error) {
	banned, found, err := Provider.GetUserBan(userID)
	if err != nil || !found {
		return false, err
	}
	return banned, nil
}

func (s *cacheService) CheckAdmin(userID string) (bool, error) {
	isAdmin, found, err := Provider.GetAdmin(userID)
	if err != nil {
		return false, err
	}
	if found {
		return isAdmin, nil
	}
	return false, nil
}
func (s *cacheService) UpdateGuildOnlineCount(userID string, isOnline bool) error {
	// Implementation will use the Provider methods
	return nil
}
