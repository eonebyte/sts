// lib/api.ts
const BASE_URL = 'http://localhost:8080';

async function fetcher(endpoint: string, options: RequestInit = {}) {
    const token = localStorage.getItem('token');
    const headers = {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` }),
        ...options.headers,
    };

    const response = await fetch(`${BASE_URL}${endpoint}`, { ...options, headers });

    if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        throw new Error(error.message || 'Terjadi kesalahan pada server');
    }

    return response.json();
}

export const api = {
    get: (url: string) => fetcher(url, { method: 'GET' }),
    post: (url: string, data: any) => fetcher(url, { method: 'POST', body: JSON.stringify(data) }),
    delete: (url: string) => fetcher(url, { method: 'DELETE' }),
};