"use client"

import React, { useEffect, useState, useCallback } from "react"
import { columns, ShipmentProgress } from "./columns"
import { DataTable } from "@/components/ui/data-table";
import { toast } from "sonner" // Atau library toast pilihan Anda
import { FileSpreadsheet, Loader2 } from "lucide-react";
import ExcelJS from 'exceljs';
import { saveAs } from 'file-saver';
import { Button } from "@/components/ui/button";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function ProgressPage() {
    const [data, setData] = useState<ShipmentProgress[]>([])
    const [isLoading, setIsLoading] = useState(true)

    const handleExportExcel = async () => {
        if (data.length === 0) return;

        const workbook = new ExcelJS.Workbook();
        const worksheet = workbook.addWorksheet('Shipment Progress');

        // 1. Definisi Header & Kolom
        worksheet.columns = [
            { header: 'NO', key: 'no', width: 5 },
            { header: 'DOC NO', key: 'documentno', width: 15 },
            { header: 'CUSTOMER', key: 'customer', width: 25 },
            { header: 'DELIVERY', key: 'delivery', width: 12 },
            { header: 'ON DPK', key: 'ondpk', width: 12 },
            { header: 'ON DRIVER', key: 'ondriver', width: 12 },
            { header: 'AT CUST', key: 'oncustomer', width: 12 },
            { header: 'OUT CUST', key: 'outcustomer', width: 12 },
            { header: 'ON MKT', key: 'comebackfat', width: 12 },
            { header: 'FINISH FAT', key: 'finishfat', width: 12 },
            { header: 'DRIVER', key: 'driver', width: 20 },
            { header: 'TNKB', key: 'tnkb', width: 15 },
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
    const fetchShipments = useCallback(async (authToken: string) => {
        try {
            setIsLoading(true)
            const res = await fetch(`${API_BASE_URL}/shipments/progress`, {
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

        if (storedToken) {
            fetchShipments(storedToken)
        } else {
            setIsLoading(false)
            toast.error("Token tidak ditemukan, silakan login")
            // Opsi: redirect ke login page di sini
        }
    }, [fetchShipments])

    return (
        <div className="container mx-auto py-4 px-0">
            <div className="flex flex-col gap-1 mb-8">
                <h1 className="text-2xl font-bold tracking-tight text-slate-900">
                    Shipment Progress
                </h1>
                <p className="text-sm text-slate-500">
                    Pantau status pengiriman surat jalan secara real-time.
                </p>
            </div>

            <Button
                variant="outline"
                size="sm"
                onClick={handleExportExcel}
                disabled={isLoading || data.length === 0}
                className="text-emerald-600 border-emerald-200 hover:bg-emerald-50 bg-white"
            >
                <FileSpreadsheet className="w-4 h-4 mr-2" />
                Export Excel
            </Button>

            <DataTable
                columns={columns}
                data={data}
                loading={isLoading}
            />
        </div>
    )
}