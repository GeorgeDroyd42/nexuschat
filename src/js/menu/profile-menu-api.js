class ProfileMenuAPI extends ChannelMenuAPI {
    constructor() {
        super();
        this.currentUser = null;
    }

renderButtons(modalId, userData) {
    this.currentUser = userData;
    
    const modal = document.getElementById(modalId);
    if (!modal) return;
    
    const content = modal.querySelector('.profile-menu-content');
    if (!content) return;
    
    this.createSidebarLayout(content);
    this.setupTabSwitching(content);
    this.populateTabContent();
    this.disableContextMenu(content);
}

disableContextMenu(content) {
    content.addEventListener('contextmenu', (e) => {
        e.preventDefault();
        return false;
    });
}

    renderItem(panel, item) {
        if (item.type === 'profile-info') {
            this.renderProfileInfo(panel);
        } else {
            super.renderItem(panel, item);
        }
    }

renderProfileInfo(panel) {
    const profileContainer = document.createElement('div');
    profileContainer.className = 'profile-picture-container';
    
    const avatarWrapper = AvatarUtils.createSecureAvatar(
        this.currentUser?.username || 'User',
        this.currentUser?.profile_picture,
        'profile-picture-large'
    );
    
    profileContainer.appendChild(avatarWrapper);
    panel.appendChild(profileContainer);

const fields = [
    { label: 'USERNAME', id: 'profile-info-username', value: this.currentUser?.username || 'Loading...' },
    { label: 'MEMBER SINCE', id: 'profile-info-created', value: this.currentUser?.created_at ? new Date(this.currentUser.created_at).toLocaleDateString() : 'Loading...' },
    { label: 'USER ID', id: 'profile-info-id', value: this.currentUser?.user_id || 'Loading...', class: 'user-id' }
];

this.renderInfoSection(panel, fields, 'profile-info-section');

const bioFields = [
    { label: 'BIO', id: 'profile-info-bio', value: this.currentUser?.bio || 'No bio set', class: 'bio-display' }
];

this.renderInfoSection(panel, bioFields, 'bio-section');
}
}

window.profileMenuAPI = new ProfileMenuAPI();

profileMenuAPI
.addTab('profile', 'Profile')
.addTab('settings', 'Settings')

.addToTab('profile', { type: 'profile-info' })

.addToTab('settings', { type: 'textarea', label: 'USERNAME', id: 'username-edit', value: '', placeholder: 'Enter new username...', rows: 1 })
.addToTab('settings', { type: 'button', text: 'Save Username', action: async () => {
    const newUsername = document.getElementById('username-edit').value.trim();
    if (!newUsername) return;
    
    try {
        const result = await UserAPI.updateUsername(newUsername);
        if (result.status === 'success') {
            const freshUserData = await window.profileManager.loadUserData();
            if (freshUserData) {
                window.profileMenuAPI.currentUser = freshUserData;
                window.profileMenuAPI.populateTabContent();
            }
            
            document.getElementById('username-edit').value = '';
            
            if (window.loadMessages && typeof window.loadMessages === 'function') {
                window.loadMessages();
            }
            
            console.log('Username updated successfully');
        }
    } catch (error) {
        console.error('Error updating username:', error);
    }
}, color: '#4CAF50' })
.addToTab('settings', { type: 'separator' })
.addToTab('settings', { type: 'textarea', label: 'BIO', id: 'bio-edit', value: '', placeholder: 'Tell us about yourself... (max 2000 characters)', rows: 4, spellcheck: false })
.addToTab('settings', { type: 'button', text: 'Save Bio', action: async () => {
    const newBio = document.getElementById('bio-edit').value.trim();
    
    try {
        const result = await UserAPI.updateBio(newBio);
        if (result.status === 'success') {
            const freshUserData = await window.profileManager.loadUserData();
            if (freshUserData) {
                window.profileMenuAPI.currentUser = freshUserData;
                window.profileMenuAPI.populateTabContent();
            }
            
            console.log('Bio updated successfully');
        }
    } catch (error) {
        console.error('Error updating bio:', error);
    }
}, color: '#2196F3' })
