'use client';

import { ColumnDef } from "@tanstack/react-table";
import { Button } from "@/components/ui/button";
import { RefreshCcw } from "lucide-react";
import { Badge } from "@/components/ui/badge";

export type SuratJalan = {
    m_inout_id: number;
    document_no: string;
    driver_name: string;
    tnkb_no: string;
    status: string;
    date_ordered: string;
};

export const columns = (onSync: (sj: SuratJalan) => void): ColumnDef<SuratJalan>[] => [
    {
        accessorKey: "document_no",
        header: "No. Dokumen",
        cell: ({ row }) => <span className="font-medium">{row.getValue("document_no")}</span>
    },
    {
        accessorKey: "date_ordered",
        header: "Tanggal",
    },
    {
        accessorKey: "driver_name",
        header: "Driver",
        cell: ({ row }) => row.getValue("driver_name") || <span className="text-slate-400 italic">Belum diset</span>
    },
    {
        accessorKey: "tnkb_no",
        header: "Plat Nomor",
        cell: ({ row }) => (
            <Badge variant="outline" className="font-mono">
                {row.getValue("tnkb_no") || "N/A"}
            </Badge>
        )
    },
    {
        accessorKey: "status",
        header: "Status",
        cell: ({ row }) => (
            <Badge className="bg-blue-100 text-blue-700 hover:bg-blue-100 border-none">
                {row.getValue("status")}
            </Badge>
        )
    },
];