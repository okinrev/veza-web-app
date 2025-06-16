import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { 
  ArrowLeft, 
  Hash, 
  MessageSquare, 
  Send, 
  Plus, 
  Users, 
  Settings, 
  Circle
} from 'lucide-react';
import UserStatusIndicator from '../components/UserStatusIndicator';
import UnreadBadge from '../components/UnreadBadge';

// Types simplifiés
interface Room {
  id: string;
  name: string;
  description?: string;
  memberCount?: number;
}

interface Message {
  id: string;
  content: string;
  senderId: string;
  sender: {
    id: string;
    username: string;
  };
  createdAt: string;
  timestamp?: string;
  type?: 'text' | 'system';
  room?: string;
}

interface Conversation {
  userId: string;
  username: string;
  displayName?: string;
  isOnline?: boolean;
  lastSeen?: string;
  lastMessage?: string | { content: string };
  lastMessageTime?: string;
  unreadCount: number;
}

// Fonction utilitaire pour extraire les valeurs sql.NullString de Go
const extractNullString = (value: any): string => {
  if (value && typeof value === 'object' && 'String' in value && 'Valid' in value) {
    return value.Valid ? value.String : '';
  }
  return value || '';
};

export function ChatPage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const { toast } = useToast();
  const messagesEndRef = useRef<HTMLDivElement>(null);
  
  // États principaux
  const [isLoading, setIsLoading] = useState(true);
  const [isConnected, setIsConnected] = useState(false);
  const [activeTab, setActiveTab] = useState<'rooms' | 'dm'>('rooms');
  
  // États pour les salons
  const [rooms, setRooms] = useState<Room[]>([]);
  const [currentRoom, setCurrentRoom] = useState<Room | null>(null);
  
  // États pour les messages directs
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [currentConversation, setCurrentConversation] = useState<Conversation | null>(null);
  
  // États pour les messages
  const [messages, setMessages] = useState<Message[]>([]);
  const [messageInput, setMessageInput] = useState('');
  const [loadingMessages, setLoadingMessages] = useState(false);
  const [typingUsers, setTypingUsers] = useState<string[]>([]);
  
  // États pour la création de salon
  const [showNewRoomForm, setShowNewRoomForm] = useState(false);
  const [newRoomName, setNewRoomName] = useState('');

  // États pour WebSocket
  const [socket, setSocket] = useState<WebSocket | null>(null);

  // Fonction pour s'assurer d'avoir un token valide avec refresh automatique
  const ensureValidToken = async (): Promise<string | null> => {
    const token = localStorage.getItem('access_token');
    if (!token) return null;
    
    try {
      // Vérifier si le token est expiré
      const payload = JSON.parse(atob(token.split('.')[1]));
      const now = Math.floor(Date.now() / 1000);
      
      // Vérifier si le token expire dans les 5 prochaines minutes
      if (payload.exp && payload.exp < (now + 300)) {
        console.log('Token expire bientôt, tentative de rafraîchissement...');
        
        // Essayer de rafraîchir le token
        const refreshToken = localStorage.getItem('refresh_token');
        if (refreshToken) {
          try {
            const response = await fetch('/api/v1/auth/refresh', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${refreshToken}`
              }
            });
            
            if (response.ok) {
              const data = await response.json();
              if (data.success && data.data.access_token) {
                localStorage.setItem('access_token', data.data.access_token);
                if (data.data.refresh_token) {
                  localStorage.setItem('refresh_token', data.data.refresh_token);
                }
                console.log('Token rafraîchi avec succès');
                return data.data.access_token;
              }
            } else {
              console.log('Échec du rafraîchissement, token peut-être encore valide');
              // Si le refresh échoue mais que le token n'est pas encore expiré, on continue
              if (payload.exp > now) {
                return token;
              }
            }
          } catch (error) {
            console.error('Erreur lors du rafraîchissement du token:', error);
            // Si le refresh échoue mais que le token n'est pas encore expiré, on continue
            if (payload.exp > now) {
              return token;
            }
          }
        }
        
        // Si le rafraîchissement échoue et que le token est expiré
        if (payload.exp <= now) {
          localStorage.removeItem('access_token');
          localStorage.removeItem('refresh_token');
          localStorage.removeItem('user');
          return null;
        }
      }
      
      return token;
    } catch (error) {
      console.error('Erreur lors de la vérification du token:', error);
      return null;
    }
  };

  // Fonction pour faire défiler vers le bas
  const scrollToBottom = () => {
    setTimeout(() => {
      if (messagesEndRef.current) {
        messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
      }
    }, 100);
  };

  // Connecter au WebSocket Rust
  const connectWebSocket = async () => {
    try {
      const token = await ensureValidToken();
      if (!token) return;

      // Fermer la connexion existante
      if (socket) {
        socket.onclose = null;
        socket.onerror = null;
        socket.close();
        setSocket(null);
      }

      const ws = new WebSocket(`ws://localhost:9001/?token=${token}`);
      
      ws.onopen = () => {
        setIsConnected(true);
        console.log('WebSocket connecté');
        toast({
          title: "Connexion établie",
          description: "Chat en temps réel activé"
        });
        
        // Si on est en mode DM avec un utilisateur sélectionné
        if (activeTab === 'dm' && currentConversation) {
          setLoadingMessages(true);
          ws.send(JSON.stringify({
            type: "dm_history",
            with: parseInt(currentConversation.userId),
            limit: 50
          }));
          
          setTimeout(() => {
            setLoadingMessages(false);
          }, 10000);
        }
      };
      
      ws.onclose = (event) => {
        setIsConnected(false);
        console.log('WebSocket fermé:', event.code);
        
        // Reconnexion automatique si ce n'est pas une fermeture normale
        if (event.code !== 1000) {
          setTimeout(() => {
            if (activeTab === 'rooms' || (activeTab === 'dm' && currentConversation)) {
              connectWebSocket();
            }
          }, 5000);
        }
      };
      
      ws.onerror = (error) => {
        console.error('Erreur WebSocket:', error);
        toast({
          title: "Erreur de connexion",
          description: "Impossible de se connecter au serveur de chat",
          variant: "destructive"
        });
      };
      
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          handleWebSocketMessage(data);
        } catch (error) {
          console.error('Erreur parsing message WebSocket:', error);
        }
      };
      
      setSocket(ws);
    } catch (error) {
      console.error('Erreur connexion WebSocket:', error);
      toast({
        title: "Erreur de connexion",
        description: "Impossible de se connecter au serveur de chat",
        variant: "destructive"
      });
    }
  };

  // Gérer les messages WebSocket
  const handleWebSocketMessage = (data: any) => {
    console.log("📥 WS reçu :", data);
    
    if (activeTab === 'rooms') {
      handleRoomMessage(data);
    } else if (activeTab === 'dm') {
      handleDMMessage(data);
    }
  };

  // Gérer les messages de salon
  const handleRoomMessage = (data: any) => {
    if (data.type === "message" && data.data?.room === currentRoom?.name) {
      const newMessage: Message = {
        id: data.data.id?.toString() || Date.now().toString(),
        content: data.data.content,
        senderId: data.data.fromUser?.toString() || '0',
        sender: {
          id: data.data.fromUser?.toString() || '0',
          username: data.data.username || 'Utilisateur'
        },
        createdAt: data.data.timestamp,
        type: 'text'
      };
      setMessages(prev => [...prev, newMessage]);
      scrollToBottom();
    } else if (data.username && data.content) {
      // Format direct du message
      if (data.room === currentRoom?.name || !data.room) {
        const newMessage: Message = {
          id: Date.now().toString(),
          content: data.content,
          senderId: data.fromUser?.toString() || '0',
          sender: {
            id: data.fromUser?.toString() || '0',
            username: data.username
          },
          createdAt: data.timestamp,
          type: 'text'
        };
        setMessages(prev => [...prev, newMessage]);
        scrollToBottom();
      }
    } else if (Array.isArray(data)) {
      // Historique des messages
      const roomMessages = data
        .filter((m: any) => m.room === currentRoom?.name)
        .sort((a: any, b: any) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())
        .map((msg: any) => ({
          id: msg.id?.toString() || Date.now().toString(),
          content: msg.content,
          senderId: msg.fromUser?.toString() || '0',
          sender: {
            id: msg.fromUser?.toString() || '0',
            username: msg.username || 'Utilisateur'
          },
          createdAt: msg.timestamp,
          type: 'text' as const
        }));
      
      setMessages(roomMessages);
      setLoadingMessages(false);
      scrollToBottom();
    } else if (data.type === "room_history") {
      if (data.messages && Array.isArray(data.messages)) {
        const roomMessages = data.messages
          .filter((m: any) => m.room === currentRoom?.name)
          .sort((a: any, b: any) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())
          .map((msg: any) => ({
            id: msg.id?.toString() || Date.now().toString(),
            content: msg.content,
            senderId: msg.fromUser?.toString() || '0',
            sender: {
              id: msg.fromUser?.toString() || '0',
              username: msg.username || 'Utilisateur'
            },
            createdAt: msg.timestamp,
            type: 'text' as const
          }));
        
        setMessages(roomMessages);
      }
      setLoadingMessages(false);
      scrollToBottom();
    }
  };

  // Gérer les messages directs
  const handleDMMessage = (data: any) => {
    if (data.type === "dm" && currentConversation && 
        (data.data?.fromUser === parseInt(currentConversation.userId) || 
         data.data?.to === parseInt(currentConversation.userId))) {
      const newMessage: Message = {
        id: Date.now().toString(),
        content: data.data.content,
        senderId: data.data.fromUser?.toString() || '0',
        sender: {
          id: data.data.fromUser?.toString() || '0',
          username: data.data.username || 'Utilisateur'
        },
        createdAt: data.data.timestamp,
        type: 'text'
      };
      setMessages(prev => [...prev, newMessage]);
      scrollToBottom();
    } else if (data.type === "dm_history") {
      if (data.data && Array.isArray(data.data)) {
        const dmMessages = data.data
          .filter((msg: any) => msg.content)
          .map((msg: any) => ({
            id: msg.id?.toString() || Date.now().toString(),
            content: msg.content,
            senderId: msg.fromUser?.toString() || '0',
            sender: {
              id: msg.fromUser?.toString() || '0',
              username: msg.username || 'Utilisateur'
            },
            createdAt: msg.timestamp,
            type: 'text' as const
          }))
          .sort((a: any, b: any) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());
        
        setMessages(dmMessages);
      } else {
        setMessages([]);
      }
      setLoadingMessages(false);
      scrollToBottom();
    }
  };

  // Rejoindre un salon avec WebSocket
  const joinRoom = async (room: Room) => {
    if (currentRoom?.id === room.id) return;
    
    console.log('[Chat] Rejoindre salon:', room.name);
    setCurrentRoom(room);
    setCurrentConversation(null);
    setMessages([]);
    setLoadingMessages(true);
    
    if (socket && socket.readyState === WebSocket.OPEN) {
      // Rejoindre le salon
      socket.send(JSON.stringify({ 
        type: "join", 
        room: room.name 
      }));
      
      // Demander l'historique après un court délai
      setTimeout(() => {
        if (socket && socket.readyState === WebSocket.OPEN) {
          socket.send(JSON.stringify({ 
            type: "room_history", 
            room: room.name, 
            limit: 50 
          }));
        }
      }, 100);
      
      // Timeout de sécurité
      setTimeout(() => {
        setLoadingMessages(false);
      }, 10000);
    } else {
      // Fallback: charger depuis l'API REST
      await loadRoomMessages(room);
    }
  };

  // Démarrer une conversation directe avec WebSocket
  const startDirectMessage = async (conversation: Conversation) => {
    if (currentConversation?.userId === conversation.userId) return;
    
    console.log('[Chat] Démarrer DM avec:', conversation.username);
    
    // Fermer la connexion WebSocket actuelle proprement
    if (socket) {
      socket.onclose = null;
      socket.onerror = null;
      socket.close();
      setSocket(null);
    }
    
    setCurrentConversation(conversation);
    setCurrentRoom(null);
    setMessages([]);
    setIsConnected(false);
    setLoadingMessages(false);
    
    // Se reconnecter au WebSocket
    await connectWebSocket();
  };

  // Envoyer un message via WebSocket
  const sendMessage = async () => {
    if (!messageInput.trim()) return;
    
    const content = messageInput.trim();
    setMessageInput('');
    
    if (socket && socket.readyState === WebSocket.OPEN) {
      if (activeTab === 'rooms' && currentRoom) {
        // Envoyer message de salon
        socket.send(JSON.stringify({
          type: "message",
          room: currentRoom.name,
          content: content
        }));
      } else if (activeTab === 'dm' && currentConversation) {
        // Envoyer message direct
        const outgoingMessage: Message = {
          id: Date.now().toString(),
          content: content,
          senderId: user?.id?.toString() || '0',
          sender: {
            id: user?.id?.toString() || '0',
            username: user?.username || 'Vous'
          },
          createdAt: new Date().toISOString(),
          type: 'text'
        };
        
        // Ajouter le message localement immédiatement
        setMessages(prev => [...prev, outgoingMessage]);
        
        socket.send(JSON.stringify({
          type: "dm",
          to: parseInt(currentConversation.userId),
          content: content
        }));
      }
    } else {
      // Fallback mode démo
      const newMessage: Message = {
        id: Date.now().toString(),
        content,
        senderId: user?.id?.toString() || '0',
        sender: {
          id: user?.id?.toString() || '0',
          username: user?.username || 'Vous'
        },
        createdAt: new Date().toISOString(),
        type: 'text'
      };
      
      setMessages(prev => [...prev, newMessage]);
      
      toast({
        title: "Mode démonstration",
        description: "Message envoyé en mode local uniquement"
      });
    }
    
    scrollToBottom();
  };

  // Initialisation avec WebSocket
  useEffect(() => {
    const init = async () => {
      try {
        setIsLoading(true);
        
        // Vérifier l'authentification
        const token = await ensureValidToken();
        if (!token) {
          toast({
            title: "Authentification requise",
            description: "Veuillez vous connecter pour accéder au chat",
            variant: "destructive"
          });
          navigate('/auth');
          return;
        }

        // Charger les salons
        await loadRoomsFromBackend();
        
        // Charger les utilisateurs pour les DM
        await loadUsersForDirectMessages();
        
        // Connecter le WebSocket
        await connectWebSocket();
        
      } catch (error) {
        console.error('Erreur d\'initialisation:', error);
        toast({
          title: "Erreur d'initialisation",
          description: "Erreur lors de l'initialisation du chat",
          variant: "destructive"
        });
      } finally {
        setIsLoading(false);
      }
    };

    init();

    // Nettoyage à la fermeture
    return () => {
      if (socket) {
        socket.onclose = null;
        socket.onerror = null;
        socket.close();
      }
    };
  }, []);

  // Rafraîchir le token périodiquement
  useEffect(() => {
    const interval = setInterval(async () => {
      await ensureValidToken();
    }, 4 * 60 * 1000); // Toutes les 4 minutes

    return () => clearInterval(interval);
  }, []);

  // Charger les salons depuis le backend (utiliser les bons endpoints)
  const loadRoomsFromBackend = async () => {
    try {
      const token = await ensureValidToken();
      if (!token) return;

      // Essayer d'abord l'endpoint de l'ancien chat
      let response = await fetch('/chat/rooms', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      // Si ça échoue, essayer le nouvel endpoint
      if (!response.ok) {
        response = await fetch('/api/v1/rooms', {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        });
      }

      if (!response.ok) {
        throw new Error(`Erreur HTTP: ${response.status}`);
      }

      const data = await response.json();
      console.log('Salons chargés:', data);

      let roomsData = [];
      
      // Gérer les deux formats de réponse
      if (Array.isArray(data)) {
        // Format de l'ancien chat
        roomsData = data.map((room: any) => ({
          id: room.id?.toString() || room.name,
          name: room.name,
          description: room.description || '',
          memberCount: room.user_count || room.member_count || 0
        }));
      } else if (data.success && data.data) {
        // Format du nouveau backend
        roomsData = data.data.map((room: any) => ({
          id: room.id?.toString() || room.name,
          name: room.name,
          description: room.description || '',
          memberCount: room.member_count || 0
        }));
      }
      
      setRooms(roomsData);
      
      // Sélectionner le salon "general" par défaut
      const generalRoom = roomsData.find((r: Room) => r.name === 'general');
      if (generalRoom && !currentRoom) {
        setCurrentRoom(generalRoom);
        await loadRoomMessages(generalRoom);
      }
    } catch (error) {
      console.error('Erreur lors du chargement des salons:', error);
      toast({
        title: "Erreur de chargement",
        description: "Impossible de charger les salons",
        variant: "destructive"
      });
    }
  };

  // Charger les utilisateurs pour les messages directs
  const loadUsersForDirectMessages = async () => {
    try {
      const token = await ensureValidToken();
      if (!token) return;

      const response = await fetch('/users/except-me', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error(`Erreur HTTP: ${response.status}`);
      }

      const data = await response.json();
      console.log('Utilisateurs chargés:', data);

      let conversationsData = [];
      
      // Gérer les deux formats de réponse
      if (Array.isArray(data)) {
        // Format de l'ancien chat
        conversationsData = data.map((user: any) => ({
          userId: user.id?.toString() || '0',
          username: user.username || 'Utilisateur',
          displayName: user.username || 'Utilisateur',
          isOnline: user.isOnline || false,
          unreadCount: 0
        }));
      } else if (data.success && data.data) {
        // Format du nouveau backend
        conversationsData = data.data.map((user: any) => ({
          userId: user.id?.toString() || '0',
          username: user.username || 'Utilisateur',
          displayName: extractNullString(user.first_name) || extractNullString(user.last_name) || user.username,
          isOnline: false,
          unreadCount: 0
        }));
      }
      
      setConversations(conversationsData);
    } catch (error) {
      console.error('Erreur lors du chargement des utilisateurs:', error);
      toast({
        title: "Erreur de chargement",
        description: "Impossible de charger les utilisateurs",
        variant: "destructive"
      });
    }
  };

  // Charger les messages d'un salon
  const loadRoomMessages = async (room: Room) => {
    try {
      const token = await ensureValidToken();
      if (!token) return;

      setLoadingMessages(true);
      const response = await fetch(`/api/v1/rooms/${room.name}/messages`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error(`Erreur HTTP: ${response.status}`);
      }

      const data = await response.json();
      console.log('Messages du salon chargés:', data);

      if (data.success && data.data) {
        const messagesData = data.data.map((msg: any) => ({
          id: msg.id?.toString() || Date.now().toString(),
          content: msg.content,
          senderId: msg.user_id?.toString() || '0',
          sender: {
            id: msg.user_id?.toString() || '0',
            username: msg.username || 'Utilisateur'
          },
          createdAt: msg.created_at,
          type: 'text' as const
        }));
        
        setMessages(messagesData);
      }
    } catch (error) {
      console.error('Erreur lors du chargement des messages:', error);
      toast({
        title: "Erreur de chargement",
        description: "Impossible de charger les messages",
        variant: "destructive"
      });
    } finally {
      setLoadingMessages(false);
    }
  };

  // Créer un nouveau salon
  const createRoom = async () => {
    if (!newRoomName.trim()) return;
    
    toast({
      title: "Mode démonstration",
      description: "La création de salons nécessite le serveur backend"
    });
    
    setShowNewRoomForm(false);
    setNewRoomName('');
  };

  // Formater l'heure d'un message
  const formatMessageTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString('fr-FR', { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  };

  // Obtenir le nom d'affichage d'un utilisateur
  const getUserDisplayName = (sender: any) => {
    if (sender.username) return sender.username;
    const firstName = extractNullString(sender.first_name);
    const lastName = extractNullString(sender.last_name);
    return firstName || lastName || 'Utilisateur';
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-2 text-gray-600">Initialisation du chat...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Sidebar */}
      <div className="w-80 bg-white border-r flex flex-col">
        {/* Header */}
        <div className="p-4 border-b">
          <div className="flex items-center justify-between">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate('/dashboard')}
              className="flex items-center gap-2"
            >
              <ArrowLeft className="w-4 h-4" />
              Retour
            </Button>
            <div className="flex items-center gap-2">
              <Circle className={`w-2 h-2 ${isConnected ? 'fill-green-500 text-green-500' : 'fill-orange-500 text-orange-500'}`} />
              <span className="text-sm text-gray-600">
                {isConnected ? 'Connecté' : 'Mode démo'}
              </span>
            </div>
          </div>
        </div>
        
        {/* Tabs */}
        <div className="flex border-b">
          <button
            onClick={() => setActiveTab('rooms')}
            className={`flex-1 p-3 text-sm font-medium ${
              activeTab === 'rooms' 
                ? 'text-blue-600 border-b-2 border-blue-600' 
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            <Hash className="w-4 h-4 inline mr-1" />
            Salons
          </button>
          <button
            onClick={() => setActiveTab('dm')}
            className={`flex-1 p-3 text-sm font-medium ${
              activeTab === 'dm' 
                ? 'text-blue-600 border-b-2 border-blue-600' 
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            <MessageSquare className="w-4 h-4 inline mr-1" />
            Messages
          </button>
        </div>
        
        {/* Content */}
        <ScrollArea className="flex-1">
          {activeTab === 'rooms' ? (
            <div className="p-2">
              {/* Nouveau salon */}
              {showNewRoomForm ? (
                <Card className="mb-4">
                  <CardContent className="p-3">
                    <Input
                      placeholder="Nom du salon"
                      value={newRoomName}
                      onChange={(e) => setNewRoomName(e.target.value)}
                      onKeyDown={(e) => e.key === 'Enter' && createRoom()}
                      className="mb-2"
                    />
                    <div className="flex gap-2">
                      <Button size="sm" onClick={createRoom}>
                        Créer
                      </Button>
                      <Button size="sm" variant="outline" onClick={() => setShowNewRoomForm(false)}>
                        Annuler
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ) : (
                <Button 
                  variant="outline" 
                  size="sm" 
                  className="w-full mb-4"
                  onClick={() => setShowNewRoomForm(true)}
                >
                  <Plus className="w-4 h-4 mr-2" />
                  Nouveau salon
                </Button>
              )}
              
              {/* Liste des salons */}
              {rooms.map((room) => (
                <div
                  key={room.id}
                  onClick={() => joinRoom(room)}
                  className={`p-3 rounded-lg cursor-pointer mb-1 ${
                    currentRoom?.id === room.id 
                      ? 'bg-blue-100 text-blue-900' 
                      : 'hover:bg-gray-100'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Hash className="w-4 h-4" />
                      <span className="font-medium">{room.name}</span>
                    </div>
                    <Badge variant="secondary" className="text-xs">
                      {room.memberCount || 0}
                    </Badge>
                  </div>
                  {room.description && (
                    <p className="text-sm text-gray-600 mt-1 truncate">
                      {room.description}
                    </p>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <div className="p-2">
              {/* Liste des conversations */}
              {conversations.map((conv) => (
                <div
                  key={conv.userId}
                  onClick={() => startDirectMessage(conv)}
                  className={`p-3 rounded-lg cursor-pointer mb-1 ${
                    currentConversation?.userId === conv.userId 
                      ? 'bg-blue-100 text-blue-900' 
                      : 'hover:bg-gray-100'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="relative">
                        <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center">
                          <span className="text-sm font-semibold text-white">
                            {conv.displayName?.charAt(0).toUpperCase() || conv.username.charAt(0).toUpperCase()}
                          </span>
                        </div>
                        <div className="absolute -bottom-1 -right-1">
                          <UserStatusIndicator 
                            isOnline={conv.isOnline || false} 
                            lastSeen={conv.lastSeen} 
                            size="sm"
                          />
                        </div>
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="font-medium text-gray-900 truncate">
                          {conv.displayName || conv.username}
                        </div>
                        {conv.lastMessage && (
                          <div className="text-sm text-gray-500 truncate">
                            {typeof conv.lastMessage === 'string' ? conv.lastMessage : conv.lastMessage.content}
                          </div>
                        )}
                      </div>
                    </div>
                    <div className="flex flex-col items-end gap-1">
                      {conv.lastMessageTime && (
                        <span className="text-xs text-gray-400">
                          {formatMessageTime(conv.lastMessageTime)}
                        </span>
                      )}
                      <UnreadBadge count={conv.unreadCount} size="sm" />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
      </div>
      
      {/* Main chat area */}
      <div className="flex-1 flex flex-col">
        {currentRoom || currentConversation ? (
          <>
            {/* Chat header */}
            <div className="p-4 border-b bg-white">
              <div className="flex items-center justify-between">
                <div>
                  <h1 className="text-lg font-semibold">
                    {currentRoom ? (
                      <><Hash className="w-5 h-5 inline mr-1" />{currentRoom.name}</>
                    ) : (
                      <><MessageSquare className="w-5 h-5 inline mr-1" />{currentConversation?.username}</>
                    )}
                  </h1>
                  <p className="text-sm text-gray-600">
                    {currentRoom ? currentRoom.description : 'Message direct'}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  {currentRoom && (
                    <Badge variant="outline" className="flex items-center gap-1">
                      <Users className="w-3 h-3" />
                      {currentRoom.memberCount || 0} membre(s)
                    </Badge>
                  )}
                  <Button variant="ghost" size="sm">
                    <Settings className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </div>
            
            {/* Messages area */}
            <ScrollArea className="flex-1 p-4">
              <div className="space-y-4">
                {loadingMessages ? (
                  <div className="text-center py-4">
                    <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600 mx-auto"></div>
                    <p className="mt-2 text-sm text-gray-600">Chargement des messages...</p>
                  </div>
                ) : (
                  messages.map((message, index) => (
                    <div key={message.id || index} className={`flex gap-3 ${message.type === 'system' ? 'justify-center' : ''}`}>
                      {message.type !== 'system' && (
                        <div className="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center flex-shrink-0">
                          <span className="text-sm font-medium">
                            {getUserDisplayName(message.sender).charAt(0).toUpperCase()}
                          </span>
                        </div>
                      )}
                      <div className={`flex-1 ${message.type === 'system' ? 'text-center' : ''}`}>
                        {message.type === 'system' ? (
                          <div className="text-sm text-gray-500 italic bg-gray-100 rounded-lg px-3 py-1 inline-block">
                            {message.content}
                          </div>
                        ) : (
                          <>
                            <div className="flex items-center gap-2 mb-1">
                              <span className="font-medium text-sm">
                                {getUserDisplayName(message.sender)}
                              </span>
                              <span className="text-xs text-gray-500">
                                {formatMessageTime(message.createdAt)}
                              </span>
                            </div>
                            <p className="text-gray-900">{message.content}</p>
                          </>
                        )}
                      </div>
                    </div>
                  ))
                )}
                
                {/* Indicateur de frappe */}
                {typingUsers.length > 0 && (
                  <div className="text-sm text-gray-500 italic">
                    {typingUsers.length === 1 
                      ? `${typingUsers[0]} est en train d'écrire...`
                      : `${typingUsers.length} personnes écrivent...`
                    }
                  </div>
                )}
                
                <div ref={messagesEndRef} />
              </div>
            </ScrollArea>
            
            {/* Message input */}
            <div className="p-4 border-t bg-white">
              <div className="flex gap-2">
                <Input
                  value={messageInput}
                  onChange={(e) => setMessageInput(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
                  placeholder={`Écrire un message ${currentRoom ? `dans #${currentRoom.name}` : `à ${currentConversation?.username}`}...`}
                  className="flex-1"
                />
                <Button onClick={sendMessage} disabled={!messageInput.trim()}>
                  <Send className="w-4 h-4" />
                </Button>
              </div>
            </div>
          </>
        ) : (
          /* État vide */
          <div className="flex-1 flex items-center justify-center text-center">
            <div>
              <MessageSquare className="w-16 h-16 text-gray-400 mx-auto mb-4" />
              <h2 className="text-xl font-semibold text-gray-900 mb-2">
                Bienvenue dans le chat !
              </h2>
              <p className="text-gray-600 max-w-md">
                Sélectionnez un salon ou démarrez une conversation pour commencer à discuter.
                {!isConnected && (
                  <><br /><span className="text-orange-600 font-medium">Mode démonstration - connectez le serveur Rust sur localhost:9001.</span></>
                )}
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// Export par défaut pour la compatibilité avec les imports
export default ChatPage;