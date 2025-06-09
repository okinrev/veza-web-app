document.addEventListener('alpine:init', () => {
  Alpine.store('dmChat', {
    myUserId: null,
    otherUserId: parseInt(new URLSearchParams(window.location.search).get("user_id")),
    messages: [],
    contenu: '',
    socket: null,
    logEl: null,

    async init() {
      const token = localStorage.getItem('access_token');
      if (!token || isNaN(this.otherUserId)) return alert("Non connecté");

      const me = await fetch("/me", { headers: { Authorization: "Bearer " + token } });
      const user = await me.json();
      this.myUserId = user.id;

      // Close socket proprement
      if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
        this.socket.close();
        this.socket = null;
      }

      const socket = new WebSocket(`ws://localhost:9001/?token=${token}`);
      this.socket = socket;

      socket.onmessage = (event) => {
        const data = JSON.parse(event.data);

        if (data.type === "dm" && (data.data?.fromUser === this.otherUserId || data.data?.to === this.otherUserId)) {
          this.messages.push(data.data);
          this.scrollToBottom();
        } else if (Array.isArray(data)) {
          this.messages = data.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
          this.scrollToBottom();
        }
      };

      socket.onopen = () => {
        socket.send(JSON.stringify({
          type: "dm_history",
          with: this.otherUserId,
          limit: 50
        }));
      };
    },

    envoyerMessage() {
      const msg = this.contenu.trim();
      if (!msg || !this.socket || this.socket.readyState !== WebSocket.OPEN) return;

      this.socket.send(JSON.stringify({
        type: "dm",
        to: this.otherUserId,
        content: msg
      }));

      this.messages.push({
        fromUser: this.myUserId,
        to: this.otherUserId,
        content: msg,
        timestamp: new Date().toISOString(),
        username: 'Moi'
      });

      this.scrollToBottom();
      this.contenu = '';
    },

    scrollToBottom() {
      setTimeout(() => {
        const el = document.querySelector('[x-ref="log"]');
        if (el) el.scrollTop = el.scrollHeight;
      }, 0);
    }
  });

  Alpine.effect(() => {
    const el = document.querySelector('[x-ref="log"]');
    if (el) Alpine.store('dmChat').logEl = el;
  });
});
