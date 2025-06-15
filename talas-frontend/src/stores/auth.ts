import { create } from "zustand"
import { persist } from "zustand/middleware"
import axios from "axios"

interface User {
  id: number
  username: string
  email: string
  avatar?: string
}

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  register: (username: string, email: string, password: string) => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,

      login: async (email: string, password: string) => {
        try {
          const response = await axios.post("/api/auth/login", {
            email,
            password,
          })

          const { access_token, user } = response.data

          set({
            user,
            token: access_token,
            isAuthenticated: true,
          })

          // Configure axios default headers
          axios.defaults.headers.common["Authorization"] = `Bearer ${access_token}`
        } catch (error) {
          throw new Error("Échec de la connexion")
        }
      },

      register: async (username: string, email: string, password: string) => {
        try {
          const response = await axios.post("/api/auth/register", {
            username,
            email,
            password,
          })

          const { access_token, user } = response.data

          set({
            user,
            token: access_token,
            isAuthenticated: true,
          })

          // Configure axios default headers
          axios.defaults.headers.common["Authorization"] = `Bearer ${access_token}`
        } catch (error) {
          throw new Error("Échec de l'inscription")
        }
      },

      logout: () => {
        set({
          user: null,
          token: null,
          isAuthenticated: false,
        })

        // Remove axios default headers
        delete axios.defaults.headers.common["Authorization"]
      },
    }),
    {
      name: "auth-storage",
    }
  )
) 