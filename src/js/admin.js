let currentPage = 1;
let totalPages = 1;
window.confirmCallback = null;

async function loadUsers(page = 1) {
    try {
        const data = await UserAPI.getUsersList(page, 50);
        const currentUser = await window.profileManager.loadUserData();
        
        currentPage = page;
        totalPages = Math.ceil(data.total_count / 50);
        
        document.getElementById('total-users').textContent = data.total_count;
        document.getElementById('page-info').textContent = `Page ${currentPage} of ${totalPages}`;
        document.getElementById('prev-btn').disabled = currentPage === 1;
        document.getElementById('next-btn').disabled = currentPage >= totalPages;
        
        const userTableBody = document.getElementById('user-list-body');
        if (data.users.length === 0) {
            userTableBody.innerHTML = '<tr><td colspan="5">No users found</td></tr>';
            return;
        }
        
        userTableBody.innerHTML = '';
        data.users.forEach(user => {
            const isCurrentUser = currentUser && user.username === currentUser.username;
            const adminButton = user.is_admin ? 
                (isCurrentUser ? '' : `<button class="action-btn btn-secondary demote-admin-btn" data-username="${user.username}">Revoke Admin</button>`) : 
                `<button class="action-btn btn-warning make-admin-btn" data-username="${user.username}">Make Admin</button>`;
            
            const banButton = user.is_banned ?
                `<button class="action-btn btn-success unban-btn" data-userid="${user.user_id}" data-username="${user.username}">Unban</button>` :
                (isCurrentUser ? '' : `<button class="action-btn btn-danger ban-btn" data-userid="${user.user_id}" data-username="${user.username}">Ban</button>`);
                
        const userRow = document.createElement('tr');
            
            userRow.innerHTML = `
                <td>
                    <div class="user-info">
                        <div class="avatar-container"></div>
                        <span class="member-name">${user.username}</span>
                    </div>
                </td>
                <td>${formatTimestamp(user.created_at, 'datetime')}</td>
                <td>${user.is_admin ? 'Yes' : 'No'}</td>
                <td>${user.is_banned ? 'Banned' : 'Active'}</td>
                <td>${adminButton} ${banButton}</td>
            `;

            const avatarContainer = userRow.querySelector('.avatar-container');
            const avatarElement = window.createUserAvatarElement(user.username, user.profile_picture);
            avatarContainer.appendChild(avatarElement);
            userTableBody.appendChild(userRow);
        });
        
    } catch (error) {
        console.error('Error fetching users:', error);
        document.getElementById('user-list-body').innerHTML = 
            '<tr><td colspan="5">Error loading user data</td></tr>';
    }
}
function setupSearch() {
    const searchInput = document.getElementById('user-search');
    const clearBtn = document.getElementById('clear-search');
    
    if (!searchInput) return;
    
    searchInput.addEventListener('input', (e) => {
        const term = e.target.value.toLowerCase().trim();
        const rows = document.querySelectorAll('#user-list-body tr');
        
        rows.forEach(row => {
            const username = row.querySelector('.member-name')?.textContent?.toLowerCase() || '';
            row.style.display = !term || username.startsWith(term) ? '' : 'none';
        });
    });
    
    if (clearBtn) {
        clearBtn.addEventListener('click', () => {
            searchInput.value = '';
            document.querySelectorAll('#user-list-body tr').forEach(row => {
                row.style.display = '';
            });
        });
    }
}
document.addEventListener('DOMContentLoaded', async function() {    
    const userData = await window.profileManager.loadUserData();
    if (userData) {
        document.getElementById('admin-username').textContent = userData.username;
    }
        
    document.getElementById('prev-btn').addEventListener('click', () => {
        if (currentPage > 1) loadUsers(currentPage - 1);
    });
            document.getElementById('user-list-body').addEventListener('click', async function(e) {
                if (e.target.classList.contains('make-admin-btn')) {
                    const username = e.target.getAttribute('data-username');
                    showConfirmationDialog(`Are you sure you want to make ${username} an admin?`, async () => {
                        try {
                            await UserAPI.makeUserAdmin(username);
                            displayErrorMessage(`${username} is now an admin`, 'error-container', 'success');
                            loadUsers(currentPage);
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
                            loadUsers(currentPage);
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
                            loadUsers(currentPage);
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
                            loadUsers(currentPage);
                        } catch (error) {
                            alert('Failed to unban user');
                        }
                    });
                }

            });    
    
    document.getElementById('next-btn').addEventListener('click', () => {
        if (currentPage < totalPages) loadUsers(currentPage + 1);
    });
    
    window.modalManager.setupModal('confirm-modal', null, 'close-confirm-modal');
    window.modalManager.setupModal('confirm-modal', null, 'confirm-no');
    
    document.getElementById('confirm-yes').addEventListener('click', () => {
        if (window.confirmCallback) {
            window.confirmCallback();
            window.confirmCallback = null;
        }
        window.modalManager.closeModal('confirm-modal');
    });
    setupSearch();
    loadUsers(1);
});

document.getElementById('logout-btn').addEventListener('click', async () => {
    try {
        await AuthAPI.logout();
        window.location.href = '/login';
    } catch (error) {
        console.error('Logout failed:', error);
    }
});