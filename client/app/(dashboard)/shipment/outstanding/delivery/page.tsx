'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { Loader2 } from "lucide-react";
import { DataTable } from "@/components/ui/data-table";
import { columns, SuratJalan } from "./columns";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function OutstandingPage() {
    const { isAuthorized } = useAuth();
    const [shipments, setShipments] = useState<SuratJalan[]>([]);
    const [loading, setLoading] = useState(true);

    const fetchShipments = async (authToken: string) => {
        setLoading(true);
        try {
            const res = await fetch(`${API_BASE_URL}/shipments/outstanding/delivery`, {
                headers: { Authorization: `Bearer ${authToken}` },
            });
            const data = await res.json();
            // Menangani berbagai format response (array langsung atau di dalam properti data)
            setShipments(Array.isArray(data) ? data : data?.data || []);
        } catch (error) {
            console.error("Failed to fetch shipments:", error);
            setShipments([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        const storedToken = localStorage.getItem('token');
        if (storedToken && isAuthorized) {
            fetchShipments(storedToken);
        }
    }, [isAuthorized]);

    if (!isAuthorized) return (
        <div className="h-screen flex items-center justify-center">
            <Loader2 className="animate-spin text-blue-600 w-8 h-8" />
        </div>
    );

    return (
        <div className="p-4 space-y-4">
            <div className="flex flex-col gap-1">
                <h1 className="text-2xl font-bold tracking-tight">Outstanding Delivery</h1>
                <p className="text-muted-foreground text-sm">Daftar surat jalan yang sedang dalam proses pengiriman.</p>
            </div>

            <DataTable
                columns={columns(() => { })} // Berikan function kosong karena columns mengharapkan argumen
                data={shipments}
                loading={loading}
            />
        </div>
    );
}