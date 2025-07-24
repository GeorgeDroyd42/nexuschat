async function loadChannelContent(guildId, channelId) {
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
}



function setupMembersSidebar(guildId) {
    if (window.innerWidth > 768) {
        document.getElementById('members-sidebar').classList.add('visible');
        document.querySelector('.main-content').classList.add('with-members');
    }
    loadGuildMembers(guildId);
}

function createGuildElement(guild) {
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
        // Don't show tooltip if context menu is open
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


async function loadGuildContent(guildId) {
    const targetPath = `/v/${guildId}`;
    const html = await API.guild.getPage(guildId);
    const parser = new DOMParser();
    const doc = parser.parseFromString(html, 'text/html');
    const newContent = doc.querySelector('.main-content .container');
    
    if (newContent) {
        document.querySelector('.main-content .container').innerHTML = newContent.innerHTML;
        API.utils.processTimestamps(document.querySelector('.main-content .container'));
    }
    
    history.pushState({guildId: guildId}, '', targetPath);
    document.title = doc.title;
}

// Add both click and touch events for mobile compatibility
const chevron = guildElement.querySelector('.guild-chevron');
const channelsContainer = guildElement.querySelector('.guild-channels');

if (chevron) {
    chevron.addEventListener('click', async (e) => {
        e.stopPropagation();
        await toggleGuildChannels(guild.guild_id, chevron, channelsContainer);
    });
}

guildElement.querySelector('.guild-icon').addEventListener('click', async (e) => {
        e.preventDefault();
        e.stopPropagation();
        await window.switchToGuild(guild.guild_id);
    });
    guildElement.querySelector('.guild-icon').addEventListener('touchend', async (e) => {
        e.preventDefault();
        e.stopPropagation();
        await window.switchToGuild(guild.guild_id);
    });
    return guildElement;
}

function setupServerImageUpload() {
    setupImageUpload('server_picture', 'server-preview', 'select-server-btn');
}

