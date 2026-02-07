'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { Button } from "@/components/ui/button";
import { Loader2, PackageCheck, Send, User, Car, Check, ChevronsUpDown } from "lucide-react";
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
import { toast } from 'sonner';
import { format } from 'date-fns';
import { id } from "date-fns/locale";
import { DateRangeFilter } from '@/components/ui/date-range-filter';

interface MasterDataDriver {
    driver_by: string | number;
    Name: string;
}

interface MasterDataTnkb {
    tnkb_id: string | number;
    Name: string;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function Page() {
    const [token, setToken] = useState<string | null>(null);
    const { isAuthorized, user } = useAuth();
    const [shipments, setShipments] = useState<SuratJalan[]>([]);
    const [loading, setLoading] = useState(true);
    const [isSubmitting, setIsSubmitting] = useState(false);

    const [drivers, setDrivers] = useState<MasterDataDriver[]>([]);
    const [tnkbs, setTnkbs] = useState<MasterDataTnkb[]>([]);

    // State menyimpan ID sebagai string
    const [selectedDriver, setSelectedDriver] = useState<string>("");
    const [selectedTnkb, setSelectedTnkb] = useState<string>("");

    const [openDriver, setOpenDriver] = useState(false);
    const [openTnkb, setOpenTnkb] = useState(false);

    const [rowSelection, setRowSelection] = useState<Record<string, boolean>>({});
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

            const url = `${API_BASE_URL}/shipments/preparetoleave?${params.toString()}`

            const res = await fetch(url, {
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

    const fetchMasterData = async (authToken: string) => {
        try {
            const [resDriver, resTnkb] = await Promise.all([
                fetch(`${API_BASE_URL}/shipments/drivers`, { headers: { Authorization: `Bearer ${authToken}` } }),
                fetch(`${API_BASE_URL}/shipments/tnkbs`, { headers: { Authorization: `Bearer ${authToken}` } })
            ]);
            const dataDriver = await resDriver.json();
            const dataTnkb = await resTnkb.json();
            setDrivers(dataDriver.data || []);
            setTnkbs(dataTnkb.data || []);
        } catch (error) {
            console.error("Gagal load master data", error);
        }
    };

    useEffect(() => {
        const storedToken = localStorage.getItem('token');
        setToken(storedToken);
        if (storedToken && isAuthorized) {
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

    const selectedRowsData = shipments.filter((_, index) => rowSelection[index]);

    const handleHandoverConfirm = async () => {
        if (!selectedDriver || !selectedTnkb || !token) return;

        setIsSubmitting(true);
        const payload = {
            m_inout_ids: selectedRowsData.map(item => item.m_inout_id),
            status: "HO: DPK_TO_DRIVER",
            user_id: parseInt(user.user_id),
            driver_by: parseInt(selectedDriver), // Kirim ID ke server
            tnkb_id: parseInt(selectedTnkb),     // Kirim ID ke server
            notes: 'ho dpk to driver'
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

            toast.success("Handover Berhasil!");
            setIsModalOpen(false);
            setRowSelection({});
            setSelectedDriver("");
            setSelectedTnkb("");
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
                columns={columns}
                data={shipments}
                rowSelection={rowSelection}
                setRowSelection={setRowSelection}
                loading={loading}
            />

            <div className="sticky bottom-4 flex items-center justify-between p-4 bg-white border shadow-lg rounded-xl">
                <div className="flex flex-col">
                    <span className="text-sm font-medium">{selectedRowsData.length} Terpilih</span>
                </div>

                <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
                    <DialogTrigger asChild>
                        <Button disabled={selectedRowsData.length === 0} className="bg-green-600">
                            <PackageCheck className="w-4 h-4 mr-2" /> Handover
                        </Button>
                    </DialogTrigger>
                    <DialogContent className="sm:max-w-[450px]">
                        <DialogHeader>
                            <DialogTitle>Lengkapi Data Pengiriman</DialogTitle>
                            <DialogDescription>Cari dan pilih Driver serta Plat Nomor.</DialogDescription>
                        </DialogHeader>

                        <div className="space-y-6 py-4">
                            {/* COMBOBOX DRIVER */}
                            <div className="flex flex-col gap-2">
                                <label className="text-sm font-semibold flex items-center gap-2">
                                    <User className="w-4 h-4" /> Driver
                                </label>
                                <Popover open={openDriver} onOpenChange={setOpenDriver}>
                                    <PopoverTrigger asChild>
                                        <Button variant="outline" role="combobox" className="justify-between w-full">
                                            {selectedDriver
                                                ? drivers.find((d) => String(d.driver_by) === selectedDriver)?.Name
                                                : "Cari nama driver..."}
                                            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                                        </Button>
                                    </PopoverTrigger>
                                    <PopoverContent className="w-[400px] p-0">
                                        <Command>
                                            <CommandInput placeholder="Ketik nama driver..." />
                                            <CommandList>
                                                <CommandEmpty>Driver tidak ditemukan.</CommandEmpty>
                                                <CommandGroup>
                                                    {drivers.map((d) => (
                                                        <CommandItem
                                                            key={d.driver_by}
                                                            value={d.Name} // Value ini untuk filtering search
                                                            onSelect={() => {
                                                                setSelectedDriver(String(d.driver_by));
                                                                setOpenDriver(false);
                                                            }}
                                                        >
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

                            {/* COMBOBOX TNKB */}
                            <div className="flex flex-col gap-2">
                                <label className="text-sm font-semibold flex items-center gap-2">
                                    <Car className="w-4 h-4" /> Plat Nomor (TNKB)
                                </label>
                                <Popover open={openTnkb} onOpenChange={setOpenTnkb}>
                                    <PopoverTrigger asChild>
                                        <Button variant="outline" role="combobox" className="justify-between w-full">
                                            {selectedTnkb
                                                ? tnkbs.find((t) => String(t.tnkb_id) === selectedTnkb)?.Name
                                                : "Cari plat nomor..."}
                                            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                                        </Button>
                                    </PopoverTrigger>
                                    <PopoverContent className="w-[400px] p-0">
                                        <Command>
                                            <CommandInput placeholder="Ketik plat nomor..." />
                                            <CommandList>
                                                <CommandEmpty>TNKB tidak ditemukan.</CommandEmpty>
                                                <CommandGroup>
                                                    {tnkbs.map((t, i) => (
                                                        <CommandItem
                                                            key={`${t.tnkb_id}-${i}`}
                                                            value={t.Name}
                                                            onSelect={() => {
                                                                setSelectedTnkb(String(t.tnkb_id));
                                                                setOpenTnkb(false);
                                                            }}
                                                        >
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
                            <Button
                                onClick={handleHandoverConfirm}
                                className="bg-green-600"
                                disabled={isSubmitting || !selectedDriver || !selectedTnkb}
                            >
                                {isSubmitting ? <Loader2 className="animate-spin w-4 h-4 mr-2" /> : <Send className="w-4 h-4 mr-2" />}
                                Submit OK
                            </Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>
        </div>
    );
}