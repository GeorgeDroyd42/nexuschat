


document.addEventListener('DOMContentLoaded', () => {
    
    API.utils.processTimestamps(document);  
        window.addEventListener('popstate', async (e) => {
            if (e.state && e.state.guildId) {
                await window.GuildNavigation.forceNavigateToGuildChannel(e.state.guildId, e.state.channelId);
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
    window.GuildMembers.setupMembersSidebar(guildID);
    
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
    window.ChannelUI.loadChannels(guildID, window.channelManager);
}


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

 

    
    window.modalManager.setupModal('server-modal', 'create-guild-btn', 'back-button');
    window.modalManager.setupModal('channel-modal', 'create-channel-btn', 'cancel-channel-button');
    window.modalManager.setupModal('confirm-modal', null, 'close-invite-modal');


    window.modalManager.setupModal('channel-info-modal', null, 'close-channel-info-modal');
    window.modalManager.setupModal('guild-settings-modal', null, 'close-guild-settings-modal');

    
    createServerBtn.addEventListener('click', async (e) => {
        e.preventDefault();
        
        await handleFormSubmission({
            formElement: $('create-guild-form'),
            apiFunction: GuildAPI.create,
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
    window.GuildUI.setupServerImageUpload();
});


GuildButtons.init();
    