// Dashboard State
document.addEventListener('alpine:init', () => {
  Alpine.data('dashboard', () => ({
    stats: {
      totalUsers: 0,
      activeUsers: 0,
      totalListings: 0,
      activeListings: 0
    },
    recentActivity: [],
    notifications: [],
    isLoading: true,
    error: null,

    init() {
      this.fetchDashboardStats();
      this.fetchRecentActivity();
      
      // Refresh data every 5 minutes
      setInterval(() => {
        this.fetchDashboardStats();
        this.fetchRecentActivity();
      }, 5 * 60 * 1000);
    },

    async fetchDashboardStats() {
      try {
        this.isLoading = true;
        const response = await fetch('/api/v1/admin/dashboard', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('access_token')}`
          }
        });
        
        if (!response.ok) {
          throw new Error('Failed to fetch dashboard stats');
        }

        const data = await response.json();
        this.stats = data.stats;
        this.error = null;
      } catch (error) {
        console.error('Error fetching dashboard stats:', error);
        this.error = 'Failed to load dashboard statistics';
      } finally {
        this.isLoading = false;
      }
    },

    async fetchRecentActivity() {
      try {
        const response = await fetch('/api/v1/admin/activity', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('access_token')}`
          }
        });

        if (!response.ok) {
          throw new Error('Failed to fetch recent activity');
        }

        const data = await response.json();
        this.recentActivity = data.activities;
        this.error = null;
      } catch (error) {
        console.error('Error fetching recent activity:', error);
        this.error = 'Failed to load recent activity';
      }
    },

    getActivityIcon(type) {
      const icons = {
        user: '👤',
        listing: '📝',
        message: '💬',
        system: '⚙️'
      };
      return icons[type] || '📌';
    },

    formatTime(timestamp) {
      const date = new Date(timestamp);
      return date.toLocaleString();
    },

    showError(message) {
      this.error = message;
      setTimeout(() => {
        this.error = null;
      }, 5000);
    }
  }));
}); 