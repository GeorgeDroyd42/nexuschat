const NavigationUtils = {
    redirectToMain() {
        window.location.href = '/v/main';
    },
    
    redirectToLogin(message = null) {
        if (message) {
            sessionStorage.setItem('sessionMessage', message);
        }
        window.location.href = '/login';
    },
    
    redirectToRegister() {
        window.location.href = '/register';
    },
    
    redirectToGuild(guildId) {
        window.location.href = `/v/${guildId}`;
    },
    
    redirectToChannel(guildId, channelId) {
        window.location.href = `/v/${guildId}/${channelId}`;
    },
    
    redirectToInvite(guildId) {
        window.location.href = `/i/${guildId}`;
    }
};

window.NavigationUtils = NavigationUtils;