document.addEventListener('DOMContentLoaded', () => {
    
    API.utils.processTimestamps(document);  
        window.addEventListener('popstate', async (e) => {
            if (e.state && e.state.guildId) {
                await window.forceNavigateToGuildChannel(e.state.guildId, e.state.channelId);
            }
        });      
    const currentGuildId = getCurrentGuildId();
    if (currentGuildId) {
        setTimeout(() => {
            const activeGuild = document.querySelector(`[data-guild-id="${currentGuildId}"]`);
            if (activeGuild) activeGuild.classList.add('active');
        }, 100);
    }    
    const createGuildBtn = $('create-guild-btn');
    const serverModal = $('server-modal');
    const closeServerModalBtn = $('close-server-modal');
    const createServerBtn = $('create-server-button');
    const backBtn = $('back-button');
    const guildForm = document.querySelector('#server-modal .form-group');
    const guildID = getCurrentGuildId();
if (guildID) {
    setupMembersSidebar(guildID);
    
    const channelId = getCurrentChannelId();
    if (channelId) {
        window.channelManager.focusedChannel = channelId;
        MessageAPI.init();
        MessageAPI.loadChannelMessages(channelId);
        
        const channelTitle = document.querySelector('.channel-title');
        const messageInput = document.getElementById('message-input');
        if (channelTitle && messageInput && window.getResponsiveChannelPlaceholder) {
            const channelName = channelTitle.textContent.replace('#', '').trim();
            messageInput.placeholder = window.getResponsiveChannelPlaceholder(channelName);
        }
    }
    window.channelManager.loadChannels(guildID);
}

window.forceNavigateToGuildChannel = async (guildId, channelId = null) => {
    try {
        const data = await API.guild.getChannels(guildId);
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
            const channelsData = await API.guild.getChannels(guildId);
            const channel = channelsData.channels?.find(c => c.channel_id === channelId);
            
            const html = await API.channel.getPage(guildId, channelId);
            const parser = new DOMParser();
            const doc = parser.parseFromString(html, 'text/html');
            const newContent = doc.querySelector('.main-content .container');
            
            if (newContent) {
                document.querySelector('.main-content .container').innerHTML = newContent.innerHTML;
                API.utils.processTimestamps(document.querySelector('.main-content .container'));
                MessageAPI.init();
                MessageAPI.loadChannelMessages(channelId);
                
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
            const html = await API.guild.getPage(guildId);
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
        setupMembersSidebar(guildId);
        loadGuildMembers(guildId);
        
    } catch (error) {
        console.error('Force navigation error:', error);
        if (channelId) {
            NavigationUtils.redirectToChannel(guildId, channelId);
        } else {
            NavigationUtils.redirectToGuild(guildId);
        }
    }
};
document.addEventListener('click', (e) => {
    if (e.target.classList.contains('add-channel-mobile')) {
        const guildElement = e.target.closest('[data-guild-id]');
        if (guildElement) {
            const guildId = guildElement.dataset.guildId;
            document.getElementById('channel-guild-id').value = guildId;
            const modal = document.getElementById('channel-modal');
            if (modal) window.modalManager.openModal(modal.id);
        }
        return;
    }
    
    if (window.innerWidth <= 768) {
        const membersSidebar = document.getElementById('members-sidebar');
        const membersToggle = document.getElementById('members-toggle');
        const guildSidebar = document.querySelector('.sidebar');
        const guildToggle = document.getElementById('guild-toggle');
        
        if (membersSidebar.classList.contains('visible') && 
            !membersSidebar.contains(e.target) && 
            !membersToggle.contains(e.target)) {
            membersSidebar.classList.remove('visible');
        }
        
        const activeModal = document.querySelector('.modal-overlay.active');
        const contextMenu = document.querySelector('.context-menu');
        const isContextMenuVisible = contextMenu && contextMenu.style.display === 'block';
        
        if (guildSidebar.classList.contains('mobile-visible') && 
            !guildSidebar.contains(e.target) && 
            !guildToggle.contains(e.target) &&
            !activeModal &&
            !isContextMenuVisible &&
            !e.target.closest('.context-menu')) {
            guildSidebar.classList.remove('mobile-visible');
        }
    }
});

document.getElementById('members-toggle').addEventListener('click', () => {
    document.getElementById('members-sidebar').classList.toggle('visible');
}); 

document.getElementById('guild-toggle').addEventListener('click', () => {
    document.querySelector('.sidebar').classList.toggle('mobile-visible');
}); 

    
    window.modalManager.setupModal('server-modal', 'create-guild-btn', 'back-button');
    window.modalManager.setupModal('channel-modal', 'create-channel-btn', 'cancel-channel-button');
    window.modalManager.setupModal('confirm-modal', null, 'close-invite-modal');

const copyInviteBtn = document.getElementById('copy-invite-btn');
if (copyInviteBtn) {
    copyInviteBtn.addEventListener('click', () => {
        const inviteText = document.getElementById('invite-link-text');
        if (inviteText) {
            navigator.clipboard.writeText(inviteText.value);
            
copyInviteBtn.textContent = 'Copied!';
copyInviteBtn.classList.add('copied');

setTimeout(() => {
    copyInviteBtn.textContent = 'Copy';
    copyInviteBtn.classList.remove('copied');
}, 1500);
        }
    });
}

const createChannelBtn = document.getElementById('create-channel-btn');
if (createChannelBtn) {
    createChannelBtn.addEventListener('click', () => {
        const currentGuildId = getCurrentGuildId();
        if (currentGuildId) {
            document.getElementById('channel-guild-id').value = currentGuildId;
        }
    });
}
    window.modalManager.setupModal('channel-info-modal', null, 'close-channel-info-modal');
    window.modalManager.setupModal('guild-settings-modal', null, 'close-guild-settings-modal');

    
    createServerBtn.addEventListener('click', async (e) => {
        e.preventDefault();
        
        await handleFormSubmission({
            formElement: $('create-guild-form'),
            apiFunction: API.guild.create,
            errorContainerId: 'guild-error-container',
            validateForm: () => $('server-name').value.trim() !== '',
            operationName: 'guild creation',
onSuccess: () => {
    window.modalManager.closeModal('server-modal');
    clearFormFields(['server-name', 'server-description', 'server_picture', 'server-preview'], {
        'server-preview': "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%23cccccc'%3E%3Cpath d='M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z'/%3E%3C/svg%3E"
    });
    const errorContainer = $('guild-error-container');
    if (errorContainer) {
        errorContainer.style.display = 'none';
    }
}
        });
    });

    

    initWebSocket();
    setupServerImageUpload();
});


const settingsBtn = document.getElementById('settings-btn');
        const profileModal = document.getElementById('profile-modal');
        const closeProfileBtn = document.getElementById('close-profile-modal');
        
        if (settingsBtn) {
            settingsBtn.addEventListener('click', () => {
                window.profileManager.openProfile(true);
            });
        }
    
    if (closeProfileBtn) {
        closeProfileBtn.addEventListener('click', () => {
            window.profileManager.closeProfile();
        });
    }

async function loadGuildMembers(guildID) {
    try {
        const data = await API.guild.getMembers(guildID);
        
        if (data.error) {
            console.error('Error loading members:', data.error);
            return;
        }
        
        updateMembersList(data.members, guildID);
    } catch (error) {
        console.error('Error loading guild members:', error);
    }
}

async function getUsernameByID(userID) {
    try {
        const data = await UserAPI.getUserProfile(userID);
        return data.username || 'Unknown';
    } catch {
        return 'Unknown';
    }
}
async function getUserProfilePicture(userID) {
    try {
        const data = await UserAPI.getUserProfile(userID);
        return data.profile_picture || '';
    } catch {
        return '';
    }
}
window.refreshGuildList = async function() {
    try {
        const data = await API.guild.fetchUserGuilds();
        
        const guildList = document.getElementById('guild-list');
        if (guildList && data.guilds) {
            guildList.innerHTML = ''; // Clear existing
            data.guilds.forEach(guild => {
                const guildElement = createGuildElement(guild);
                guildList.appendChild(guildElement);
            });
        }
    } catch (error) {
        console.error('Error refreshing guild list:', error);
    }
}

async function toggleGuildChannels(guildId, chevron, channelsContainer) {
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
}

    
    window.switchToGuild = async (guildId) => {
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
                await loadGuildContent(guildId);
            }
                    
            setActiveGuild(guildId);
            loadGuildMembers(guildId);

        } catch (error) {
            NavigationUtils.redirectToGuild(guildId);
        }
    };
    