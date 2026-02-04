    'use client';

    import { useState, useEffect } from 'react';
    import Link from 'next/link';
    import { useAuth } from '@/hooks/useAuth';
    import { Button } from "@/components/ui/button";
    import { Plus, Loader2, PackageCheck, Send } from "lucide-react";
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
        const [isSubmitting, setIsSubmitting] = useState(false); // State untuk loading saat submit

        const [rowSelection, setRowSelection] = useState({});
        const [isModalOpen, setIsModalOpen] = useState(false);

        // Fungsi untuk fetch data (dipisah agar bisa dipanggil ulang setelah submit)
        const fetchShipments = async (authToken: string) => {
            setLoading(true);
            try {
                const res = await fetch(`${API_BASE_URL}/shipments/pending`, {
                    headers: { Authorization: `Bearer ${authToken}` },
                });
                if (!res.ok) throw new Error("Failed to fetch");
                const data = await res.json();
                const rows = Array.isArray(data) ? data : data?.data || [];
                setShipments(rows.map((s: any) => ({
                    ...s,
                    id: s.m_inout_id,
                    customer_name: s.customer_name ?? "-",
                    driver_name: s.driver_name ?? "-",
                    tnkb_no: s.tnkb_no ?? "-",
                })));
            } catch (error) {
                console.error(error);
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

        // Ambil data yang sedang dicentang
        const selectedRowsData = shipments.filter((_, index) =>
            Object.keys(rowSelection).includes(index.toString())
        );

        // Fungsi Submit Handover
        const handleHandoverConfirm = async () => {
            if (!token || !user?.user_id || selectedRowsData.length === 0) {
                toast.error("Sesi tidak valid, silakan login kembali");
                return;
            }

            setIsSubmitting(true);

            // Persiapkan Payload sesuai permintaan
            const payload = {
                m_inout_ids: selectedRowsData.map(item => item.m_inout_id),
                status: "HO: DEL_TO_DPK",
                user_id: parseInt(user.user_id)
            };

            try {
                const response = await fetch(`${API_BASE_URL}/handover/init`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${token}`
                    },
                    body: JSON.stringify(payload)
                });

                if (!response.ok) throw new Error("Handover failed");

                toast.success("Handover Berhasil!", {
                    description: `${selectedRowsData.length} Surat Jalan telah diproses.`
                });
                setIsModalOpen(false);
                setRowSelection({}); // Reset checkbox
                fetchShipments(token); // Refresh data tabel (agar yang sudah HO hilang dari 'pending')

            } catch (error) {
                console.error(error);
                toast.error("Handover Gagal", {
                    description: "Terjadi kesalahan pada server saat memproses data."
                });
            } finally {
                setIsSubmitting(false);
            }
        };

        if (!isAuthorized) return <div className="h-screen flex items-center justify-center"><Loader2 className="animate-spin text-blue-600" /></div>;

        return (
            <div className="space-y-4">
                {/* <div className="flex justify-between items-center px-1">
                    <h1 className="text-xl font-bold text-slate-800">Surat Jalan Pending</h1>
                </div> */}

                <DataTable
                    columns={columns}
                    data={shipments}
                    rowSelection={rowSelection}
                    setRowSelection={setRowSelection}
                    loading={loading}
                // Jika DataTable anda memakai loading internal:
                // loading={loading} 
                />

                {/* Footer Action Bar */}
                <div className="sticky bottom-4 flex items-center justify-between p-4 bg-white border shadow-lg rounded-xl transition-all">
                    <div className="flex flex-col">
                        <span className="text-sm font-medium text-slate-900">
                            {selectedRowsData.length} Surat Jalan Terpilih
                        </span>
                        <span className="text-xs text-slate-500 text-nowrap">Siap untuk diproses handover</span>
                    </div>

                    <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
                        <DialogTrigger asChild>
                            <Button
                                disabled={selectedRowsData.length === 0 || loading}
                                className="bg-green-600 hover:bg-green-700 shadow-md"
                            >
                                <PackageCheck className="w-4 h-4 mr-2" /> Handover
                            </Button>
                        </DialogTrigger>
                        <DialogContent className="sm:max-w-[450px]">
                            <DialogHeader>
                                <DialogTitle>Konfirmasi Handover</DialogTitle>
                                <DialogDescription>
                                    Anda akan mengirim {selectedRowsData.length} dokumen ke <strong>DPK</strong>. Tindakan ini tidak dapat dibatalkan.
                                </DialogDescription>
                            </DialogHeader>

                            <div className="mt-4 border rounded-lg divide-y max-h-[250px] overflow-y-auto">
                                {selectedRowsData.map((item) => (
                                    <div key={item.m_inout_id} className="p-3 flex justify-between items-center text-sm">
                                        <span className="font-bold text-blue-600">{item.document_no}</span>
                                        <span className="text-slate-500 italic">{item.customer_name}</span>
                                    </div>
                                ))}
                            </div>

                            <DialogFooter className="mt-6 gap-2 sm:gap-0">
                                <Button
                                    variant="ghost"
                                    onClick={() => setIsModalOpen(false)}
                                    disabled={isSubmitting}
                                >
                                    Batal
                                </Button>
                                <Button
                                    onClick={handleHandoverConfirm}
                                    className="bg-green-600 hover:bg-green-700 min-w-[120px]"
                                    disabled={isSubmitting}
                                >
                                    {isSubmitting ? (
                                        <>
                                            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                                            Memproses...
                                        </>
                                    ) : (
                                        <>
                                            <Send className="w-4 h-4 mr-2" /> Submit OK
                                        </>
                                    )}
                                </Button>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                </div>
            </div>
        );
    }