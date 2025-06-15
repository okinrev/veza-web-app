import { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "@/shared/stores/authStore";
import { api } from "@/shared/api/api";
import { wsClient } from "@/shared/api/websocket";
import { Button, type ButtonProps } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge, type BadgeProps } from "@/components/ui/badge";
import { toast } from "@/components/ui/use-toast";
import { 
  ArrowLeft, 
  Users, 
  MessageSquare, 
  Send, 
  RefreshCw, 
  Bell, 
  LogOut,
  Search,
  Plus,
  Trash2
} from "lucide-react";

interface Room {
  id: string;
  name: string;
  description?: string;
  user_count: number;
}

interface User {
  id: string;
  username: string;
  email: string;
  isOnline: boolean;
}

interface Message {
  id: string;
  content: string;
  fromUser: string;
  username?: string;
  timestamp: string;
  status?: 'sent' | 'delivered' | 'read';
}

interface ChatStats {
  totalRooms: number;
  activeUsers: number;
  todayMessages: number;
}

interface DMStats {
  totalMessages: number;
  todayMessages: number;
}

export function ChatPage() {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();
  const [activeTab, setActiveTab] = useState<'rooms' | 'dm'>('rooms');
  const [isConnected, setIsConnected] = useState(false);
  const [currentRoom, setCurrentRoom] = useState<string | null>(null);
  const [otherUserId, setOtherUserId] = useState<string | null>(null);
  const [rooms, setRooms] = useState<Room[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [filteredUsers, setFilteredUsers] = useState<User[]>([]);
  const [messages, setMessages] = useState<Message[]>([]);
  const [messageContent, setMessageContent] = useState('');
  const [isTyping, setIsTyping] = useState<string[]>([]);
  const [loadingMessages, setLoadingMessages] = useState(false);
  const [loadingUsers, setLoadingUsers] = useState(false);
  const [sending, setSending] = useState(false);
  const [creating, setCreating] = useState(false);
  const [newRoomName, setNewRoomName] = useState('');
  const [userSearch, setUserSearch] = useState('');
  const [notificationsEnabled, setNotificationsEnabled] = useState(true);
  const [notifications, setNotifications] = useState<Array<{ id: number; type: 'success' | 'error' | 'info'; message: string; show: boolean }>>([]);
  const [roomStats, setRoomStats] = useState<ChatStats>({ totalRooms: 0, activeUsers: 0, todayMessages: 0 });
  const [dmStats, setDMStats] = useState<DMStats>({ totalMessages: 0, todayMessages: 0 });
  const [otherUserInfo, setOtherUserInfo] = useState<User | null>(null);
  const [connectedUsers, setConnectedUsers] = useState<User[]>([]);
  const [recentConversations, setRecentConversations] = useState<Array<{ userId: string; username: string; lastMessage: string }>>([]);
  
  const messagesContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!user) {
      navigate('/login');
      return;
    }

    initializeWebSocket();
    loadInitialData();

    return () => {
      wsClient.disconnect();
    };
  }, [user, navigate]);

  const initializeWebSocket = () => {
    wsClient.connect();
    
    wsClient.subscribe('message', (data) => {
      setMessages(prev => [...prev, data.message]);
      scrollToBottom();
    });

    wsClient.subscribe('typing', (data) => {
      setIsTyping(prev => [...prev, data.username]);
      setTimeout(() => {
        setIsTyping(prev => prev.filter(u => u !== data.username));
      }, 3000);
    });

    wsClient.subscribe('user_joined', (data) => {
      setConnectedUsers(prev => [...prev, data.user]);
    });

    wsClient.subscribe('user_left', (data) => {
      setConnectedUsers(prev => prev.filter(u => u.id !== data.userId));
    });

    wsClient.subscribe('room_stats', (data) => {
      setRoomStats(data.stats);
    });

    wsClient.subscribe('dm_stats', (data) => {
      setDMStats(data.stats);
    });
  };

  const loadInitialData = async () => {
    try {
      const usersRes = await api.get('/users/except-me');
      setUsers(usersRes.data.data);
      setFilteredUsers(usersRes.data.data);

      setRooms([
        { id: 'general', name: 'Général', description: 'Salon général', user_count: 0 },
        { id: 'help', name: 'Aide', description: 'Support et aide', user_count: 0 }
      ]);
    } catch (error) {
      showNotification('error', 'Erreur lors du chargement des données');
    }
  };

  const joinRoom = async (roomName: string) => {
    try {
      setCurrentRoom(roomName);
      setLoadingMessages(true);
      
      wsClient.send({
        type: 'room_history',
        room: roomName,
        limit: 50
      });

      wsClient.send({
        type: 'join',
        room: roomName
      });
    } catch (error) {
      showNotification('error', 'Erreur lors de la connexion au salon');
    } finally {
      setLoadingMessages(false);
    }
  };

  const leaveRoom = () => {
    if (currentRoom) {
      wsClient.send({
        type: 'leave_room',
        room: currentRoom
      });
    }
    setCurrentRoom(null);
    setMessages([]);
    setConnectedUsers([]);
  };

  const createRoom = async () => {
    try {
      setCreating(true);
      wsClient.send({
        type: 'create_room',
        name: newRoomName,
        description: "Nouveau salon",
        is_private: false
      });
      setNewRoomName('');
      showNotification('success', 'Salon créé avec succès');
    } catch (error) {
      showNotification('error', 'Erreur lors de la création du salon');
    } finally {
      setCreating(false);
    }
  };

  const sendMessage = async () => {
    if (!messageContent.trim() || sending) return;

    try {
      setSending(true);
      const endpoint = activeTab === 'rooms' 
        ? `/chat/rooms/${currentRoom}/messages`
        : `/chat/messages/${otherUserId}`;

      await api.post(endpoint, { content: messageContent });
      setMessageContent('');
    } catch (error) {
      showNotification('error', 'Erreur lors de l\'envoi du message');
    } finally {
      setSending(false);
    }
  };

  const handleTyping = () => {
    wsClient.send({
      type: 'typing',
      room: currentRoom,
      userId: otherUserId
    });
  };

  const searchUsers = () => {
    const filtered = users.filter(user => 
      user.username.toLowerCase().includes(userSearch.toLowerCase()) ||
      user.email.toLowerCase().includes(userSearch.toLowerCase())
    );
    setFilteredUsers(filtered);
  };

  const selectUser = async (userId: string) => {
    try {
      setOtherUserId(userId);
      setLoadingMessages(true);
      
      wsClient.send({
        type: 'dm_history',
        with: userId,
        limit: 50
      });
    } catch (error) {
      showNotification('error', 'Erreur lors du chargement des messages');
    } finally {
      setLoadingMessages(false);
    }
  };

  const clearChat = () => {
    setMessages([]);
    setOtherUserId(null);
    setOtherUserInfo(null);
  };

  const toggleNotifications = () => {
    setNotificationsEnabled(!notificationsEnabled);
    showNotification('info', `Notifications ${notificationsEnabled ? 'désactivées' : 'activées'}`);
  };

  const showNotification = (type: 'success' | 'error' | 'info', message: string) => {
    const id = Date.now();
    setNotifications(prev => [...prev, { id, type, message, show: true }]);
    setTimeout(() => {
      setNotifications(prev => prev.map(n => n.id === id ? { ...n, show: false } : n));
    }, 3000);
  };

  const scrollToBottom = () => {
    if (messagesContainerRef.current) {
      messagesContainerRef.current.scrollTop = messagesContainerRef.current.scrollHeight;
    }
  };

  const formatTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString('fr-FR', { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  };

  return (
    <div className="flex h-screen">
      {/* Sidebar */}
      <div className="w-80 border-r bg-background">
        <div className="flex h-14 items-center border-b px-4">
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={() => navigate('/')}
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <h2 className="ml-2 text-lg font-semibold">Chat</h2>
        </div>

        {/* Tabs */}
        <div className="flex border-b">
          <Button
            type="button"
            variant={activeTab === 'rooms' ? 'default' : 'ghost'}
            className="flex-1"
            onClick={() => setActiveTab('rooms')}
          >
            <MessageSquare className="h-4 w-4" />
            Salons
          </Button>
          <Button
            type="button"
            variant={activeTab === 'dm' ? 'default' : 'ghost'}
            className="flex-1"
            onClick={() => setActiveTab('dm')}
          >
            <Users className="h-4 w-4" />
            Messages
          </Button>
        </div>

        {/* Content */}
        <ScrollArea className="h-[calc(100vh-8rem)]">
          {activeTab === 'rooms' ? (
            <div className="p-4">
              <div className="mb-4 flex items-center gap-2">
                <Input
                  placeholder="Nouveau salon..."
                  value={newRoomName}
                  onChange={(e) => setNewRoomName(e.target.value)}
                />
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  onClick={createRoom}
                  disabled={creating || !newRoomName.trim()}
                >
                  <Plus className="h-4 w-4" />
                </Button>
              </div>

              <div className="space-y-2">
                {rooms.map((room) => (
                  <Button
                    key={room.id}
                    type="button"
                    variant={currentRoom === room.id ? 'default' : 'ghost'}
                    className="w-full justify-start"
                    onClick={() => joinRoom(room.id)}
                  >
                    <MessageSquare className="h-4 w-4" />
                    <span className="flex-1">{room.name}</span>
                    <Badge {...({variant: 'secondary'} as BadgeProps)}>{room.user_count}</Badge>
                  </Button>
                ))}
              </div>
            </div>
          ) : (
            <div className="p-4">
              <div className="mb-4">
                <Input
                  placeholder="Rechercher un utilisateur..."
                  value={userSearch}
                  onChange={(e) => {
                    setUserSearch(e.target.value);
                    searchUsers();
                  }}
                />
              </div>

              <div className="space-y-2">
                {filteredUsers.map((user) => (
                  <Button
                    key={user.id}
                    type="button"
                    variant={otherUserId === user.id ? 'default' : 'ghost'}
                    className="w-full justify-start"
                    onClick={() => selectUser(user.id)}
                  >
                    <Users className="h-4 w-4" />
                    <span className="flex-1">{user.username}</span>
                    {user.isOnline && (
                      <Badge {...({variant: 'default'} as BadgeProps)}>En ligne</Badge>
                    )}
                  </Button>
                ))}
              </div>
            </div>
          )}
        </ScrollArea>
      </div>

      {/* Main chat area */}
      <div className="flex flex-1 flex-col">
        {currentRoom || otherUserId ? (
          <>
            <div className="flex h-14 items-center justify-between border-b px-4">
              <div className="flex items-center gap-2">
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={leaveRoom}
                >
                  <ArrowLeft className="h-4 w-4" />
                </Button>
                <h3 className="font-semibold">
                  {currentRoom
                    ? rooms.find((r) => r.id === currentRoom)?.name
                    : otherUserInfo?.username}
                </h3>
              </div>
              <div className="flex items-center gap-2">
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={toggleNotifications}
                >
                  <Bell className="h-4 w-4" />
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={clearChat}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </div>

            <ScrollArea
              ref={messagesContainerRef}
              className="flex-1 p-4"
            >
              {loadingMessages ? (
                <div className="flex h-full items-center justify-center">
                  <RefreshCw className="h-6 w-6 animate-spin" />
                </div>
              ) : (
                <div className="space-y-4">
                  {messages.map((message) => (
                    <div
                      key={message.id}
                      className={`flex ${
                        message.fromUser === user?.id
                          ? 'justify-end'
                          : 'justify-start'
                      }`}
                    >
                      <div
                        className={`max-w-[80%] rounded-lg p-3 ${
                          message.fromUser === user?.id
                            ? 'bg-primary text-primary-foreground'
                            : 'bg-muted'
                        }`}
                      >
                        <div className="mb-1 text-xs opacity-70">
                          {message.username}
                        </div>
                        <div>{message.content}</div>
                        <div className="mt-1 text-xs opacity-70">
                          {formatTime(message.timestamp)}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </ScrollArea>

            <div className="border-t p-4">
              <div className="flex gap-2">
                <Input
                  placeholder="Écrivez votre message..."
                  value={messageContent}
                  onChange={(e) => setMessageContent(e.target.value)}
                  onKeyPress={(e) => {
                    if (e.key === 'Enter' && !e.shiftKey) {
                      e.preventDefault();
                      sendMessage();
                    }
                  }}
                  onKeyDown={handleTyping}
                />
                <Button
                  type="button"
                  variant="default"
                  size="icon"
                  onClick={sendMessage}
                  disabled={sending || !messageContent.trim()}
                >
                  <Send className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </>
        ) : (
          <div className="flex h-full items-center justify-center">
            <div className="text-center">
              <MessageSquare className="mx-auto h-12 w-12 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-semibold">
                Sélectionnez un salon ou un utilisateur
              </h3>
              <p className="mt-2 text-sm text-muted-foreground">
                Commencez à discuter en rejoignant un salon ou en envoyant un message direct
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Notifications */}
      <div className="fixed bottom-4 right-4 space-y-2 z-50">
        {notifications.map((notification) => (
          <div
            key={notification.id}
            className={`${
              notification.type === 'success'
                ? 'bg-green-500'
                : notification.type === 'info'
                ? 'bg-blue-500'
                : 'bg-red-500'
            } text-white px-6 py-3 rounded-lg shadow-lg max-w-sm transition-opacity duration-300 ${
              notification.show ? 'opacity-100' : 'opacity-0'
            }`}
          >
            {notification.message}
          </div>
        ))}
      </div>
    </div>
  );
} 