document.addEventListener('DOMContentLoaded', () => {
    const joinBtn = document.getElementById('join-guild-btn');
    if (joinBtn) {
        joinBtn.addEventListener('click', async () => {
            const inviteCode = joinBtn.dataset.inviteCode;
            
            // Update button state
            joinBtn.textContent = 'Joining...';
            joinBtn.disabled = true;
            
            try {
                const result = await API.invite.joinByInvite(inviteCode);

                
                if (result.error) {
                    // Show error and restore button
                    displayErrorMessage(result.error, 'guild-error-container');
                    joinBtn.textContent = 'Join Guild';
                    joinBtn.disabled = false;
                } else {
                    // Success - navigate to guild using existing functions
                    if (result.redirect_url) {
    window.location.href = result.redirect_url;
} else {
    window.location.href = '/v/main';
}
                }
            } catch (error) {
                console.error('Error joining guild:', error);
                displayErrorMessage('Failed to join guild', 'guild-error-container');
                joinBtn.textContent = 'Join Guild';
                joinBtn.disabled = false;
            }
        });
    }
});