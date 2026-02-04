'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { Button } from "@/components/ui/button";
import { Loader2, PackageCheck, Send, AlertCircle } from "lucide-react";
import { DataTable } from "@/components/ui/data-table";
import { columns, SuratJalan } from "./columns";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { toast } from 'sonner';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function Page() {
    const [token, setToken] = useState<string | null>(null);
    const { isAuthorized, user } = useAuth();
    const [shipments, setShipments] = useState<SuratJalan[]>([]);
    const [loading, setLoading] = useState(true);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [rowSelection, setRowSelection] = useState<Record<string, boolean>>({});
    const [isModalOpen, setIsModalOpen] = useState(false);

    const fetchShipments = async (authToken: string) => {
        setLoading(true);
        try {
            const res = await fetch(`${API_BASE_URL}/shipments/comebacktodelivery`, {
                headers: { Authorization: `Bearer ${authToken}` },
            });
            const data = await res.json();
            const rows = Array.isArray(data) ? data : data?.data || [];
            setShipments(rows.map((s: any) => ({
                ...s,
                id: s.m_inout_id,
                customer_name: s.customer_name ?? "-",
            })));
        } catch (error) {
            setShipments([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        const storedToken = localStorage.getItem('token');
        setToken(storedToken);
        if (storedToken && isAuthorized) {
            fetchShipments(storedToken);
        }
    }, [isAuthorized]);

    const selectedRowsData = shipments.filter((_, index) => rowSelection[index]);

    const handleHandoverConfirm = async () => {
        if (selectedRowsData.length === 0 || !token) return;

        // Ambil data driver dan tnkb dari baris pertama yang dipilih

        setIsSubmitting(true);
        const payload = {
            m_inout_ids: selectedRowsData.map(item => item.m_inout_id),
            status: "HO: DPK_TO_DEL", // Sesuaikan statusnya
            user_id: parseInt(user.user_id),
            notes: 'handover from dpk to delivery'
        };

        try {
            const response = await fetch(`${API_BASE_URL}/handover/process`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(payload)
            });

            if (!response.ok) throw new Error();

            toast.success("Handover ke Delivery Berhasil!");
            setIsModalOpen(false);
            setRowSelection({});
            fetchShipments(token);
        } catch (error) {
            toast.error("Gagal memproses handover");
        } finally {
            setIsSubmitting(false);
        }
    };

    if (!isAuthorized) return <div className="h-screen flex items-center justify-center"><Loader2 className="animate-spin text-blue-600" /></div>;

    return (
        <div className="space-y-4">
            <DataTable
                columns={columns}
                data={shipments}
                rowSelection={rowSelection}
                setRowSelection={setRowSelection}
                loading={loading}
            />

            <div className="sticky bottom-4 flex items-center justify-between p-4 bg-white border shadow-lg rounded-xl">
                <div className="flex flex-col">
                    <span className="text-sm font-bold text-blue-600">{selectedRowsData.length} Surat Jalan Terpilih</span>
                    <span className="text-xs text-slate-500">Pastikan data driver sudah sesuai di tabel.</span>
                </div>

                <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
                    <DialogTrigger asChild>
                        <Button disabled={selectedRowsData.length === 0} className="bg-blue-600 hover:bg-blue-700">
                            <PackageCheck className="w-4 h-4 mr-2" /> Proses Handover
                        </Button>
                    </DialogTrigger>
                    <DialogContent className="sm:max-w-[400px]">
                        <DialogHeader>
                            <DialogTitle>Konfirmasi Handover</DialogTitle>
                            <DialogDescription>
                                Anda akan memproses handover untuk {selectedRowsData.length} surat jalan ke bagian Delivery.
                            </DialogDescription>
                        </DialogHeader>

                        <div className="bg-blue-50 p-4 rounded-lg border border-blue-100 flex gap-3">
                            <AlertCircle className="w-5 h-5 text-blue-600 shrink-0" />
                            <div className="text-sm text-blue-800">
                                <p className="font-semibold">Informasi Sistem:</p>
                                <p>Data Driver dan Plat Nomor akan mengikuti data yang tertera pada dokumen Surat Jalan.</p>
                            </div>
                        </div>

                        <DialogFooter className="mt-4">
                            <Button variant="ghost" onClick={() => setIsModalOpen(false)}>Batal</Button>
                            <Button
                                onClick={handleHandoverConfirm}
                                className="bg-blue-600 hover:bg-blue-700"
                                disabled={isSubmitting}
                            >
                                {isSubmitting ? <Loader2 className="animate-spin w-4 h-4 mr-2" /> : <Send className="w-4 h-4 mr-2" />}
                                Konfirmasi & Kirim
                            </Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>
        </div>
    );
}