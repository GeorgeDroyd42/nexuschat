class PermissionManager {
    constructor() {
        this.cache = new Map();
    }

    async getGuildPermissions(guildId) {
        if (this.cache.has(guildId)) {
            return this.cache.get(guildId);
        }

        try {
            const response = await fetch(`/api/guild/${guildId}/permissions`, {
                headers: { 'X-CSRF-Token': window.csrfToken }
            });
            
            if (response.ok) {
                const data = await response.json();
                const permissions = {
                    isOwner: data.is_owner || false,
                    ...data.permissions
                };
                
                this.cache.set(guildId, permissions);
                return permissions;
            }
        } catch (error) {
            console.error('Error fetching guild permissions:', error);
        }
        
        return { isOwner: false, canManageChannels: false, canDeleteMessages: false, canCreateInvite: false, canManageGuild: false };
    }

    async hasPermission(guildId, permission) {
        const permissions = await this.getGuildPermissions(guildId);
        return permissions[permission] || false;
    }

async updateGuildUI(guildId) {
    const permissions = await this.getGuildPermissions(guildId);
    
    if (permissions.canEditChannels) {
        document.querySelectorAll('.channel-settings-btn').forEach(btn => btn.style.display = 'block');
    } else {
        document.querySelectorAll('.channel-settings-btn').forEach(btn => btn.style.display = 'none');
    }
    
    if (permissions.canCreateChannels) {
        const createBtn = document.getElementById('create-channel-btn');
        if (createBtn) createBtn.style.display = 'block';
        const createMobileBtn = document.querySelector('.add-channel-mobile');
        if (createMobileBtn) createMobileBtn.style.display = 'block';
    } else {
        const createBtn = document.getElementById('create-channel-btn');
        if (createBtn) createBtn.style.display = 'none';
        const createMobileBtn = document.querySelector('.add-channel-mobile');
        if (createMobileBtn) createMobileBtn.style.display = 'none';
    }
}

    clearCache(guildId) {
        if (guildId) {
            this.cache.delete(guildId);
        } else {
            this.cache.clear();
        }
    }
}

window.PermissionManager = new PermissionManager();