'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Lock } from "lucide-react";
import Image from 'next/image';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function LoginPage() {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const router = useRouter();

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);

        try {
            const res = await fetch(`${API_BASE_URL}/auth/login`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    username,
                    password,
                }),
            });

            const resBody = await res.json();
            const { success, message, data } = resBody;

            if (!res.ok || !success) {
                throw new Error(message || 'Username atau password salah');
            }

            localStorage.setItem('token', data.access_token);
            localStorage.setItem('refresh_token', data.refresh_token);

            router.push('/');
        } catch (err) {
            console.log(err);

            alert("Gagal koneksi ke server");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-slate-100 p-4">
            <Card className="w-full max-w-md shadow-xl">
                <CardHeader className="text-center">
                    <div className="mx-auto p-0 rounded-full w-fit mb-2">
                        <Image
                            src="/sts.png"
                            alt="Driver Portal Logo"
                            width={80}
                            height={80}
                            className="rounded-lg object-cover"
                            priority
                        />
                    </div>
                    {/* <CardTitle className="text-2xl font-bold text-slate-900">STS</CardTitle> */}
                    <CardDescription>Masukkan kredensial Anda untuk mengakses sistem.</CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleLogin} className="space-y-4">
                        <div className="space-y-2">
                            <Label>Username</Label>
                            <Input type="text" value={username} onChange={(e) => setUsername(e.target.value)} required />
                        </div>
                        <div className="space-y-2">
                            <Label>Password</Label>
                            <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
                        </div>
                        <Button type="submit" className="w-full bg-blue-700 hover:bg-blue-800" disabled={loading}>
                            {loading ? 'Memproses...' : 'Login'}
                        </Button>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}