const ChannelHandlers = {
    async handleChannelSelect(channel, channelManager) {
        const guildId = channel.guild_id || channelManager.currentGuildId;
        const targetPath = `/v/${guildId}/${channel.channel_id}`;
        
        if (isCurrentPath(targetPath)) {
            return;
        }
        
        if (guildId !== channelManager.currentGuildId) {
            await window.GuildNavigation.switchToGuild(guildId);
            return;
        }
        
        try {
            const html = await API.channel.getPage(guildId, channel.channel_id);
            const parser = new DOMParser();
            const doc = parser.parseFromString(html, 'text/html');
            const newContent = doc.querySelector('.main-content .container');
            
            if (newContent) {
                document.querySelector('.main-content .container').innerHTML = newContent.innerHTML;
                
                API.utils.processTimestamps(document.querySelector('.main-content .container'));
                if (window.MessageAPI) {
                    window.MessageAPI.init();
                    if (channel.channel_id) {
                        window.MessageAPI.loadChannelMessages(channel.channel_id);
                    }
                }
                
                const messageInput = document.getElementById('message-input');
                if (messageInput && window.getResponsiveChannelPlaceholder) {
                    messageInput.placeholder = window.getResponsiveChannelPlaceholder(channel.name);
                }
            }
            
            history.pushState({channelId: channel.channel_id, guildId: guildId}, '', targetPath);
            document.title = doc.title;
            
            channelManager.focusedChannel = channel.channel_id;
            window.setActiveChannel(channel.channel_id);
            if (window.innerWidth > 768) {
                document.getElementById('members-sidebar').classList.add('visible');
                document.querySelector('.main-content').classList.add('with-members');
            }
            
        } catch (error) {
            NavigationUtils.redirectToChannel(guildId, channel.channel_id);
        }
    }
};

window.ChannelHandlers = ChannelHandlers;