document.addEventListener('DOMContentLoaded', () => {
  registerCommonEventListeners();
  startSessionRefresh();
  
  if (isMainPage()) {
    const membersSidebar = document.getElementById('members-sidebar');
    
    if (membersSidebar) membersSidebar.classList.remove('visible');
  }
});

window.GuildOwnership = {
    async checkOwnership(guildId) {
        return await window.PermissionManager.hasPermission(guildId, 'canManageChannels');
    },
    
    async showOwnerElements(guildId) {
        await window.PermissionManager.updateGuildUI(guildId);
        return await this.checkOwnership(guildId);
    },
    
    clearCache(guildId) {
        window.PermissionManager.clearCache(guildId);
    }
};