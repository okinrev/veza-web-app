interface CacheItem<T> {
  data: T
  timestamp: number
  expiresIn: number
}

class CacheService {
  private cache: Map<string, CacheItem<any>>
  private readonly DEFAULT_EXPIRY = 5 * 60 * 1000 // 5 minutes

  constructor() {
    this.cache = new Map()
  }

  set<T>(key: string, data: T, expiresIn: number = this.DEFAULT_EXPIRY): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      expiresIn,
    })
  }

  get<T>(key: string): T | null {
    const item = this.cache.get(key)
    if (!item) return null

    const isExpired = Date.now() - item.timestamp > item.expiresIn
    if (isExpired) {
      this.cache.delete(key)
      return null
    }

    return item.data
  }

  delete(key: string): void {
    this.cache.delete(key)
  }

  clear(): void {
    this.cache.clear()
  }
}

export const cacheService = new CacheService() 