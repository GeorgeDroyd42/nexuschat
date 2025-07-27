const GuildMembers = {
    currentGuildId: null,
    isLoadingMembers: false,
    hasMoreMembers: true,
    currentPage: 1,
    allMembers: [],
    
    async loadGuildMembers(guildID, page = 1) {
        if (this.isLoadingMembers) return;
        
        this.currentGuildId = guildID;
        this.isLoadingMembers = true;
        
        try {
            const data = await GuildAPI.getMembers(guildID, page);
            
            if (data.success) {
                console.log(`📄 Page ${page} received:`, {
                    totalMembers: data.members.length,
                    hasMore: data.has_more,
                    onlineCount: data.members.filter(m => m.is_online).length,
                    offlineCount: data.members.filter(m => !m.is_online).length
                });
                
                console.log(`👥 Members on page ${page}:`, data.members.map(m => `${m.username}(${m.is_online ? 'online' : 'offline'})`));
                
                if (page === 1) {
                    this.allMembers = data.members;
                    console.log(`🔄 Reset allMembers to page 1 data`);
                } else {
                    this.allMembers = [...this.allMembers, ...data.members];
                    console.log(`📎 Added page ${page} to existing members, total now: ${this.allMembers.length}`);
                }
                this.hasMoreMembers = data.has_more;
                this.currentPage = page;
                
                console.log(`🎯 Final allMembers order:`, this.allMembers.map(m => `${m.username}(${m.is_online ? 'online' : 'offline'})`));
                updateMembersList(this.allMembers, guildID);
            } else if (data.error) {
                console.error('Error loading members:', data.error);
            }
        } catch (error) {
            console.error('Error loading guild members:', error);
        } finally {
            this.isLoadingMembers = false;
        }
    },

    async loadMoreMembers() {
        if (!this.hasMoreMembers || this.isLoadingMembers) return;
        await this.loadGuildMembers(this.currentGuildId, this.currentPage + 1);
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
            
            const sidebar = document.getElementById('members-sidebar');
            sidebar.removeEventListener('scroll', this.scrollHandler);
            this.scrollHandler = () => {
                if (sidebar.scrollTop + sidebar.clientHeight >= sidebar.scrollHeight - 50) {
                    this.loadMoreMembers();
                }
            };
            sidebar.addEventListener('scroll', this.scrollHandler);
        }
        this.loadGuildMembers(guildId);
    }
};

window.GuildMembers = GuildMembers;