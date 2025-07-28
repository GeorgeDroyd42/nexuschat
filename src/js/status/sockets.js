const StatusSockets = {
    handleUserStatusChanged(data) {
        if (isCurrentGuild(data.guild_id)) {
            window.StatusCore.updateMemberStatus(data.user_id, data.is_online, data.guild_id);
            window.StatusUI.refreshGuildSettingsMembers();
        }
    }
};

window.StatusSockets = StatusSockets;