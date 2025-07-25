const GuildUI = {
    async refreshGuildList() {
        try {
            const data = await GuildAPI.fetchUserGuilds();
            
            const guildList = document.getElementById('guild-list');
            if (guildList && data.guilds) {
                guildList.innerHTML = '';
                data.guilds.forEach(guild => {
                    const guildElement = window.GuildManager.createGuildElement(guild);
                    guildList.appendChild(guildElement);
                });
            }
        } catch (error) {
            console.error('Error refreshing guild list:', error);
        }
    },

    async toggleGuildChannels(guildId, chevron, channelsContainer) {
        const isExpanded = channelsContainer.style.display === 'block';
        
        if (!isExpanded) {
            try {
                const data = await GuildAPI.getChannels(guildId);
                window.ChannelUI.updateMobileChannelsForGuild(guildId, data.channels, window.channelManager);
                channelsContainer.style.display = 'block';
                chevron.classList.add('expanded');
            } catch (error) {
                console.error('Error loading channels:', error);
            }
        } else {
            channelsContainer.style.display = 'none';
            chevron.classList.remove('expanded');
        }
    },

    setupServerImageUpload() {
        setupImageUpload('server_picture', 'server-preview', 'select-server-btn');
    },

    highlightActiveGuild(guildId) {
        if (guildId) {
            setTimeout(() => {
                const activeGuild = document.querySelector(`[data-guild-id="${guildId}"]`);
                if (activeGuild) activeGuild.classList.add('active');
            }, 100);
        }
    }
};

window.GuildUI = GuildUI;