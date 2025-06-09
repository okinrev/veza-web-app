import axios from 'axios';

const API_URL = '/api';

export const login = async (email, password) => {
  try {
    const response = await axios.post(`${API_URL}/auth/login`, {
      email,
      password
    });

    if (response.data.token) {
      localStorage.setItem('user', JSON.stringify(response.data));
    }

    return response.data;
  } catch (error) {
    throw error.response?.data || { message: 'Une erreur est survenue lors de la connexion' };
  }
};

export const register = async (username, email, password) => {
  try {
    const response = await axios.post(`${API_URL}/auth/register`, {
      username,
      email,
      password
    });

    return response.data;
  } catch (error) {
    throw error.response?.data || { message: 'Une erreur est survenue lors de l\'inscription' };
  }
};

export const forgotPassword = async (email) => {
  try {
    const response = await axios.post(`${API_URL}/auth/forgot-password`, {
      email
    });

    return response.data;
  } catch (error) {
    throw error.response?.data || { message: 'Une erreur est survenue lors de la demande de réinitialisation' };
  }
};

export const resetPassword = async (token, password) => {
  try {
    const response = await axios.post(`${API_URL}/auth/reset-password`, {
      token,
      password
    });

    return response.data;
  } catch (error) {
    throw error.response?.data || { message: 'Une erreur est survenue lors de la réinitialisation du mot de passe' };
  }
};

export const logout = () => {
  localStorage.removeItem('user');
};

export const getCurrentUser = () => {
  return JSON.parse(localStorage.getItem('user'));
};

export const isAuthenticated = () => {
  const user = getCurrentUser();
  return !!user?.token;
};

export const getAuthHeader = () => {
  const user = getCurrentUser();
  if (user && user.token) {
    return { Authorization: `Bearer ${user.token}` };
  }
  return {};
}; 