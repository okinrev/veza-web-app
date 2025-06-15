import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "@/shared/stores/authStore";
import { 
  MessageSquare, 
  FileText, 
  Music, 
  Users, 
  BookOpen, 
  ShoppingBag,
  Bell,
  Target,
  TrendingUp,
  Calendar,
  Clock,
  Star,
  Download,
  Share2,
  Heart
} from "lucide-react";

interface Activity {
  id: number;
  type: 'resource' | 'track' | 'message' | 'product';
  title: string;
  description: string;
  timestamp: string;
  user: string;
  icon: JSX.Element;
}

interface Goal {
  id: number;
  title: string;
  progress: number;
  target: number;
  unit: string;
  deadline: string;
}

interface Notification {
  id: number;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  timestamp: string;
  read: boolean;
}

export function DashboardPage() {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [notifications, setNotifications] = useState<Notification[]>([
    {
      id: 1,
      type: 'success',
      title: 'Nouveau message',
      message: 'Vous avez reçu un nouveau message de John Doe',
      timestamp: '2024-03-15T10:30:00',
      read: false
    },
    {
      id: 2,
      type: 'info',
      title: 'Mise à jour disponible',
      message: 'Une nouvelle version de l\'application est disponible',
      timestamp: '2024-03-15T09:15:00',
      read: true
    },
    {
      id: 3,
      type: 'warning',
      title: 'Stock faible',
      message: 'Le stock de certains produits est faible',
      timestamp: '2024-03-14T16:45:00',
      read: false
    }
  ]);

  const goals: Goal[] = [
    {
      id: 1,
      title: "Tracks écoutés",
      progress: 75,
      target: 100,
      unit: "tracks",
      deadline: "2024-04-01"
    },
    {
      id: 2,
      title: "Ressources téléchargées",
      progress: 45,
      target: 50,
      unit: "ressources",
      deadline: "2024-03-31"
    },
    {
      id: 3,
      title: "Messages envoyés",
      progress: 120,
      target: 150,
      unit: "messages",
      deadline: "2024-03-25"
    }
  ];

  const activities: Activity[] = [
    {
      id: 1,
      type: 'resource',
      title: 'Nouvelle ressource partagée',
      description: 'John Doe a partagé une nouvelle ressource',
      timestamp: '2024-03-15T10:30:00',
      user: 'John Doe',
      icon: <FileText className="h-4 w-4" />
    },
    {
      id: 2,
      type: 'track',
      title: 'Nouveau track ajouté',
      description: 'Jane Smith a ajouté un nouveau track',
      timestamp: '2024-03-15T09:15:00',
      user: 'Jane Smith',
      icon: <Music className="h-4 w-4" />
    },
    {
      id: 3,
      type: 'message',
      title: 'Nouveau message',
      description: 'Mike Johnson vous a envoyé un message',
      timestamp: '2024-03-14T16:45:00',
      user: 'Mike Johnson',
      icon: <MessageSquare className="h-4 w-4" />
    }
  ];

  const quickActions = [
    {
      title: "Chat",
      description: "Discuter avec d'autres utilisateurs",
      icon: <MessageSquare className="h-6 w-6" />,
      path: "/chat",
      color: "bg-blue-500"
    },
    {
      title: "Ressources",
      description: "Accéder aux ressources partagées",
      icon: <FileText className="h-6 w-6" />,
      path: "/resources",
      color: "bg-green-500"
    },
    {
      title: "Tracks",
      description: "Explorer les tracks",
      icon: <Music className="h-6 w-6" />,
      path: "/tracks",
      color: "bg-purple-500"
    },
    {
      title: "Produits",
      description: "Voir les produits disponibles",
      icon: <ShoppingBag className="h-6 w-6" />,
      path: "/products",
      color: "bg-orange-500"
    },
    {
      title: "Communauté",
      description: "Découvrir la communauté",
      icon: <Users className="h-6 w-6" />,
      path: "/community",
      color: "bg-pink-500"
    },
    {
      title: "Documentation",
      description: "Consulter la documentation",
      icon: <BookOpen className="h-6 w-6" />,
      path: "/docs",
      color: "bg-indigo-500"
    }
  ];

  const markNotificationAsRead = (id: number) => {
    setNotifications(prevNotifications =>
      prevNotifications.map(notification =>
        notification.id === id
          ? { ...notification, read: true }
          : notification
      )
    );
  };

  const unreadNotifications = notifications.filter(n => !n.read).length;

  return (
    <div className="p-6 space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Tableau de bord</h1>
        <div className="flex items-center space-x-4">
          <div className="text-sm text-muted-foreground">
            Bienvenue, {user?.username} !
          </div>
          <Button variant="ghost" size="icon" className="relative">
            <Bell className="h-5 w-5" />
            {unreadNotifications > 0 && (
              <Badge
                variant="destructive"
                className="absolute -top-1 -right-1 h-5 w-5 flex items-center justify-center p-0"
              >
                {unreadNotifications}
              </Badge>
            )}
          </Button>
        </div>
      </div>

      {/* Statistiques */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Messages</CardTitle>
            <MessageSquare className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">24</div>
            <p className="text-xs text-muted-foreground">+12% depuis hier</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Ressources</CardTitle>
            <FileText className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">156</div>
            <p className="text-xs text-muted-foreground">+8% cette semaine</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Tracks</CardTitle>
            <Music className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">42</div>
            <p className="text-xs text-muted-foreground">+5% ce mois</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Membres actifs</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">1,234</div>
            <p className="text-xs text-muted-foreground">+23% cette semaine</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-7">
        {/* Actions rapides */}
        <div className="col-span-4">
          <div className="grid gap-6 md:grid-cols-2">
            {quickActions.map((action) => (
              <Card key={action.title} className="hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="flex items-center space-x-4">
                    <div className={`p-2 rounded-lg ${action.color} text-white`}>
                      {action.icon}
                    </div>
                    <div>
                      <CardTitle className="text-lg">{action.title}</CardTitle>
                      <p className="text-sm text-muted-foreground">{action.description}</p>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <Button 
                    variant="outline" 
                    className="w-full"
                    onClick={() => navigate(action.path)}
                  >
                    Accéder
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>

        {/* Objectifs */}
        <div className="col-span-3">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Target className="h-5 w-5" />
                <span>Objectifs</span>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-6">
                {goals.map((goal) => (
                  <div key={goal.id} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium">{goal.title}</p>
                        <p className="text-sm text-muted-foreground">
                          {goal.progress} / {goal.target} {goal.unit}
                        </p>
                      </div>
                      <div className="text-sm text-muted-foreground">
                        <Calendar className="h-4 w-4 inline mr-1" />
                        {new Date(goal.deadline).toLocaleDateString()}
                      </div>
                    </div>
                    <Progress value={(goal.progress / goal.target) * 100} className="h-2" />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-7">
        {/* Activités récentes */}
        <div className="col-span-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Clock className="h-5 w-5" />
                <span>Activités récentes</span>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-[300px]">
                <div className="space-y-4">
                  {activities.map((activity) => (
                    <div key={activity.id} className="flex items-start space-x-4 p-4 border rounded-lg">
                      <div className={`p-2 rounded-lg ${
                        activity.type === 'resource' ? 'bg-green-100' :
                        activity.type === 'track' ? 'bg-purple-100' :
                        activity.type === 'message' ? 'bg-blue-100' :
                        'bg-orange-100'
                      }`}>
                        {activity.icon}
                      </div>
                      <div className="flex-1">
                        <div className="flex items-center justify-between">
                          <p className="font-medium">{activity.title}</p>
                          <span className="text-sm text-muted-foreground">
                            {new Date(activity.timestamp).toLocaleTimeString()}
                          </span>
                        </div>
                        <p className="text-sm text-muted-foreground">{activity.description}</p>
                        <div className="flex items-center space-x-2 mt-2">
                          <Badge variant="secondary">{activity.user}</Badge>
                          <Badge variant="outline">{activity.type}</Badge>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        </div>

        {/* Notifications */}
        <div className="col-span-3">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Bell className="h-5 w-5" />
                <span>Notifications</span>
                {unreadNotifications > 0 && (
                  <Badge variant="destructive" className="ml-2">
                    {unreadNotifications}
                  </Badge>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-[300px]">
                <div className="space-y-4">
                  {notifications.map((notification) => (
                    <div
                      key={notification.id}
                      className={`p-4 border rounded-lg ${
                        !notification.read ? 'bg-muted/50' : ''
                      }`}
                      onClick={() => markNotificationAsRead(notification.id)}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-2">
                          <Badge
                            variant={
                              notification.type === 'success' ? 'default' :
                              notification.type === 'warning' ? 'secondary' :
                              notification.type === 'error' ? 'destructive' :
                              'outline'
                            }
                          >
                            {notification.type}
                          </Badge>
                          <p className="font-medium">{notification.title}</p>
                        </div>
                        <span className="text-sm text-muted-foreground">
                          {new Date(notification.timestamp).toLocaleTimeString()}
                        </span>
                      </div>
                      <p className="text-sm text-muted-foreground mt-2">
                        {notification.message}
                      </p>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
} 