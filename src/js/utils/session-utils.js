async function handleLogout() {
    try {
        const response = await AuthAPI.logout();
        
        if (response.ok) {
            NavigationUtils.redirectToLogin();
        }
        return response;
    } catch (error) {
        console.error('Error logging out:', error);
    }    
}

function startSessionRefresh() {
    const refreshInterval = 12 * 60 * 60 * 1000; // 4 seconds
    
    async function doRefreshWithRetry(retryCount = 0) {
        try {
            const response = await AuthAPI.refreshSession();
            
            if (response.ok) {
                const data = await response.json();
                console.log('Session refreshed successfully, expires at:', new Date(data.expires_at * 1000));
            } else {
                console.log('Session refresh failed:', response.status);
                if (retryCount < 2) {
                    const retryDelay = retryCount === 0 ? 5 * 60 * 1000 : 15 * 60 * 1000;
                    setTimeout(() => doRefreshWithRetry(retryCount + 1), retryDelay);
                }
            }
        } catch (error) {
            console.error('Session refresh error:', error);
            if (retryCount < 2) {
                const retryDelay = retryCount === 0 ? 5 * 60 * 1000 : 15 * 60 * 1000;
                setTimeout(() => doRefreshWithRetry(retryCount + 1), retryDelay);
            }
        }
    }
    
    setInterval(() => doRefreshWithRetry(), refreshInterval);
}

function registerCommonEventListeners() {
  const logoutBtn = $('logout-btn');
  if (logoutBtn) {
    logoutBtn.addEventListener('click', (e) => {
      e.preventDefault();
      handleLogout();
    });
  }
  
  const mobileLogoutBtn = $('mobile-logout-btn');
  if (mobileLogoutBtn) {
    mobileLogoutBtn.addEventListener('click', (e) => {
      e.preventDefault();
      handleLogout();
    });
  }
}