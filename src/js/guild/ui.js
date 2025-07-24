const GuildUI = {
    async refreshGuildList() {
        try {
            const data = await API.guild.fetchUserGuilds();
            
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
                const data = await API.guild.getChannels(guildId);
                window.channelManager.updateMobileChannelsForGuild(guildId, data.channels, channelsContainer);
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
    }
};

window.GuildUI = GuildUI;