'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { Loader2 } from "lucide-react";
import { DataTable } from "@/components/ui/data-table";
import { columns, SuratJalan } from "./columns";
import { toast } from "sonner"; // Opsi: tambahkan toast untuk feedback
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function OutstandingPage() {
    const { isAuthorized } = useAuth();
    const [shipments, setShipments] = useState<SuratJalan[]>([]);
    const [loading, setLoading] = useState(true);

    // State untuk AlertDialog
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [selectedShipment, setSelectedShipment] = useState<{ id: number, status: string } | null>(null);

    const fetchShipments = async (authToken: string) => {
        setLoading(true);
        try {
            const res = await fetch(`${API_BASE_URL}/shipments/outstanding/delivery`, {
                headers: { Authorization: `Bearer ${authToken}` },
            });
            const data = await res.json();
            setShipments(Array.isArray(data) ? data : data?.data || []);
        } catch (error) {
            toast.error("Gagal mengambil data pengiriman"); // Toast error fetch
            setShipments([]);
        } finally {
            setLoading(false);
        }
    };

    // Fungsi trigger dialog (dipanggil dari button di table)
    const onCancelClick = (m_inout_id: number, status: string) => {
        setSelectedShipment({ id: m_inout_id, status });
        setIsDialogOpen(true);
    };

    // Fungsi eksekusi pembatalan setelah dikonfirmasi di dialog
    const handleConfirmCancel = async () => {
        if (!selectedShipment) return;

        const storedToken = localStorage.getItem('token');
        const { id: m_inout_id, status } = selectedShipment;

        const promise = async () => {
            const res = await fetch(`${API_BASE_URL}/shipments/outstanding/cancel`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${storedToken}`
                },
                body: JSON.stringify({ m_inout_id, status })
            });

            if (!res.ok) {
                const errorData = await res.json();
                throw new Error(errorData.message || "Gagal membatalkan");
            }

            if (storedToken) fetchShipments(storedToken);
            return res.json();
        };

        toast.promise(promise(), {
            loading: 'Sedang memproses pembatalan...',
            success: 'Pengiriman berhasil dibatalkan',
            error: (err) => `Gagal: ${err.message}`,
        });

        setIsDialogOpen(false); // Tutup dialog setelah aksi dimulai
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
                columns={columns(onCancelClick)} // Kirim fungsi handleCancel ke columns
                data={shipments}
                loading={loading}
            />

            {/* Shadcn UI AlertDialog */}
            <AlertDialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>Konfirmasi Pembatalan</AlertDialogTitle>
                        <AlertDialogDescription>
                            Apakah Anda yakin ingin membatalkan pengiriman ini? Tindakan ini akan mengembalikan status dokumen ke tahap sebelumnya.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>Batal</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={handleConfirmCancel}
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        >
                            Ya, Batalkan
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
}