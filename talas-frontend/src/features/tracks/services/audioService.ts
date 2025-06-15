import { apiClient } from '@/shared/services/apiClient';
import { Track, ApiResponse, TracksResponse } from '../types';

export const audioService = {
  async getTracks(params?: { search?: string; artist?: string; tags?: string[] }): Promise<TracksResponse> {
    try {
      console.log('Fetching tracks with params:', params);
      const response = await apiClient.get<ApiResponse<Track[]>>('/tracks', { 
        params: {
          ...params,
          page: 1,
          limit: 100 // Augmenter la limite pour récupérer plus de pistes
        }
      });
      console.log('Raw API Response:', response);
      console.log('Tracks data:', response.data);
      
      if (!response.data.success) {
        throw new Error(response.data.message || 'Erreur lors de la récupération des pistes');
      }

      // Transformer la réponse pour correspondre à l'interface TracksResponse
      const tracksResponse: TracksResponse = {
        tracks: Array.isArray(response.data.data) ? response.data.data : [response.data.data],
        total: Array.isArray(response.data.data) ? response.data.data.length : 1,
        page: 1,
        limit: 100
      };

      console.log('Transformed tracks response:', tracksResponse);
      return tracksResponse;
    } catch (error) {
      console.error('Error fetching tracks:', error);
      throw error;
    }
  },

  async getTrack(id: number) {
    const response = await apiClient.get<ApiResponse<Track>>(`/tracks/${id}`);
    return response.data.data;
  },

  async uploadTrack(file: File, metadata: Partial<Track>, onProgress?: (progress: number) => void) {
    const formData = new FormData();
    formData.append('file', file);
    Object.entries(metadata).forEach(([key, value]) => {
      formData.append(key, value.toString());
    });

    const response = await apiClient.post<ApiResponse<Track>>('/tracks', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (progressEvent.total) {
          const progress = (progressEvent.loaded / progressEvent.total) * 100;
          onProgress?.(progress);
        }
      },
    });

    return response.data.data;
  },

  async updateTrack(id: number, metadata: Partial<Track>) {
    const response = await apiClient.put<ApiResponse<Track>>(`/tracks/${id}`, metadata);
    return response.data.data;
  },

  async deleteTrack(id: number) {
    await apiClient.delete(`/tracks/${id}`);
  },

  async getStreamUrl(id: number): Promise<{ url: string }> {
    console.log('Getting stream URL for track:', id);
    try {
      // Récupérer les détails de la piste
      const response = await apiClient.get<ApiResponse<Track>>(`/tracks/${id}`);
      console.log('Received track details:', response);
      
      if (!response.data.success) {
        throw new Error(response.data.message || 'Erreur lors de la récupération des détails de la piste');
      }
      
      const track = response.data.data;
      if (!track.filename) {
        throw new Error('Nom de fichier non disponible');
      }
      
      // Construire l'URL du fichier en utilisant l'endpoint de streaming
      const fileUrl = `${apiClient.defaults.baseURL}/tracks/${id}/stream`;
      console.log('Constructed file URL:', fileUrl);
      
      return { url: fileUrl };
    } catch (error) {
      console.error('Error getting stream URL:', error);
      throw error;
    }
  },

  async incrementPlayCount(id: number) {
    await apiClient.post(`/tracks/${id}/play`);
  }
}; 