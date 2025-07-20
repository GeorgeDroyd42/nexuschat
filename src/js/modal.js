document.addEventListener('DOMContentLoaded', () => {
    window.profileManager.loadUserData();
    
    const userProfileBtn = document.getElementById('user-profile-btn');
    
    if (userProfileBtn) {
        userProfileBtn.addEventListener('click', () => {
            window.modalManager.openModal('user-modal');
            window.profileManager.loadUserData();
        });
    }
    
    window.modalManager.setupModal('user-modal', null, 'close-modal');
});
