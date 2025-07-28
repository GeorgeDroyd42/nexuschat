const GuildNavigation = {
    async navigateToGuild(guildId, channelId = null) {
        try {
            const guildList = $('guild-list');
            if (guildList) {
                sessionStorage.setItem('guildListScrollPosition', guildList.scrollTop);
            }
            
            const channelsData = await window.GuildAPI.getChannels(guildId);
            await window.ChannelUI.loadChannels(guildId, window.channelManager);
            
            if (channelId) {
                const channel = channelsData.channels?.find(c => c.channel_id === channelId);
                const html = await window.ChannelAPI.getPage(guildId, channelId);
                const parser = new DOMParser();
                const doc = parser.parseFromString(html, 'text/html');
                const newContent = doc.querySelector('.main-content .container');
                
                if (newContent) {
                    document.querySelector('.main-content .container').innerHTML = newContent.innerHTML;
                    window.API.utils.processTimestamps(document.querySelector('.main-content .container'));
                    window.MessageUI.init();
                    window.MessageManager.loadChannelMessages(channelId);
                    
                    const messageInput = document.getElementById('message-input');
                    if (messageInput && channel && window.getResponsiveChannelPlaceholder) {
                        messageInput.placeholder = window.getResponsiveChannelPlaceholder(channel.name);
                    }
                }
                
                window.channelManager.focusedChannel = channelId;
                setActiveChannel(channelId);
                history.pushState({guildId: guildId, channelId: channelId}, '', `/v/${guildId}/${channelId}`);
                
                if (window.innerWidth > 768) {
                    document.getElementById('members-sidebar').classList.add('visible');
                    document.querySelector('.main-content').classList.add('with-members');
                }
            } else if (channelsData.channels && channelsData.channels.length > 0) {
                const firstChannel = channelsData.channels[0];
                await this.navigateToGuild(guildId, firstChannel.channel_id);
                return;
            } else {
                const html = await window.GuildAPI.getPage(guildId);
                const parser = new DOMParser();
                const doc = parser.parseFromString(html, 'text/html');
                const newContent = doc.querySelector('.main-content .container');
                
                if (newContent) {
                    document.querySelector('.main-content .container').innerHTML = newContent.innerHTML;
                    window.API.utils.processTimestamps(document.querySelector('.main-content .container'));
                }
                
                history.pushState({guildId: guildId}, '', `/v/${guildId}`);
            }
            
            document.title = document.querySelector('title')?.textContent || `Guild ${guildId}`;
            setActiveGuild(guildId);
            window.GuildMembers.setupMembersSidebar(guildId);
            window.GuildMembers.loadGuildMembers(guildId);
            
        } catch (error) {
            console.error('Navigation error:', error);
            if (channelId) {
                window.NavigationUtils.redirectToChannel(guildId, channelId);
            } else {
                window.NavigationUtils.redirectToGuild(guildId);
            }
        }
    },

    async switchToGuild(guildId) {
        if (isCurrentGuild(guildId)) {
            return;
        }
        await this.navigateToGuild(guildId);
    },
};

window.GuildNavigation = GuildNavigation;

document.addEventListener('DOMContentLoaded', () => {
    window.addEventListener('popstate', async (e) => {
        if (e.state && e.state.guildId) {
            await window.GuildNavigation.navigateToGuild(e.state.guildId, e.state.channelId);
        }
    });
});