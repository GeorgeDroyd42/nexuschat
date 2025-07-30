window.adminState = {
    currentPage: 1,
    totalPages: 1
};
window.confirmCallback = null;

async function loadUsers(page = 1) {
    try {
        const data = await UserAPI.getUsersList(page, 50);
        const currentUser = await window.profileManager.loadUserData();
        
        window.adminState.currentPage = page;
        window.adminState.totalPages = Math.ceil(data.total_count / 50);
        
        document.getElementById('total-users').textContent = data.total_count;
        document.getElementById('page-info').textContent = `Page ${window.adminState.currentPage} of ${window.adminState.totalPages}`;
        document.getElementById('prev-btn').disabled = window.adminState.currentPage === 1;
        document.getElementById('next-btn').disabled = window.adminState.currentPage >= window.adminState.totalPages;
        
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
                        <span class="member-name"></span>
                    </div>
                </td>
                <td>${formatTimestamp(user.created_at, 'datetime')}</td>
                <td>${user.is_admin ? 'Yes' : 'No'}</td>
                <td>${user.is_banned ? 'Banned' : 'Active'}</td>
                <td>${adminButton} ${banButton}</td>
            `;
            userRow.querySelector('.member-name').textContent = user.username;

            const avatarContainer = userRow.querySelector('.avatar-container');
            const avatarElement = window.AvatarUtils.createSecureAvatar(user.username, user.profile_picture);
            avatarContainer.appendChild(avatarElement);
            userTableBody.appendChild(userRow);
        });
        
    } catch (error) {
        console.error('Error fetching users:', error);
        document.getElementById('user-list-body').innerHTML = 
            '<tr><td colspan="5">Error loading user data</td></tr>';
    }
}
window.loadUsers = loadUsers;
document.addEventListener('DOMContentLoaded', async function() {    
    const userData = await window.profileManager.loadUserData();
    if (userData) {
        document.getElementById('admin-username').textContent = userData.username;
    }
    
    window.modalManager.setupModal('confirm-modal', null, 'close-confirm-modal');
    window.modalManager.setupModal('confirm-modal', null, 'confirm-no');
    
    document.getElementById('confirm-yes').addEventListener('click', () => {
        if (window.confirmCallback) {
            window.confirmCallback();
            window.confirmCallback = null;
        }
        window.modalManager.closeModal('confirm-modal');
    });
    AdminSearch.setupSearch();
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