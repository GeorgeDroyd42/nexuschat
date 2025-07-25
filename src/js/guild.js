document.addEventListener('DOMContentLoaded', () => {
    
    API.utils.processTimestamps(document);  
        window.addEventListener('popstate', async (e) => {
            if (e.state && e.state.guildId) {
                await window.GuildNavigation.forceNavigateToGuildChannel(e.state.guildId, e.state.channelId);
            }
        });      
    const currentGuildId = getCurrentGuildId();
    if (currentGuildId) {
        GuildUI.highlightActiveGuild(currentGuildId);
        
        window.GuildMembers.setupMembersSidebar(currentGuildId);
    const channelId = getCurrentChannelId();
        if (channelId) {
            window.channelManager.focusedChannel = channelId;
            MessageUI.init();
            MessageManager.loadChannelMessages(channelId);
            
            const channelTitle = document.querySelector('.channel-title');
            const messageInput = document.getElementById('message-input');
            if (channelTitle && messageInput && window.getResponsiveChannelPlaceholder) {
                const channelName = channelTitle.textContent.replace('#', '').trim();
                messageInput.placeholder = window.getResponsiveChannelPlaceholder(channelName);
            }
        }
        window.ChannelUI.loadChannels(currentGuildId, window.channelManager);
    }
    initWebSocket();
    window.GuildUI.setupServerImageUpload();
    GuildButtons.init();
});
    