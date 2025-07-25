const AuthUI = {
    init() {
        this.setupImageUpload();
        this.handleSessionMessages();
    },

    setupImageUpload() {
        setupImageUpload('profile_picture', 'profile-preview', 'select-profile-btn');
    },

    handleSessionMessages() {
        const sessionMessage = sessionStorage.getItem('sessionMessage');
        const banMessage = sessionStorage.getItem('banMessage');
        
        if (sessionMessage) {
            displayErrorMessage(sessionMessage);
            sessionStorage.removeItem('sessionMessage');
        } else if (banMessage) {
            displayErrorMessage(banMessage);
            sessionStorage.removeItem('banMessage');
        }
    }
};

window.AuthUI = AuthUI;

document.addEventListener('DOMContentLoaded', () => {
    AuthUI.init();
});