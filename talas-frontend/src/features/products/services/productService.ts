import { productCache, CACHE_KEYS } from './productCache';
import { apiClient } from '@/lib/api';

export interface Product {
  id: number;
  name: string;
  description: string;
  price: number;
  image: string;
  category: string;
  stock: number;
}

export interface CreateProductData {
  name: string;
  description: string;
  price: number;
  image: string;
  category: string;
  stock: number;
}

export interface UpdateProductData extends Partial<CreateProductData> {
  id: number;
}

class ProductService {
  private static instance: ProductService;
  private readonly baseUrl = '/api/products';

  private constructor() {}

  public static getInstance(): ProductService {
    if (!ProductService.instance) {
      ProductService.instance = new ProductService();
    }
    return ProductService.instance;
  }

  async getProducts(): Promise<Product[]> {
    try {
      // Vérifier le cache
      const cachedProducts = productCache.get<Product[]>(CACHE_KEYS.PRODUCTS);
      if (cachedProducts) {
        return cachedProducts;
      }

      // Récupérer les données de l'API
      const response = await apiClient.get<Product[]>(this.baseUrl);
      const products = response.data;

      // Mettre en cache
      productCache.set(CACHE_KEYS.PRODUCTS, products);

      return products;
    } catch (error) {
      console.error('Erreur lors de la récupération des produits:', error);
      throw error;
    }
  }

  async getProduct(id: number): Promise<Product> {
    try {
      // Vérifier le cache
      const cachedProduct = productCache.get<Product>(`${CACHE_KEYS.PRODUCT}_${id}`);
      if (cachedProduct) {
        return cachedProduct;
      }

      // Récupérer les données de l'API
      const response = await apiClient.get<Product>(`${this.baseUrl}/${id}`);
      const product = response.data;

      // Mettre en cache
      productCache.set(`${CACHE_KEYS.PRODUCT}_${id}`, product);

      return product;
    } catch (error) {
      console.error(`Erreur lors de la récupération du produit ${id}:`, error);
      throw error;
    }
  }

  async createProduct(data: CreateProductData): Promise<Product> {
    try {
      const response = await apiClient.post<Product>(this.baseUrl, data);
      const newProduct = response.data;

      // Invalider le cache des produits
      productCache.delete(CACHE_KEYS.PRODUCTS);

      return newProduct;
    } catch (error) {
      console.error('Erreur lors de la création du produit:', error);
      throw error;
    }
  }

  async updateProduct(data: UpdateProductData): Promise<Product> {
    try {
      const response = await apiClient.put<Product>(`${this.baseUrl}/${data.id}`, data);
      const updatedProduct = response.data;

      // Invalider les caches
      productCache.delete(CACHE_KEYS.PRODUCTS);
      productCache.delete(`${CACHE_KEYS.PRODUCT}_${data.id}`);

      return updatedProduct;
    } catch (error) {
      console.error(`Erreur lors de la mise à jour du produit ${data.id}:`, error);
      throw error;
    }
  }

  async deleteProduct(id: number): Promise<void> {
    try {
      await apiClient.delete(`${this.baseUrl}/${id}`);

      // Invalider les caches
      productCache.delete(CACHE_KEYS.PRODUCTS);
      productCache.delete(`${CACHE_KEYS.PRODUCT}_${id}`);
    } catch (error) {
      console.error(`Erreur lors de la suppression du produit ${id}:`, error);
      throw error;
    }
  }

  async getCategories(): Promise<string[]> {
    try {
      // Vérifier le cache
      const cachedCategories = productCache.get<string[]>(CACHE_KEYS.CATEGORIES);
      if (cachedCategories) {
        return cachedCategories;
      }

      // Récupérer les données de l'API
      const response = await apiClient.get<string[]>(`${this.baseUrl}/categories`);
      const categories = response.data;

      // Mettre en cache
      productCache.set(CACHE_KEYS.CATEGORIES, categories);

      return categories;
    } catch (error) {
      console.error('Erreur lors de la récupération des catégories:', error);
      throw error;
    }
  }
}

export const productService = ProductService.getInstance(); 