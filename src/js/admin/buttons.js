const AdminButtons = {
    setupPagination() {
        document.getElementById('prev-btn').addEventListener('click', () => {
            if (window.adminState.currentPage > 1) window.loadUsers(window.adminState.currentPage - 1);
        });

        document.getElementById('next-btn').addEventListener('click', () => {
            if (window.adminState.currentPage < window.adminState.totalPages) window.loadUsers(window.adminState.currentPage + 1);
        });
    },

    init() {
        this.setupPagination();
        this.setupUserActions();
    },
    setupUserActions() {
            document.getElementById('user-list-body').addEventListener('click', async function(e) {
                if (e.target.classList.contains('make-admin-btn')) {
                    const username = e.target.getAttribute('data-username');
                    showConfirmationDialog(`Are you sure you want to make ${username} an admin?`, async () => {
                        try {
                            await UserAPI.makeUserAdmin(username);
                            displayErrorMessage(`${username} is now an admin`, 'error-container', 'success');
                            window.loadUsers(window.adminState.currentPage);
                        } catch (error) {
                            displayErrorMessage('Failed to make user an admin', 'error-container', 'error');
                        }
                    });
                } else if (e.target.classList.contains('demote-admin-btn')) {
                    const username = e.target.getAttribute('data-username');
                    showConfirmationDialog(`Are you sure you want to revoke admin privileges from ${username}?`, async () => {
                        try {
                            await UserAPI.demoteUserAdmin(username);
                            displayErrorMessage(`${username} is no longer an admin`, 'error-container', 'success');
                            window.loadUsers(window.adminState.currentPage);
                        } catch (error) {
                            displayErrorMessage('Failed to revoke admin privileges', 'error-container', 'error');
                        }
                    });
                } else if (e.target.classList.contains('ban-btn')) {
                    const username = e.target.getAttribute('data-username');
                    const userID = e.target.getAttribute('data-userid');
                    showConfirmationDialog(`Are you sure you want to ban ${username}?`, async () => {
                        try {
                            await UserAPI.banUser(userID);
                            window.loadUsers(window.adminState.currentPage);
                        } catch (error) {
                            alert('Failed to ban user');
                        }
                    });
                } else if (e.target.classList.contains('unban-btn')) {
                    const username = e.target.getAttribute('data-username');
                    const userID = e.target.getAttribute('data-userid');
                    showConfirmationDialog(`Are you sure you want to unban ${username}?`, async () => {
                        try {
                            await UserAPI.unbanUser(userID);
                            window.loadUsers(window.adminState.currentPage);
                        } catch (error) {
                            alert('Failed to unban user');
                        }
                    });
                }
            });
        },    
};

window.AdminButtons = AdminButtons;
document.addEventListener('DOMContentLoaded', () => {
    AdminButtons.init();
});