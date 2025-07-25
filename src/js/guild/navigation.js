const GuildNavigation = {
    async loadChannelContent(guildId, channelId) {
        const targetPath = `/v/${guildId}/${channelId}`;
        const html = await API.channel.getPage(guildId, channelId);
        const parser = new DOMParser();
        const doc = parser.parseFromString(html, 'text/html');
        const newContent = doc.querySelector('.main-content .container');
        
        if (newContent) {
            document.querySelector('.main-content .container').innerHTML = newContent.innerHTML;
            API.utils.processTimestamps(document.querySelector('.main-content .container'));
            MessageUI.init();
            MessageManager.loadChannelMessages(channelId);
        }
        
        history.pushState({channelId: channelId, guildId: guildId}, '', targetPath);
        document.title = doc.title;    
    },

    async switchToGuild(guildId) {
        if (isCurrentGuild(guildId)) {
            return;
        }
        
        const guildList = $('guild-list');
        if (guildList) {
            sessionStorage.setItem('guildListScrollPosition', guildList.scrollTop);
        }
        
        try {
            const channelsData = await GuildAPI.getChannels(guildId);
            
            await window.ChannelUI.loadChannels(guildId, window.channelManager);
                        
            if (channelsData.channels && channelsData.channels.length > 0) {
                window.channelManager.focusedChannel = channelsData.channels[0].channel_id;
                await window.ChannelHandlers.handleChannelSelect(channelsData.channels[0], window.channelManager);
            } else {
                await this.loadGuildContent(guildId);
            }
                    
            setActiveGuild(guildId);
            window.GuildMembers.loadGuildMembers(guildId);

        } catch (error) {
            NavigationUtils.redirectToGuild(guildId);
        }
    },

    async forceNavigateToGuildChannel(guildId, channelId = null) {
        try {
            const data = await GuildAPI.getChannels(guildId);
            if (data.channels) {
                const channelsList = document.getElementById('channels-list');
                if (channelsList) {
                    channelsList.innerHTML = '';
                    data.channels.forEach(channel => {
                        const channelElement = document.createElement('div');
                        channelElement.className = 'channel-item';
                        channelElement.innerHTML = `<span class="channel-name">#${channel.name}</span>`;
                        channelElement.addEventListener('click', () => {
                            window.location.href = `/v/${guildId}/${channel.channel_id}`;
                        });
                        channelsList.appendChild(channelElement);
                    });
                }
            }
            
            if (channelId) {
                const channelsData = await GuildAPI.getChannels(guildId);
                const channel = channelsData.channels?.find(c => c.channel_id === channelId);
                
                const html = await API.channel.getPage(guildId, channelId);
                const parser = new DOMParser();
                const doc = parser.parseFromString(html, 'text/html');
                const newContent = doc.querySelector('.main-content .container');
                
                if (newContent) {
                    document.querySelector('.main-content .container').innerHTML = newContent.innerHTML;
                    API.utils.processTimestamps(document.querySelector('.main-content .container'));
                    MessageUI.init();
                    MessageManager.loadChannelMessages(channelId);
                    
                    const messageInput = document.getElementById('message-input');
                    if (messageInput && channel && window.getResponsiveChannelPlaceholder) {
                        messageInput.placeholder = window.getResponsiveChannelPlaceholder(channel.name);
                    }
                }
                
                document.title = doc.title;
                window.channelManager.focusedChannel = channelId;
                setActiveChannel(channelId);
                
                if (window.innerWidth > 768) {
                    document.getElementById('members-sidebar').classList.add('visible');
                    document.querySelector('.main-content').classList.add('with-members');
                }
            } else {
                const html = await GuildAPI.getPage(guildId);
                const parser = new DOMParser();
                const doc = parser.parseFromString(html, 'text/html');
                const newContent = doc.querySelector('.main-content .container');
                
                if (newContent) {
                    document.querySelector('.main-content .container').innerHTML = newContent.innerHTML;
                    API.utils.processTimestamps(document.querySelector('.main-content .container'));
                }
                
                document.title = doc.title;
            }
            
            setActiveGuild(guildId);
            window.GuildMembers.setupMembersSidebar(guildId);
            window.GuildMembers.loadGuildMembers(guildId);
            
        } catch (error) {
            console.error('Force navigation error:', error);
            if (channelId) {
                NavigationUtils.redirectToChannel(guildId, channelId);
            } else {
                NavigationUtils.redirectToGuild(guildId);
            }
        }
    },    
};

window.GuildNavigation = GuildNavigation;

document.addEventListener('DOMContentLoaded', () => {
    window.addEventListener('popstate', async (e) => {
        if (e.state && e.state.guildId) {
            await window.GuildNavigation.forceNavigateToGuildChannel(e.state.guildId, e.state.channelId);
        }
    });
});