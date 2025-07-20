package cache

func (c *RedisCache) GetAdmin(userID string) (bool, bool, error) {
	cacheKey := c.keys.Admin(userID)
	var isAdmin bool
	found, err := c.Get(cacheKey, &isAdmin)
	if found && err == nil {
		return isAdmin, true, nil
	}

	if c.dbProvider != nil {
		isAdmin, found, err := c.dbProvider.GetAdminFromDB(userID)
		if err != nil {
			return false, false, err
		}
		if found {
			c.Set(cacheKey, isAdmin, c.config.AdminTTL)
			return isAdmin, true, nil
		}
	}

	return false, false, nil
}

func (c *RedisCache) SetAdmin(userID string, isAdmin bool) error {
	cacheKey := c.keys.Admin(userID)
	return c.Set(cacheKey, isAdmin, c.config.AdminTTL)
}

func (c *RedisCache) DeleteAdmin(userID string) error {
	cacheKey := c.keys.Admin(userID)
	return c.Delete(cacheKey)
}
