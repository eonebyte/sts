// src/hooks/useAuth.ts
import { useEffect, useState } from 'react';
import { usePathname, useRouter } from 'next/navigation';

export function useAuth() {
    const [isAuthorized, setIsAuthorized] = useState(false);
    const [user, setUser] = useState<any>(null); // Tambahkan state user
    const router = useRouter();
    const pathname = usePathname();

    useEffect(() => {
        const verifyToken = async () => {
            const token = localStorage.getItem('token');

            if (!token) {
                router.push('/login');
                return;
            }

            try {
                const res = await fetch('http://localhost:8080/me', {
                    method: 'GET',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json',
                    },
                });

                const result = await res.json();

                if (res.ok) {
                    const userData = result.data;
                    setIsAuthorized(true);
                    setUser(userData);

                    if (pathname === '/' || pathname === '/login') {
                        if (userData.title === 'driver') {
                            router.push('/driver');
                        } else {
                            router.push('/');
                        }
                    }

                } else {
                    throw new Error("Token Invalid");
                }
            } catch (error) {
                console.error("Auth Error:", error);
                localStorage.removeItem('token');
                router.push('/login');
            }
        };

        verifyToken();
    }, [pathname, router]);

    // Sekarang return object yang berisi status dan data user
    return { isAuthorized, user };
}

export function logout() {
    localStorage.removeItem('token');
    window.location.href = '/login';
}