const GuildManager = {
    createGuildElement(guild) {
        const template = document.getElementById('guild-template');
        const guildElement = template.content.cloneNode(true).querySelector('.guild-pill');
        guildElement.setAttribute('data-guild-id', guild.guild_id);
        
        const guildIcon = guildElement.querySelector('.guild-icon');
        const guildAvatar = window.AvatarUtils.createSecureAvatar(guild.name, guild.profile_picture_url, 'avatar-circle');
        guildIcon.innerHTML = '';
        guildIcon.appendChild(guildAvatar);
        
        if (window.innerWidth > 768) {
            guildElement.addEventListener('mouseenter', () => {
                if (window.ctxMenu && window.ctxMenu.isOpen) return;
                const tooltip = document.createElement('div');
                tooltip.className = 'guild-tooltip';
                tooltip.textContent = guild.name;
                const rect = guildElement.getBoundingClientRect();
                tooltip.style.left = rect.right - 50 + 'px';
                tooltip.style.top = rect.top + rect.height / 2 + 'px';
                tooltip.style.transform = 'translateY(-50%)';
                tooltip.style.opacity = '1';
                tooltip.style.visibility = 'visible';
                document.body.appendChild(tooltip);
                guildElement.currentTooltip = tooltip;
            });

            guildElement.addEventListener('mouseleave', () => {
                if (guildElement.currentTooltip) {
                    guildElement.currentTooltip.remove();
                    guildElement.currentTooltip = null;
                }
            });   

            guildElement.addEventListener('contextmenu', () => {
                if (guildElement.currentTooltip) {
                    guildElement.currentTooltip.remove();
                    guildElement.currentTooltip = null;
                }
            });
        }

        const chevron = guildElement.querySelector('.guild-chevron');
        const channelsContainer = guildElement.querySelector('.guild-channels');

        if (chevron) {
            chevron.addEventListener('click', async (e) => {
                e.stopPropagation();
                await window.GuildUI.toggleGuildChannels(guild.guild_id, chevron, channelsContainer);
            });
        }

        guildElement.querySelector('.guild-icon').addEventListener('click', async (e) => {
            e.preventDefault();
            e.stopPropagation();
            await window.GuildNavigation.switchToGuild(guild.guild_id);
        });
        guildElement.querySelector('.guild-icon').addEventListener('touchend', async (e) => {
            e.preventDefault();
            e.stopPropagation();
            await window.GuildNavigation.switchToGuild(guild.guild_id);
        });
        return guildElement;
    }
    
};

window.GuildManager = GuildManager;

// initialize guild functionality when DOM is ready
document.addEventListener('DOMContentLoaded', async () => {
    const currentGuildId = getCurrentGuildId();
    if (currentGuildId) {
        GuildUI.highlightActiveGuild(currentGuildId);        
        window.GuildMembers.setupMembersSidebar(currentGuildId);
        window.ChannelUI.loadChannels(currentGuildId, window.channelManager);
    }
    
    await updateChannelsHeader(currentGuildId);
    window.GuildUI.setupServerImageUpload();
});