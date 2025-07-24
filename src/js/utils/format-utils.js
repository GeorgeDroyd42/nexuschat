function formatTimestamp(timestamp, format = 'date') {
    if (!timestamp) return 'Unknown';
    const date = new Date(timestamp);
    if (isNaN(date.getTime())) return timestamp;
    
    switch (format) {
        case 'time':
            return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        case 'datetime':
            return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        case 'relative':
            return getRelativeTime(date);
        case 'date':
        default:
            return date.toLocaleDateString();
    }
}

function getRelativeTime(date) {
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);
    
    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
}

window.formatTimestamp = formatTimestamp;
window.getRelativeTime = getRelativeTime;

function isMainPage() {
    return window.location.pathname === '/v/main';
}

