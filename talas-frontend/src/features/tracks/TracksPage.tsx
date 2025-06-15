import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Slider } from "@/components/ui/slider";
import { 
  Search, 
  Filter, 
  Play, 
  Pause, 
  Heart, 
  Share2, 
  Clock, 
  Users,
  Music,
  ListMusic,
  Plus,
  MoreVertical,
  Download,
  Volume2,
  VolumeX
} from "lucide-react";

interface Track {
  id: number;
  title: string;
  artist: string;
  duration: string;
  genre: string;
  bpm: number;
  key: string;
  progress: number;
  likes: number;
  plays: number;
  coverUrl?: string;
  tags: string[];
  createdAt: string;
  audioUrl?: string;
}

interface Playlist {
  id: number;
  name: string;
  description: string;
  tracks: Track[];
  coverUrl?: string;
  createdAt: string;
}

export function TracksPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedGenre, setSelectedGenre] = useState<string | null>(null);
  const [selectedBpm, setSelectedBpm] = useState<string | null>(null);
  const [selectedKey, setSelectedKey] = useState<string | null>(null);
  const [isPlaying, setIsPlaying] = useState<number | null>(null);
  const [currentTrack, setCurrentTrack] = useState<Track | null>(null);
  const [volume, setVolume] = useState(100);
  const [isMuted, setIsMuted] = useState(false);
  const [showCreatePlaylist, setShowCreatePlaylist] = useState(false);
  const [newPlaylistName, setNewPlaylistName] = useState("");
  const [newPlaylistDescription, setNewPlaylistDescription] = useState("");

  // Données de test
  const tracks: Track[] = [
    {
      id: 1,
      title: "Summer Vibes",
      artist: "DJ Cool",
      duration: "3:45",
      genre: "House",
      bpm: 128,
      key: "Am",
      progress: 75,
      likes: 1234,
      plays: 5678,
      coverUrl: "https://picsum.photos/200",
      audioUrl: "https://example.com/audio1.mp3",
      tags: ["house", "summer", "dance"],
      createdAt: "2024-03-15"
    },
    {
      id: 2,
      title: "Midnight Dreams",
      artist: "Luna",
      duration: "4:20",
      genre: "Techno",
      bpm: 135,
      key: "Fm",
      progress: 45,
      likes: 856,
      plays: 2345,
      coverUrl: "https://picsum.photos/201",
      audioUrl: "https://example.com/audio2.mp3",
      tags: ["techno", "dark", "underground"],
      createdAt: "2024-03-14"
    },
    {
      id: 3,
      title: "Sunset Groove",
      artist: "Groove Master",
      duration: "5:15",
      genre: "Deep House",
      bpm: 122,
      key: "Gm",
      progress: 90,
      likes: 2345,
      plays: 7890,
      coverUrl: "https://picsum.photos/202",
      audioUrl: "https://example.com/audio3.mp3",
      tags: ["deep house", "sunset", "chill"],
      createdAt: "2024-03-13"
    }
  ];

  const playlists: Playlist[] = [
    {
      id: 1,
      name: "Mes favoris",
      description: "Mes tracks préférés",
      tracks: [tracks[0], tracks[2]],
      coverUrl: "https://picsum.photos/203",
      createdAt: "2024-03-15"
    },
    {
      id: 2,
      name: "Workout",
      description: "Pour s'entraîner",
      tracks: [tracks[1]],
      coverUrl: "https://picsum.photos/204",
      createdAt: "2024-03-14"
    }
  ];

  const genres = [
    { id: "all", label: "Tous" },
    { id: "House", label: "House" },
    { id: "Techno", label: "Techno" },
    { id: "Deep House", label: "Deep House" },
    { id: "Trance", label: "Trance" }
  ];

  const bpmRanges = [
    { id: "all", label: "Tous" },
    { id: "slow", label: "Lent (60-100)" },
    { id: "medium", label: "Moyen (100-130)" },
    { id: "fast", label: "Rapide (130+)" }
  ];

  const keys = [
    { id: "all", label: "Tous" },
    { id: "Am", label: "La mineur" },
    { id: "Fm", label: "Fa mineur" },
    { id: "Gm", label: "Sol mineur" }
  ];

  const filteredTracks = tracks.filter(track => {
    const matchesSearch = track.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         track.artist.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesGenre = selectedGenre === null || selectedGenre === "all" || track.genre === selectedGenre;
    const matchesBpm = selectedBpm === null || selectedBpm === "all" || 
      (selectedBpm === "slow" && track.bpm < 100) ||
      (selectedBpm === "medium" && track.bpm >= 100 && track.bpm < 130) ||
      (selectedBpm === "fast" && track.bpm >= 130);
    const matchesKey = selectedKey === null || selectedKey === "all" || track.key === selectedKey;
    return matchesSearch && matchesGenre && matchesBpm && matchesKey;
  });

  const togglePlay = (track: Track) => {
    if (isPlaying === track.id) {
      setIsPlaying(null);
      setCurrentTrack(null);
    } else {
      setIsPlaying(track.id);
      setCurrentTrack(track);
    }
  };

  const handleVolumeChange = (value: number) => {
    setVolume(value);
    if (value === 0) {
      setIsMuted(true);
    } else {
      setIsMuted(false);
    }
  };

  const toggleMute = () => {
    setIsMuted(!isMuted);
  };

  const createPlaylist = () => {
    // Logique de création de playlist
    setShowCreatePlaylist(false);
    setNewPlaylistName("");
    setNewPlaylistDescription("");
  };

  return (
    <div className="p-6 space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Tracks</h1>
        <div className="flex gap-2">
          <Dialog open={showCreatePlaylist} onOpenChange={setShowCreatePlaylist}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="h-4 w-4 mr-2" />
                Nouvelle playlist
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Créer une nouvelle playlist</DialogTitle>
              </DialogHeader>
              <div className="space-y-4">
                <div>
                  <label className="text-sm font-medium">Nom</label>
                  <Input
                    value={newPlaylistName}
                    onChange={(e) => setNewPlaylistName(e.target.value)}
                    placeholder="Nom de la playlist"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">Description</label>
                  <Input
                    value={newPlaylistDescription}
                    onChange={(e) => setNewPlaylistDescription(e.target.value)}
                    placeholder="Description de la playlist"
                  />
                </div>
                <Button onClick={createPlaylist}>Créer</Button>
              </div>
            </DialogContent>
          </Dialog>
          <Button>
            <Share2 className="h-4 w-4 mr-2" />
            Partager un track
          </Button>
        </div>
      </div>

      {/* Statistiques */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Tracks</CardTitle>
            <Music className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">156</div>
            <p className="text-xs text-muted-foreground">+12% ce mois</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Écoutes</CardTitle>
            <Play className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">15.9K</div>
            <p className="text-xs text-muted-foreground">+8% cette semaine</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Playlists</CardTitle>
            <ListMusic className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">42</div>
            <p className="text-xs text-muted-foreground">+5% ce mois</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Artistes</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">89</div>
            <p className="text-xs text-muted-foreground">+23% cette semaine</p>
          </CardContent>
        </Card>
      </div>

      {/* Playlists */}
      <div>
        <h2 className="text-xl font-semibold mb-4">Mes playlists</h2>
        <ScrollArea className="h-[200px]">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {playlists.map((playlist) => (
              <Card key={playlist.id} className="hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="aspect-square w-full overflow-hidden rounded-lg">
                    <img
                      src={playlist.coverUrl}
                      alt={playlist.name}
                      className="object-cover w-full h-full"
                    />
                  </div>
                  <div className="mt-4">
                    <CardTitle className="text-lg">{playlist.name}</CardTitle>
                    <p className="text-sm text-muted-foreground">{playlist.description}</p>
                    <p className="text-sm text-muted-foreground mt-2">{playlist.tracks.length} tracks</p>
                  </div>
                </CardHeader>
              </Card>
            ))}
          </div>
        </ScrollArea>
      </div>

      {/* Filtres et recherche */}
      <div className="space-y-4">
        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex-1">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Rechercher des tracks..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9"
              />
            </div>
          </div>
          <div className="flex gap-2">
            <Select value={selectedGenre || "all"} onValueChange={setSelectedGenre}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Genre" />
              </SelectTrigger>
              <SelectContent>
                {genres.map((genre) => (
                  <SelectItem key={genre.id} value={genre.id}>
                    {genre.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={selectedBpm || "all"} onValueChange={setSelectedBpm}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="BPM" />
              </SelectTrigger>
              <SelectContent>
                {bpmRanges.map((range) => (
                  <SelectItem key={range.id} value={range.id}>
                    {range.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={selectedKey || "all"} onValueChange={setSelectedKey}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Tonalité" />
              </SelectTrigger>
              <SelectContent>
                {keys.map((key) => (
                  <SelectItem key={key.id} value={key.id}>
                    {key.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      {/* Grille de tracks */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {filteredTracks.map((track) => (
          <Card key={track.id} className="hover:shadow-lg transition-shadow">
            <CardHeader className="relative">
              <div className="aspect-square w-full overflow-hidden rounded-lg">
                <img
                  src={track.coverUrl}
                  alt={track.title}
                  className="object-cover w-full h-full"
                />
                <Button
                  variant="secondary"
                  size="icon"
                  className="absolute bottom-4 right-4"
                  onClick={() => togglePlay(track)}
                >
                  {isPlaying === track.id ? (
                    <Pause className="h-4 w-4" />
                  ) : (
                    <Play className="h-4 w-4" />
                  )}
                </Button>
              </div>
              <div className="flex items-center justify-between mt-4">
                <div>
                  <CardTitle className="text-lg">{track.title}</CardTitle>
                  <p className="text-sm text-muted-foreground">{track.artist}</p>
                </div>
                <div className="flex gap-2">
                  <Button variant="ghost" size="icon">
                    <Heart className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="icon">
                    <Download className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="icon">
                    <MoreVertical className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center justify-between text-sm">
                  <div className="flex items-center space-x-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <span>{track.duration}</span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <span>{track.bpm} BPM</span>
                    <span>•</span>
                    <span>{track.key}</span>
                  </div>
                </div>
                <Progress value={track.progress} className="h-2" />
                <div className="flex flex-wrap gap-2">
                  {track.tags.map((tag) => (
                    <Badge key={tag} variant="secondary">
                      {tag}
                    </Badge>
                  ))}
                </div>
                <div className="flex items-center justify-between text-sm text-muted-foreground">
                  <div className="flex items-center space-x-4">
                    <div className="flex items-center">
                      <Heart className="h-4 w-4 mr-1" />
                      {track.likes}
                    </div>
                    <div className="flex items-center">
                      <Play className="h-4 w-4 mr-1" />
                      {track.plays}
                    </div>
                  </div>
                  <span>{track.createdAt}</span>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Lecteur audio */}
      {currentTrack && (
        <div className="fixed bottom-0 left-0 right-0 bg-background border-t p-4">
          <div className="container mx-auto flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <img
                src={currentTrack.coverUrl}
                alt={currentTrack.title}
                className="w-12 h-12 rounded"
              />
              <div>
                <p className="font-medium">{currentTrack.title}</p>
                <p className="text-sm text-muted-foreground">{currentTrack.artist}</p>
              </div>
            </div>
            <div className="flex-1 max-w-2xl mx-8">
              <div className="flex items-center space-x-4">
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => togglePlay(currentTrack)}
                >
                  {isPlaying === currentTrack.id ? (
                    <Pause className="h-4 w-4" />
                  ) : (
                    <Play className="h-4 w-4" />
                  )}
                </Button>
                <div className="flex-1">
                  <Progress value={currentTrack.progress} className="h-2" />
                </div>
                <span className="text-sm text-muted-foreground">{currentTrack.duration}</span>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <Button
                variant="ghost"
                size="icon"
                onClick={toggleMute}
              >
                {isMuted ? (
                  <VolumeX className="h-4 w-4" />
                ) : (
                  <Volume2 className="h-4 w-4" />
                )}
              </Button>
              <div className="w-24">
                <Slider
                  value={[isMuted ? 0 : volume]}
                  onValueChange={([value]) => handleVolumeChange(value)}
                  max={100}
                  step={1}
                />
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
} 