class ProfileManager {
    constructor() {
        this.currentUser = null;
        this.isLoading = false;
    }

    async loadUserData() {
        if (this.isLoading) return this.currentUser;
        
        this.isLoading = true;
        try {
            this.currentUser = await fetchCurrentUser();
            this.updateAllElements(this.currentUser);
            return this.currentUser;
        } catch (error) {
            console.error('ProfileManager: Failed to load user data:', error);
            return null;
        } finally {
            this.isLoading = false;
        }
    }

updateAllElements(userData) {
    if (!userData) return;
    
    const textElements = {
        '#username-display': userData.username,
        '#modal-username': userData.username,
        '#admin-username': userData.username,
        '#user-display-name': userData.username,
        '#modal-bio': userData.bio || 'No bio set'
    };

    Object.entries(textElements).forEach(([selector, text]) => {
        const el = document.querySelector(selector);
        if (el) el.textContent = text;
    });

    if (userData.created_at) {
        const formattedDate = formatTimestamp(userData.created_at, 'date');
        const dateElements = ['#user-created'];
        dateElements.forEach(selector => {
            const el = document.querySelector(selector);
            if (el) el.textContent = formattedDate;
        });
    }

const imageElements = [
    { selector: '#profile-preview', className: 'avatar-circle-lg' },
    { selector: '#user-avatar', className: 'avatar-circle-sm' }
];
imageElements.forEach(({ selector, className }) => {
    const el = document.querySelector(selector);
    if (el && window.AvatarUtils) {
        const newAvatar = window.AvatarUtils.createSecureAvatar(userData.username, userData.profile_picture, className);
        el.replaceWith(newAvatar);
    }
});
}
}

window.profileManager = new ProfileManager();