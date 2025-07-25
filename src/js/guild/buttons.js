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
                window.profileManager.openProfile(true);
            });
        }
    
        if (closeProfileBtn) {
            closeProfileBtn.addEventListener('click', () => {
                window.profileManager.closeProfile();
            });
        }
    },

    init() {
        this.setupMembersToggle();
        this.setupGuildToggle();
        this.setupCopyInvite();
        this.setupCreateChannel();
        this.setupSettings();
    }
};

window.GuildButtons = GuildButtons;