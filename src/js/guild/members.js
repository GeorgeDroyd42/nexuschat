const GuildMembers = {
    currentGuildId: null,
    isLoadingMembers: false,
    hasMoreMembers: true,
    
    async loadGuildMembers(guildID) {
        if (this.isLoadingMembers) return;
        
        this.currentGuildId = guildID;
        this.isLoadingMembers = true;
        this.hasMoreMembers = true;
        
        try {
            const data = await GuildAPI.getMembers(guildID);
            
            if (data.success) {
                this.hasMoreMembers = data.has_more;
                updateMembersList(data.members, guildID);
            } else if (data.error) {
                console.error('Error loading members:', data.error);
            }
        } catch (error) {
            console.error('Error loading guild members:', error);
        } finally {
            this.isLoadingMembers = false;
        }
    },

    async getUsernameByID(userID) {
        try {
            const data = await UserAPI.getUserProfile(userID);
            return data.username || 'Unknown';
        } catch {
            return 'Unknown';
        }
    },

    async getUserProfilePicture(userID) {
        try {
            const data = await UserAPI.getUserProfile(userID);
            return data.profile_picture || '';
        } catch {
            return '';
        }
    },

    setupMembersSidebar(guildId) {
        if (window.innerWidth > 768) {
            document.getElementById('members-sidebar').classList.add('visible');
            document.querySelector('.main-content').classList.add('with-members');
        }
        this.loadGuildMembers(guildId);
    }
};

window.GuildMembers = GuildMembers;