const GuildButtons = {
    setupMembersToggle() {
        document.getElementById('members-toggle').addEventListener('click', () => {
            document.getElementById('members-sidebar').classList.toggle('visible');
        });
    },

    setupGuildToggle() {
        document.getElementById('guild-toggle').addEventListener('click', () => {
            document.querySelector('.sidebar').classList.toggle('mobile-visible');
        });
    },

    setupCopyInvite() {
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
    },

    setupCreateChannel() {
        const createChannelBtn = document.getElementById('create-channel-btn');
        if (createChannelBtn) {
            createChannelBtn.addEventListener('click', () => {
                const currentGuildId = getCurrentGuildId();
                if (currentGuildId) {
                    document.getElementById('channel-guild-id').value = currentGuildId;
                }
            });
        }
    },

    setupSettings() {
        const settingsBtn = document.getElementById('settings-btn');
        const closeProfileBtn = document.getElementById('close-profile-modal');
        
        if (settingsBtn) {
            settingsBtn.addEventListener('click', () => {
                window.modalManager.openModal('profile-modal');
                
                window.profileManager.loadUserData().then(userData => {
                    if (window.profileMenuAPI && userData) {
                        window.profileMenuAPI.renderButtons('profile-modal', userData);
                    }
                });
            });
        }
    
        if (closeProfileBtn) {
            closeProfileBtn.addEventListener('click', () => {
                window.modalManager.closeModal('profile-modal');
                
                const userModal = document.querySelector('#user-modal');
                if (userModal) userModal.classList.remove('active');
            });
        }
    },

    setupModals() {
        window.modalManager.setupModal('server-modal', 'create-guild-btn', 'back-button');
        window.modalManager.setupModal('channel-modal', 'create-channel-btn', 'cancel-channel-button');
        window.modalManager.setupModal('confirm-modal', null, 'close-invite-modal');
        window.modalManager.setupModal('channel-info-modal', null, 'close-channel-info-modal');
        window.modalManager.setupModal('guild-settings-modal', null, 'close-guild-settings-modal');
    },

    setupCreateServer() {
        const createServerBtn = $('create-server-button');
        if (createServerBtn) {
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
        }
    },

    setupMobileHandlers() {
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
    },

    init() {
        this.setupModals();
        this.setupMembersToggle();
        this.setupGuildToggle();
        this.setupCopyInvite();
        this.setupCreateChannel();
        this.setupSettings();
        this.setupCreateServer();
        this.setupMobileHandlers();
        
        CharCountAPI.addMultiple([
            { id: 'server-description', options: { maxLength: 500, warningThreshold: 450, errorThreshold: 480 } },
            { id: 'channel-description', options: { maxLength: 500, warningThreshold: 450, errorThreshold: 480 } }
        ]);
    }
};

window.GuildButtons = GuildButtons;

// Self-initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    GuildButtons.init();
});