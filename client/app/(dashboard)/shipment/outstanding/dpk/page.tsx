'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { Car, Check, ChevronsUpDown, Loader2, Send, User } from "lucide-react";
import { DataTable } from "@/components/ui/data-table";
import { columns, SuratJalan } from "./columns";
import { toast } from "sonner"; // Opsi: tambahkan toast untuk feedback
import { id } from "date-fns/locale";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
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
import {
    Command,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
} from "@/components/ui/command";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";
import { Button } from '@/components/ui/button';
import { format } from 'date-fns';
import { DateRangeFilter } from '@/components/ui/date-range-filter';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface MasterDataDriver { driver_by: string | number; Name: string; }
interface MasterDataTnkb { tnkb_id: string | number; Name: string; }

export default function OutstandingPage() {
    const [token, setToken] = useState<string | null>(null);
    const { isAuthorized, user } = useAuth();
    const [shipments, setShipments] = useState<SuratJalan[]>([]);
    const [loading, setLoading] = useState(true);

    // State untuk AlertDialog
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [selectedShipment, setSelectedShipment] = useState<{ id: number, status: string } | null>(null);

    const [drivers, setDrivers] = useState<MasterDataDriver[]>([]);
    const [tnkbs, setTnkbs] = useState<MasterDataTnkb[]>([]);

    const [selectedSJ, setSelectedSJ] = useState<SuratJalan | null>(null);
    const [selectedDriver, setSelectedDriver] = useState<string>("");
    const [selectedTnkb, setSelectedTnkb] = useState<string>("");

    const [isSubmitting, setIsSubmitting] = useState(false);


    const [openDriver, setOpenDriver] = useState(false);
    const [openTnkb, setOpenTnkb] = useState(false);
    const [isModalOpen, setIsModalOpen] = useState(false);

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

            const res = await fetch(`${API_BASE_URL}/shipments/outstanding/dpk?${params.toString()}`, {
                headers: { Authorization: `Bearer ${authToken}` },
            });
            const data = await res.json();
            setShipments(Array.isArray(data) ? data : data?.data || []);
        } catch (error) {
            console.error("Failed to fetch shipments:", error);
            toast.error("Gagal memuat daftar pengiriman");
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

    const fetchMasterData = async (authToken: string) => {
        try {
            const [resD, resT] = await Promise.all([
                fetch(`${API_BASE_URL}/shipments/drivers`, { headers: { Authorization: `Bearer ${authToken}` } }),
                fetch(`${API_BASE_URL}/shipments/tnkbs`, { headers: { Authorization: `Bearer ${authToken}` } })
            ]);
            const d = await resD.json();
            const t = await resT.json();
            setDrivers(d.data || []);
            setTnkbs(t.data || []);
        } catch (error) { console.error(error); }
    };

    useEffect(() => {
        const storedToken = localStorage.getItem('token');
        if (storedToken && isAuthorized) {
            setToken(storedToken);
            fetchShipments(storedToken);
            fetchMasterData(storedToken);
        }
    }, [isAuthorized]);

    // Fetch ulang ketika date range berubah
    useEffect(() => {
        if (token && isAuthorized && (dateRange.from || dateRange.to)) {
            fetchShipments(token, dateRange.from, dateRange.to);
        }
    }, [dateRange]);

    const handleOpenEdit = (sj: SuratJalan) => {
        setSelectedSJ(sj);
        // Pre-fill data lama jika ada kecocokan di master data
        const d = drivers.find(drv => drv.Name === sj.driver_name);
        const t = tnkbs.find(v => v.Name === sj.tnkb_no);
        setSelectedDriver(d ? String(d.driver_by) : "");
        setSelectedTnkb(t ? String(t.tnkb_id) : "");
        setIsModalOpen(true);
    };

    const handleEditSubmit = async () => {
        if (!selectedSJ || !token) return;

        setIsSubmitting(true);
        try {
            const response = await fetch(`${API_BASE_URL}/shipments/edit/drivertnkb`, {
                method: 'POST', // atau 'PUT' sesuai spek backend Anda
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token}`
                },
                body: JSON.stringify({
                    m_inout_id: selectedSJ.m_inout_id,
                    driver_by: selectedDriver ? parseInt(selectedDriver) : null,
                    tnkb_id: selectedTnkb ? parseInt(selectedTnkb) : null,
                    updated_by: user?.user_id // Opsional: mengirim siapa yang mengedit
                })
            });

            if (!response.ok) throw new Error("Gagal mengupdate data");

            toast.success("Data Surat Jalan berhasil diperbarui");
            setIsModalOpen(false);
            fetchShipments(token); // Refresh table
        } catch (error) {
            toast.error("Terjadi kesalahan saat mengupdate data");
            console.error(error);
        } finally {
            setIsSubmitting(false);
        }
    };

    if (!isAuthorized) return (
        <div className="h-screen flex items-center justify-center">
            <Loader2 className="animate-spin text-blue-600 w-8 h-8" />
        </div>
    );

    return (
        <div className="p-4 space-y-4">


            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 px-1">
                <div>
                    <p className="text-xl">
                        Outstanding
                    </p>
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
                columns={columns(onCancelClick, handleOpenEdit)} // Kirim fungsi handleCancel ke columns
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

            <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
                <DialogContent className="sm:max-w-[450px]">
                    <DialogHeader>
                        <DialogTitle>Edit Driver & TNKB</DialogTitle>
                        <DialogDescription>
                            Edit data untuk: <span className="font-bold text-slate-900">{selectedSJ?.document_no}</span>
                        </DialogDescription>
                    </DialogHeader>

                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <label className="text-sm font-semibold flex items-center gap-2"><User className="w-4 h-4" /> Driver</label>
                            <Popover open={openDriver} onOpenChange={setOpenDriver}>
                                <PopoverTrigger asChild>
                                    <Button variant="outline" className="w-full justify-between">
                                        {selectedDriver ? drivers.find(d => String(d.driver_by) === selectedDriver)?.Name : "Pilih Driver"}
                                        <ChevronsUpDown className="ml-2 h-4 w-4 opacity-50" />
                                    </Button>
                                </PopoverTrigger>
                                <PopoverContent className="w-[400px] p-0">
                                    <Command>
                                        <CommandInput placeholder="Cari driver..." />
                                        <CommandList>
                                            <CommandEmpty>Tidak ditemukan.</CommandEmpty>
                                            <CommandGroup>
                                                {drivers.map((d) => (
                                                    <CommandItem key={d.driver_by} onSelect={() => { setSelectedDriver(String(d.driver_by)); setOpenDriver(false); }}>
                                                        <Check className={cn("mr-2 h-4 w-4", selectedDriver === String(d.driver_by) ? "opacity-100" : "opacity-0")} />
                                                        {d.Name}
                                                    </CommandItem>
                                                ))}
                                            </CommandGroup>
                                        </CommandList>
                                    </Command>
                                </PopoverContent>
                            </Popover>
                        </div>

                        <div className="space-y-2">
                            <label className="text-sm font-semibold flex items-center gap-2"><Car className="w-4 h-4" /> Plat Nomor (TNKB)</label>
                            <Popover open={openTnkb} onOpenChange={setOpenTnkb}>
                                <PopoverTrigger asChild>
                                    <Button variant="outline" className="w-full justify-between">
                                        {selectedTnkb ? tnkbs.find(t => String(t.tnkb_id) === selectedTnkb)?.Name : "Pilih Plat Nomor"}
                                        <ChevronsUpDown className="ml-2 h-4 w-4 opacity-50" />
                                    </Button>
                                </PopoverTrigger>
                                <PopoverContent className="w-[400px] p-0">
                                    <Command>
                                        <CommandInput placeholder="Cari plat..." />
                                        <CommandList>
                                            <CommandEmpty>Tidak ditemukan.</CommandEmpty>
                                            <CommandGroup>
                                                {tnkbs.map((t) => (
                                                    <CommandItem key={t.tnkb_id} onSelect={() => { setSelectedTnkb(String(t.tnkb_id)); setOpenTnkb(false); }}>
                                                        <Check className={cn("mr-2 h-4 w-4", selectedTnkb === String(t.tnkb_id) ? "opacity-100" : "opacity-0")} />
                                                        {t.Name}
                                                    </CommandItem>
                                                ))}
                                            </CommandGroup>
                                        </CommandList>
                                    </Command>
                                </PopoverContent>
                            </Popover>
                        </div>
                    </div>

                    <DialogFooter>
                        <Button variant="ghost" onClick={() => setIsModalOpen(false)}>Batal</Button>
                        <Button onClick={handleEditSubmit} disabled={isSubmitting || !selectedDriver || !selectedTnkb} className="bg-green-600 hover:bg-green-700 text-white">
                            {isSubmitting ? <Loader2 className="animate-spin w-4 h-4 mr-2" /> : <Send className="w-4 h-4 mr-2" />}
                            Update Data
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}

