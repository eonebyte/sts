'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import {
    Loader2,
    History,
    RefreshCcw,
    Calendar,
    FileSpreadsheet
} from "lucide-react";
import { DataTable } from "@/components/ui/data-table";
import { columns, HistorySJ, GroupedHistorySJ } from "./columns";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import ExcelJS from 'exceljs';
import { saveAs } from 'file-saver';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function HistoryPage() {
    const { isAuthorized } = useAuth();
    const [data, setData] = useState<GroupedHistorySJ[]>([]);
    const [loading, setLoading] = useState(true);

    const [dateFrom, setDateFrom] = useState(
        new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
    );
    const [dateTo, setDateTo] = useState(
        new Date().toISOString().split('T')[0]
    );

    const fetchHistory = async () => {
        setLoading(true);
        const token = localStorage.getItem('token');
        try {
            const res = await fetch(`${API_BASE_URL}/shipments/history?dateFrom=${dateFrom}&dateTo=${dateTo}`, {
                headers: { Authorization: `Bearer ${token}` },
            });
            const result = await res.json();
            const rawData: HistorySJ[] = result.success ? result.data : [];

            // LOGIC GROUPING: Menggabungkan bundle yang memiliki document_no yang sama
            const grouped = rawData.reduce((acc: GroupedHistorySJ[], current) => {
                const existing = acc.find(item => item.document_no === current.document_no);
                if (existing) {
                    existing.bundles.push({
                        no: current.bundle_no,
                        path: current.attachment_path
                    });
                } else {
                    acc.push({
                        m_inout_id: current.m_inout_id,
                        document_no: current.document_no,
                        movement_date: current.movement_date,
                        customer_name: current.customer_name,
                        driver_name: current.driver_name,
                        status: current.status,
                        bundles: [{
                            no: current.bundle_no,
                            path: current.attachment_path
                        }]
                    });
                }
                return acc;
            }, []);

            setData(grouped);
        } catch (error) {
            console.error("Fetch history error:", error);
            setData([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (isAuthorized) fetchHistory();
    }, [isAuthorized]);

    const handleExportExcel = async () => {
        if (data.length === 0) return;

        const workbook = new ExcelJS.Workbook();
        const worksheet = workbook.addWorksheet('History Handover');

        worksheet.columns = [
            { header: 'NO', key: 'no', width: 5 },
            { header: 'NO. SURAT JALAN', key: 'document_no', width: 20 },
            { header: 'TANGGAL', key: 'movement_date', width: 15 },
            { header: 'CUSTOMER', key: 'customer_name', width: 25 },
            { header: 'DRIVER', key: 'driver_name', width: 20 },
            { header: 'STATUS', key: 'status', width: 20 },
            { header: 'NO. BUNDLE', key: 'bundle_no', width: 35 },
            { header: 'LINK PDF UTAMA', key: 'pdf_link', width: 15 },
        ];

        data.forEach((item, index) => {
            const bundleList = item.bundles.map(b => b.no).filter(Boolean).join(", ");

            const row = worksheet.addRow({
                no: index + 1,
                document_no: item.document_no,
                movement_date: item.movement_date,
                customer_name: item.customer_name,
                driver_name: item.driver_name || '-',
                status: item.status,
                bundle_no: bundleList || '-',
                pdf_link: item.bundles[0]?.path ? "Lihat PDF" : "-"
            });

            if (item.bundles[0]?.path) {
                const cell = row.getCell('pdf_link');
                cell.value = {
                    text: 'Lihat PDF',
                    hyperlink: `${API_BASE_URL}/${item.bundles[0].path}`,
                };
                cell.font = { color: { argb: 'FF0000FF' }, underline: true };
            }
        });

        worksheet.getRow(1).eachCell((cell) => {
            cell.font = { bold: true, color: { argb: 'FFFFFFFF' } };
            cell.fill = {
                type: 'pattern',
                pattern: 'solid',
                fgColor: { argb: 'FF1F2937' }
            };
            cell.alignment = { vertical: 'middle', horizontal: 'center' };
        });

        const buffer = await workbook.xlsx.writeBuffer();
        const blob = new Blob([buffer], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' });
        saveAs(blob, `History_Handover_${dateFrom}_to_${dateTo}.xlsx`);
    };

    if (!isAuthorized) {
        return (
            <div className="h-screen flex items-center justify-center">
                <Loader2 className="animate-spin text-blue-600" />
            </div>
        );
    }

    return (
        <div className="p-4 space-y-4">
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div className="flex items-center gap-2">
                    <div className="p-2 bg-blue-100 rounded-lg">
                        <History className="w-5 h-5 text-blue-600" />
                    </div>
                    <div>
                        <h1 className="text-xl font-bold text-slate-800">History Handover</h1>
                        <p className="text-xs text-slate-500">Daftar semua surat jalan dan berkas bundle PDF.</p>
                    </div>
                </div>

                <div className="flex items-center gap-2 w-full md:w-auto">
                    <div className="flex items-center bg-white border rounded-lg px-2 shadow-sm">
                        <Calendar className="w-4 h-4 text-slate-400 mr-2" />
                        <Input
                            type="date"
                            value={dateFrom}
                            onChange={(e) => setDateFrom(e.target.value)}
                            className="border-none focus-visible:ring-0 text-xs h-8 w-32"
                        />
                        <span className="text-slate-400 px-1">-</span>
                        <Input
                            type="date"
                            value={dateTo}
                            onChange={(e) => setDateTo(e.target.value)}
                            className="border-none focus-visible:ring-0 text-xs h-8 w-32"
                        />
                    </div>

                    <Button
                        variant="outline"
                        size="sm"
                        onClick={handleExportExcel}
                        disabled={loading || data.length === 0}
                        className="text-emerald-600 border-emerald-200 hover:bg-emerald-50 shadow-sm"
                    >
                        <FileSpreadsheet className="w-4 h-4 mr-2" />
                        Export
                    </Button>

                    <Button variant="outline" size="sm" onClick={fetchHistory} disabled={loading}>
                        <RefreshCcw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
                    </Button>
                </div>
            </div>

            <div className="bg-white rounded-xl shadow-sm border">
                <DataTable
                    columns={columns}
                    data={data}
                    loading={loading}
                />
            </div>
        </div>
    );
}