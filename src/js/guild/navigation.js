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
            MessageAPI.init();
            MessageAPI.loadChannelMessages(channelId);
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
            const channelsData = await API.guild.getChannels(guildId);
            
            window.channelManager.loadChannels(guildId);
                        
            if (channelsData.channels && channelsData.channels.length > 0) {
                window.channelManager.focusedChannel = channelsData.channels[0].channel_id;
                await window.channelManager.handleChannelSelect(channelsData.channels[0]);
            } else {
                await this.loadGuildContent(guildId);
            }
                    
            setActiveGuild(guildId);
            window.GuildMembers.loadGuildMembers(guildId);

        } catch (error) {
            NavigationUtils.redirectToGuild(guildId);
        }
    }
};

window.GuildNavigation = GuildNavigation;