const GuildManager = {
    createGuildElement(guild) {
        const template = document.getElementById('guild-template');
        const guildElement = template.content.cloneNode(true).querySelector('.guild-pill');
        guildElement.setAttribute('data-guild-id', guild.guild_id);
        
        const guildIcon = guildElement.querySelector('.guild-icon');
        const guildIconHtml = guild.profile_picture_url && guild.profile_picture_url.trim() !== '' ?
            `<img src="${guild.profile_picture_url}" alt="${guild.name}" class="guild-image">` :
            (window.AvatarUtils ? window.AvatarUtils.show404guild(guild.name) : `<span class="guild-initial">${guild.name.charAt(0).toUpperCase()}</span>`);
        guildIcon.innerHTML = guildIconHtml;

        const imgEl = guildIcon.querySelector('img.guild-image');
        if (imgEl) {
            imgEl.onerror = function() {
                this.outerHTML = window.AvatarUtils ? window.AvatarUtils.show404guild(guild.name) : `<span class="guild-initial">${guild.name.charAt(0).toUpperCase()}</span>`;
            };
        }
        
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