// WebSocket Service pour le chat - Compatible avec l'ancien chat.html
// Se connecte au serveur Rust chat_server sur le port 9001

import { useAuthStore } from '@/shared/stores/authStore';

// Types pour les messages entrants (du serveur Rust vers le client)
type RustInboundMessageType =
  | 'message'      // Message de salon 
  | 'dm'           // Message direct
  | 'dm_history'   // Historique DM
  | 'join_ack'     // Confirmation de join
  | 'message_sent' // Confirmation message envoyé
  | 'dm_sent'      // Confirmation DM envoyé
  | 'error';       // Erreur

// Types pour les messages sortants (du client vers le serveur Rust)
type RustOutboundMessageType = 
  | 'join'         // Rejoindre un salon
  | 'message'      // Envoyer message salon
  | 'dm'           // Envoyer message direct
  | 'room_history' // Demander historique salon
  | 'dm_history';  // Demander historique DM

// Structure des messages sortants vers le serveur Rust
interface RustOutboundMessage {
  type: RustOutboundMessageType;
  room?: string;      // Pour join, message, room_history
  content?: string;   // Pour message, dm
  to?: number;        // Pour dm (user_id destinataire)
  with?: number;      // Pour dm_history (user_id correspondant)
  limit?: number;     // Pour *_history (nombre max de messages)
}

// Structure des messages entrants du serveur Rust
interface RustInboundMessage {
  type: RustInboundMessageType;
  data?: any;
  message?: string;   // Pour les erreurs
}

// Message individuel tel que reçu du serveur Rust
export interface RustChatMessage {
  id?: number;
  fromUser?: number;
  username: string;
  content: string;
  timestamp: string;
  room?: string;      // Pour les messages de salon
  to?: number;        // Pour les messages directs
}

// Events émis par le WebSocket manager
type ChatEventType =
  | 'connected'
  | 'disconnected'
  | 'message_received'    // Nouveau message de salon
  | 'dm_received'         // Nouveau message direct reçu
  | 'room_history'        // Historique salon reçu
  | 'dm_history'          // Historique DM reçu
  | 'join_confirmed'      // Confirmation de join salon
  | 'message_sent'        // Confirmation message salon envoyé
  | 'dm_sent'             // Confirmation DM envoyé
  | 'error';              // Erreur serveur

export class RustChatWebSocketManager {
  private ws: WebSocket | null = null;
  private eventListeners: { [K in ChatEventType]: Function[] } = {
    connected: [],
    disconnected: [],
    message_received: [],
    dm_received: [],
    room_history: [],
    dm_history: [],
    join_confirmed: [],
    message_sent: [],
    dm_sent: [],
    error: []
  };
  
  private currentRoom: string | null = null;
  private historyMessages: RustChatMessage[] = [];
  private historyTimeout: NodeJS.Timeout | null = null;
  
  // === MÉTHODES D'ÉCOUTE ===
  
  on<T extends ChatEventType>(event: T, callback: Function) {
    this.eventListeners[event].push(callback);
  }
  
  off<T extends ChatEventType>(event: T, callback: Function) {
    const listeners = this.eventListeners[event];
    const index = listeners.indexOf(callback);
    if (index > -1) {
      listeners.splice(index, 1);
    }
  }
  
