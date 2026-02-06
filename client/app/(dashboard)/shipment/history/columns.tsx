import { ColumnDef } from "@tanstack/react-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { FileText } from "lucide-react";

// Tipe data dasar dari API
export type HistorySJ = {
    m_inout_id: number;
    document_no: string;
    movement_date: string;
    customer_name: string;
    driver_name: string;
    status: string;
    bundle_no: string | null;
    attachment_path: string | null;
};

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Tipe data setelah di-grouping
export type GroupedHistorySJ = Omit<HistorySJ, 'bundle_no' | 'attachment_path'> & {
    bundles: { no: string | null; path: string | null }[];
};

export const columns: ColumnDef<GroupedHistorySJ>[] = [
    {
        id: "no",
        header: "NO",
        cell: ({ row }) => (
            <div className="text-center w-8 text-slate-500 font-medium">
                {row.index + 1}
            </div>
        ),
    },
    { accessorKey: "document_no", header: "No. Surat Jalan" },
    { accessorKey: "customer_name", header: "Customer" },
    { accessorKey: "driver_name", header: "Driver" },
    {
        accessorKey: "status",
        header: "Status Terakhir",
        cell: ({ row }) => (
            <Badge variant="outline" className="bg-slate-50 text-[10px] whitespace-nowrap">
                {row.original.status}
            </Badge>
        ),
    },
    {
        id: "bundles",
        header: "No. Bundle",
        // 1. Pastikan accessorFn mengembalikan string kosong jika tidak ada data
        accessorFn: (row) => {
            return row.bundles.map(b => b.no || "").join(" ") || "";
        },
        // 2. Tambahkan flag ini untuk memastikan kolom selalu bisa difilter
        enableColumnFilter: true,
        filterFn: (row, columnId, filterValue) => {
            const cellValue = row.getValue(columnId) as string;
            if (!filterValue) return true;
            return cellValue.toLowerCase().includes(filterValue.toLowerCase());
        },
        cell: ({ row }) => (
            <div className="flex flex-col gap-1 py-1">
                {row.original.bundles.map((b, i) => (
                    <span key={i} className="text-[10px] bg-slate-100 px-2 py-0.5 rounded border border-slate-200 w-fit font-mono">
                        {b.no || "-"}
                    </span>
                ))}
            </div>
        ),
    },
    {
        id: "actions",
        header: "PDF",
        cell: ({ row }) => (
            <div className="flex flex-col gap-1 items-center">
                {row.original.bundles.map((b, i) => (
                    b.path ? (
                        <Button
                            key={i}
                            variant="ghost"
                            size="sm"
                            className="h-6 w-6 text-red-600 hover:text-red-700 hover:bg-red-50 p-0"
                            onClick={() => window.open(`${API_BASE_URL}/${b.path}`, '_blank')}
                        >
                            <FileText className="w-4 h-4" />
                        </Button>
                    ) : <span key={i} className="text-[10px] text-slate-300 h-6 flex items-center">-</span>
                ))}
            </div>
        ),
    },
];