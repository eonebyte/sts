"use client"

import React, { useEffect, useState, useCallback } from "react"
import { columns, ShipmentProgress } from "./columns"
import { DataTable } from "@/components/ui/data-table";
import { toast } from "sonner" // Atau library toast pilihan Anda
import { FileSpreadsheet, History, Loader2, RefreshCcw } from "lucide-react";
import ExcelJS from 'exceljs';
import { saveAs } from 'file-saver';
import { Button } from "@/components/ui/button";
import { format } from 'date-fns';
import { id } from "date-fns/locale";
import { DateRangeFilter } from '@/components/ui/date-range-filter';
import { useAuth } from "@/hooks/useAuth";


const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function ProgressPage() {
    const { isAuthorized } = useAuth();
    const [token, setToken] = useState<string | null>(null);
    const [data, setData] = useState<ShipmentProgress[]>([])
    const [isLoading, setIsLoading] = useState(true)

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

    const handleExportExcel = async () => {
        if (data.length === 0) return;

        const workbook = new ExcelJS.Workbook();
        const worksheet = workbook.addWorksheet('Shipment Progress');

        // 1. Definisi Header & Kolom
        worksheet.columns = [
            { header: 'NO', key: 'no', width: 5 },
            { header: 'DOC NO', key: 'documentno', width: 15 },
            { header: 'CUSTOMER', key: 'customer', width: 25 },
            { header: 'DRIVER', key: 'driver', width: 20 },
            { header: 'TNKB', key: 'tnkb', width: 15 },
            { header: 'DELIVERY', key: 'delivery', width: 12 },
            { header: 'ON DPK', key: 'ondpk', width: 12 },
            { header: 'ON DRIVER', key: 'ondriver', width: 12 },
            { header: 'AT CUST', key: 'oncustomer', width: 12 },
            { header: 'OUT CUST', key: 'outcustomer', width: 12 },
            { header: 'ON MKT', key: 'comebackfat', width: 12 },
            { header: 'FINISH FAT', key: 'finishfat', width: 12 },
        ];

        // Helper untuk mengubah 1/0 menjadi teks
        const formatStatus = (val: number) => val === 1 ? "DONE" : "PENDING";

        // 2. Mapping Data ke Row
        data.forEach((item, index) => {
            const row = worksheet.addRow({
                no: index + 1,
                documentno: item.documentno,
                customer: item.customer,
                delivery: formatStatus(item.delivery),
                ondpk: formatStatus(item.ondpk),
                ondriver: formatStatus(item.ondriver),
                oncustomer: formatStatus(item.oncustomer),
                outcustomer: formatStatus(item.outcustomer),
                comebackfat: formatStatus(item.comebackfat),
                finishfat: formatStatus(item.finishfat),
                driver: item.driver || "-",
                tnkb: item.tnkb || "-",
            });

            // 3. Styling baris berdasarkan nilai (Opsional: Hijau jika DONE)
            row.eachCell((cell, colNumber) => {
                if (cell.value === "DONE" && colNumber > 3 && colNumber < 11) {
                    cell.font = { color: { argb: 'FF007500' }, bold: true }; // Hijau gelap
                }
            });
        });

        // 4. Styling Header
        worksheet.getRow(1).eachCell((cell) => {
            cell.font = { bold: true, color: { argb: 'FFFFFFFF' } };
            cell.fill = {
                type: 'pattern',
                pattern: 'solid',
                fgColor: { argb: 'FF1E293B' } // Slate-800
            };
            cell.alignment = { vertical: 'middle', horizontal: 'center' };
        });

        // 5. Download
        const buffer = await workbook.xlsx.writeBuffer();
        const blob = new Blob([buffer], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' });
        saveAs(blob, `Shipment_Progress_${new Date().toISOString().split('T')[0]}.xlsx`);
    };

    // Fungsi fetch diletakkan di dalam component agar bisa mengakses state/context jika perlu
    const fetchShipments = useCallback(async (authToken: string, from?: Date, to?: Date) => {
        try {
            setIsLoading(true)

            const params = new URLSearchParams();
            if (from) {
                params.append('dateFrom', format(from, 'yyyy-MM-dd'));
            }
            if (to) {
                params.append('dateTo', format(to, 'yyyy-MM-dd'));
            }

            const url = `${API_BASE_URL}/shipments/progress?${params.toString()}`

            const res = await fetch(url, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${authToken}`,
                    'Content-Type': 'application/json'
                },
            })

            if (!res.ok) {
                if (res.status === 401) throw new Error('Sesi telah berakhir, silakan login kembali')
                throw new Error('Gagal mengambil data dari server')
            }

            const result = await res.json()
            setData(result.data || [])
        } catch (error: any) {
            toast.error(error.message)
            console.error("Fetch error:", error)
        } finally {
            setIsLoading(false)
        }
    }, [])

    useEffect(() => {
        // Ambil token dari localStorage saat component mount
        const storedToken = localStorage.getItem('token')
        setToken(storedToken)
        if (isAuthorized && storedToken) {
            fetchShipments(storedToken)
        }
    }, [isAuthorized])

    // Fetch ulang ketika date range berubah
    useEffect(() => {
        if (token && isAuthorized && (dateRange.from || dateRange.to)) {
            fetchShipments(token, dateRange.from, dateRange.to);
        }
    }, [dateRange]);

    return (
        <div className="container mx-auto py-0 px-0">

            <div className="flex mb-2 mt-0 flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div className="flex items-center gap-2">
                    <div className="p-2 bg-blue-100 rounded-lg">
                        <History className="w-5 h-5 text-blue-600" />
                    </div>
                    <div>
                        <h1 className="text-xl font-bold text-slate-800">Shipment Progress</h1>
                        <p className="text-xs text-slate-500">Pantau status pengiriman surat jalan secara real-time.
                        </p>
                    </div>
                </div>

                <div className="flex items-center gap-2 w-full md:w-auto">

                    <DateRangeFilter
                        dateRange={dateRange}
                        setDateRange={setDateRange}
                        handleResetFilter={handleResetFilter}
                    />

                    <Button
                        variant="outline"
                        size="sm"
                        onClick={handleExportExcel}
                        disabled={isLoading || data.length === 0}
                        className="text-emerald-600 border-emerald-200 hover:bg-emerald-50 bg-white"
                    >
                        <FileSpreadsheet className="w-4 h-4 mr-2" />
                    </Button>
                </div>
            </div>

            <DataTable
                columns={columns}
                data={data}
                loading={isLoading}
            />
        </div>
    )
}