  private emit<T extends ChatEventType>(event: T, data?: any) {
    console.log(`[Rust Chat WebSocket] Événement émis: ${event}`, data);
    this.eventListeners[event].forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error(`[Rust Chat WebSocket] Erreur dans le callback ${event}:`, error);
      }
    });
  }
  
  // === CONNEXION/DÉCONNEXION ===
  
  connect() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('[Rust Chat WebSocket] Déjà connecté');
      return;
    }
    
    // Essayer plusieurs sources de token (compatibilité avec l'ancien chat)
    const token = localStorage.getItem('access_token') || 
                 localStorage.getItem('authToken') || 
                 useAuthStore.getState().token;
    
    if (!token) {
      console.error('[Rust Chat WebSocket] Pas de token disponible');
      console.log('[Rust Chat WebSocket] Vérification localStorage:', {
        access_token: !!localStorage.getItem('access_token'),
        authToken: !!localStorage.getItem('authToken'),
        storeToken: !!useAuthStore.getState().token
      });
      this.emit('error', { error: 'Token manquant' });
      return;
    }
    
    console.log('[Rust Chat WebSocket] Token trouvé dans:', 
      localStorage.getItem('access_token') ? 'access_token' :
      localStorage.getItem('authToken') ? 'authToken' : 'store'
    );
    
    try {
      // URL identique à l'ancien chat.html
      const wsUrl = `ws://localhost:9001/?token=${token}`;
      console.log('[Rust Chat WebSocket] Connexion à:', wsUrl);
      
      this.ws = new WebSocket(wsUrl);
      
      this.ws.onopen = () => {
        console.log('[Rust Chat WebSocket] Connexion établie');
        this.emit('connected');
      };
      
      this.ws.onclose = (event) => {
        console.log('[Rust Chat WebSocket] Connexion fermée:', event.code, event.reason);
        this.emit('disconnected');
        
        // Reconnexion automatique si pas une fermeture normale
        if (event.code !== 1000) {
          setTimeout(() => {
            console.log('[Rust Chat WebSocket] Tentative de reconnexion...');
            this.connect();
          }, 5000);
        }
      };
      
      this.ws.onerror = (error) => {
        console.error('[Rust Chat WebSocket] Erreur:', error);
        this.emit('error', { error: 'Erreur de connexion WebSocket' });
      };
      
      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.handleRustMessage(data);
        } catch (error) {
          console.error('[Rust Chat WebSocket] Erreur parsing message:', error, event.data);
        }
      };
      
    } catch (error) {
      console.error('[Rust Chat WebSocket] Erreur de connexion:', error);
      this.emit('error', { error: 'Impossible de se connecter' });
    }
  }
  
  disconnect() {
    if (this.ws) {
      this.ws.onclose = null; // Empêcher la reconnexion
      this.ws.close(1000, 'Déconnexion demandée');
      this.ws = null;
    }
    this.currentRoom = null;
    this.historyMessages = [];
    if (this.historyTimeout) {
      clearTimeout(this.historyTimeout);
      this.historyTimeout = null;
    }
  }
  
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
  
  // === GESTION DES MESSAGES ENTRANTS ===
  
  private handleRustMessage(message: any) {
    console.log('[Rust Chat WebSocket] Message reçu:', message);
    
    // Messages avec type explicite (confirmations, erreurs...)
    if (message.type) {
      switch (message.type) {
        case 'join_ack':
          console.log('[Rust Chat WebSocket] Confirmation de join salon:', message.data);
          this.emit('join_confirmed', message.data);
          break;
          
        case 'message_sent':
          console.log('[Rust Chat WebSocket] Confirmation message salon envoyé:', message.data);
          this.emit('message_sent', message.data);
          break;
          
        case 'dm_sent':
          console.log('[Rust Chat WebSocket] Confirmation DM envoyé:', message.data);
          this.emit('dm_sent', message.data);
          break;
          
        case 'dm_history':
          console.log('[Rust Chat WebSocket] Historique DM reçu:', message.data);
          const dmHistory = Array.isArray(message.data) ? message.data : [];
          this.emit('dm_history', dmHistory);
          break;
          
        case 'message':
          console.log('[Rust Chat WebSocket] Message salon reçu:', message.data);
          if (message.data) {
            this.emit('message_received', message.data);
          }
          break;
          
        case 'dm':
          console.log('[Rust Chat WebSocket] Message privé reçu:', message.data);
          if (message.data) {
            this.emit('dm_received', message.data);
          }
          break;
          
        case 'error':
          console.error('[Rust Chat WebSocket] Erreur serveur:', message.data);
          this.emit('error', message.data);
          break;
          
        default:
          console.warn('[Rust Chat WebSocket] Type de message non géré:', message.type);
      }
    }
    // Messages d'historique de salon (envoyés individuellement)
    else if (message.id && message.content && message.username && message.timestamp) {
      console.log('[Rust Chat WebSocket] Message d\'historique salon détecté:', message);
      
      // Collecter les messages d'historique avec un délai pour les grouper
      this.historyMessages.push({
        id: message.id,
        content: message.content,
        username: message.username,
        timestamp: message.timestamp,
        fromUser: message.fromUser || 0
      });

      // Reset le timeout pour grouper les messages
      if (this.historyTimeout) {
        clearTimeout(this.historyTimeout);
      }
      
      // Émettre l'historique groupé après 100ms de silence
      this.historyTimeout = setTimeout(() => {
        console.log('[Rust Chat WebSocket] Envoi historique groupé:', this.historyMessages.length, 'messages');
        this.emit('room_history', [...this.historyMessages]);
        this.historyMessages = [];
        this.historyTimeout = null;
      }, 100);
    }
    else {
      console.warn('[Rust Chat WebSocket] Format de message non reconnu:', message);
    }
  }

  // === MÉTHODES POUR ENVOYER DES MESSAGES AU SERVEUR RUST ===
  
  private sendToRust(payload: any) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[Rust Chat WebSocket] WebSocket non connecté');
      this.emit('error', { error: 'WebSocket non connecté' });
      return;
    }
    
    const message = JSON.stringify(payload);
    console.log('[Rust Chat WebSocket] Envoi vers serveur:', message);
    this.ws.send(message);
  }
  
  /**
   * Rejoindre un salon
   */
  joinRoom(roomName: string) {
    this.currentRoom = roomName;
    this.sendToRust({
      type: 'join',
      room: roomName
    });
  }

  /**
   * Envoyer un message dans un salon
   */
  sendRoomMessage(roomName: string, content: string) {
    this.sendToRust({
      type: 'message',
      room: roomName,
      content
    });
  }

  /**
   * Envoyer un message privé
   */
  sendDirectMessage(toUserId: number, content: string) {
    this.sendToRust({
      type: 'dm',
      to: toUserId,
      content
    });
  }

  /**
   * Demander l'historique d'un salon
   */
  getRoomHistory(roomName: string, limit = 50) {
    this.sendToRust({
      type: 'room_history',
      room: roomName,
      limit
    });
  }

  /**
   * Demander l'historique des messages privés avec un utilisateur
   */
  getDMHistory(withUserId: number, limit = 50) {
    this.sendToRust({
      type: 'dm_history',
      with: withUserId,
      limit
    });
  }
}

// Instance singleton
export const rustChatWebSocket = new RustChatWebSocketManager(); 