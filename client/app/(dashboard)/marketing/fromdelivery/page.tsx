'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { Button } from "@/components/ui/button";
import { Plus, Loader2, ClipboardCheck, Send } from "lucide-react"; // Ganti ikon ke ClipboardCheck
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
import { DateRangeFilter } from '@/components/ui/date-range-filter';
import { format } from 'date-fns';
import { id } from "date-fns/locale";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function ReceiptPage() {
    const [token, setToken] = useState<string | null>(null);
    const { isAuthorized, user } = useAuth();
    const [shipments, setShipments] = useState<SuratJalan[]>([]);
    const [loading, setLoading] = useState(true);
    const [isSubmitting, setIsSubmitting] = useState(false);

    const [rowSelection, setRowSelection] = useState({});
    const [isModalOpen, setIsModalOpen] = useState(false);

    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [selectedShipment, setSelectedShipment] = useState<{ id: number, status: string } | null>(null);

    // === Start Date Range ===
    const [dateRange, setDateRange] = useState<{ from: Date | undefined; to: Date | undefined }>({
        from: new Date(new Date().getFullYear(), new Date().getMonth(), 1), // Default: 1st of current month
        to: new Date(new Date().getFullYear(), new Date().getMonth() + 1, 0), // Last day of current month
    });

    const handleResetFilter = () => {
        const currentMonth = new Date();
        const firstDay = new Date(currentMonth.getFullYear(), currentMonth.getMonth(), 1);
        const lastDay = new Date(currentMonth.getFullYear(), currentMonth.getMonth() + 1, 0);
        setDateRange({ from: firstDay, to: lastDay });
    };
    // === End Date Range ===

    const fetchShipments = async (authToken: string, from?: Date, to?: Date) => {
        setLoading(true);
        try {
            const params = new URLSearchParams();
            if (from) {
                params.append('dateFrom', format(from, 'yyyy-MM-dd'));
            }
            if (to) {
                params.append('dateTo', format(to, 'yyyy-MM-dd'));
            }

            const url = `${API_BASE_URL}/shipments/receiptcomebacktomarketing?${params.toString()}`
            // Sesuai endpoint fetching Anda sebelumnya
            const res = await fetch(url, {
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

    // Fetch ulang ketika date range berubah
    useEffect(() => {
        if (token && isAuthorized && (dateRange.from || dateRange.to)) {
            fetchShipments(token, dateRange.from, dateRange.to);
        }
    }, [dateRange]);

    const selectedRowsData = shipments.filter((_, index) =>
        Object.keys(rowSelection).includes(index.toString())
    );

    // Fungsi Submit Receipt (diperbarui ke endpoint /process)
    const handleReceiptConfirm = async () => {
        if (!token || !user?.user_id || selectedRowsData.length === 0) {
            toast.error("Sesi tidak valid, silakan login kembali");
            return;
        }

        setIsSubmitting(true);

        const payload = {
            m_inout_ids: selectedRowsData.map(item => item.m_inout_id),
            status: "RE: MKT_FROM_DEL", // Status baru
            user_id: parseInt(user.user_id),
            notes: "receipt delivery from dpk" // Notes sesuai permintaan
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

            if (!response.ok) throw new Error("Receipt failed");

            toast.success("Receipt Berhasil!", {
                description: `${selectedRowsData.length} Surat Jalan telah diterima.`
            });

            setIsModalOpen(false);
            setRowSelection({});
            fetchShipments(token);

        } catch (error) {
            console.error(error);
            toast.error("Receipt Gagal", {
                description: "Terjadi kesalahan pada server saat memproses penerimaan."
            });
        } finally {
            setIsSubmitting(false);
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

    if (!isAuthorized) return <div className="h-screen flex items-center justify-center"><Loader2 className="animate-spin text-blue-600" /></div>;

    return (
        <div className="space-y-4">
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 px-1">
                <div>
                    <p className="text-sm text-slate-500 mt-1">
                        {dateRange.from && dateRange.to && (
                            <>
                                Periode: {format(dateRange.from, "dd MMM yyyy", { locale: id })} - {format(dateRange.to, "dd MMM yyyy", { locale: id })}
                            </>
                        )}
                    </p>
                </div>

                <DateRangeFilter
                    dateRange={dateRange}
                    setDateRange={setDateRange}
                    handleResetFilter={handleResetFilter}
                />
            </div>

            <DataTable
                columns={columns(onCancelClick)}
                data={shipments}
                rowSelection={rowSelection}
                setRowSelection={setRowSelection}
                loading={loading}
            />

            <div className="sticky bottom-4 flex items-center justify-between p-4 bg-white border shadow-lg rounded-xl transition-all">
                <div className="flex flex-col">
                    <span className="text-sm font-medium text-slate-900">
                        {selectedRowsData.length} Surat Jalan Terpilih
                    </span>
                    <span className="text-xs text-slate-500 text-nowrap">Siap untuk diproses receipt</span>
                </div>

                <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
                    <DialogTrigger asChild>
                        <Button
                            disabled={selectedRowsData.length === 0 || loading}
                            className="bg-indigo-600 hover:bg-indigo-700 shadow-md"
                        >
                            <ClipboardCheck className="w-4 h-4 mr-2" /> Receipt
                        </Button>
                    </DialogTrigger>
                    <DialogContent className="sm:max-w-[450px]">
                        <DialogHeader>
                            <DialogTitle>Konfirmasi Penerimaan (Receipt)</DialogTitle>
                            <DialogDescription>
                                Anda akan memproses penerimaan untuk {selectedRowsData.length} dokumen.
                            </DialogDescription>
                        </DialogHeader>

                        <div className="mt-4 border rounded-lg divide-y max-h-[250px] overflow-y-auto">
                            {selectedRowsData.map((item) => (
                                <div key={item.m_inout_id} className="p-3 flex justify-between items-center text-sm">
                                    <span className="font-bold text-indigo-600">{item.document_no}</span>
                                    <span className="text-slate-500 italic">{item.customer_name}</span>
                                </div>
                            ))}
                        </div>

                        <DialogFooter className="mt-6 gap-2 sm:gap-0">
                            <Button variant="ghost" onClick={() => setIsModalOpen(false)} disabled={isSubmitting}>
                                Batal
                            </Button>
                            <Button
                                onClick={handleReceiptConfirm}
                                className="bg-indigo-600 hover:bg-indigo-700 min-w-[120px]"
                                disabled={isSubmitting}
                            >
                                {isSubmitting ? (
                                    <>
                                        <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                                        Memproses...
                                    </>
                                ) : (
                                    <>
                                        <Send className="w-4 h-4 mr-2" /> Confirm OK
                                    </>
                                )}
                            </Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>

            {/* Shadcn UI AlertDialog */}
            <AlertDialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>Konfirmasi Reject</AlertDialogTitle>
                        <AlertDialogDescription>
                            Apakah Anda yakin ingin reject SJ ini? Tindakan ini akan mengembalikan SJ Ke Delivery.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>Batal</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={handleConfirmCancel}
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        >
                            Ya, Reject
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
}