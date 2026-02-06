'use client';

import { ColumnDef } from "@tanstack/react-table";
import { Button } from "@/components/ui/button";
import { Edit2, XCircle } from "lucide-react"; // Icon untuk batal
import { Badge } from "@/components/ui/badge";

export type SuratJalan = {
    m_inout_id: number;
    adw_sts_id: number; // Tambahkan ini sesuai kebutuhan submit
    document_no: string;
    driver_name: string;
    tnkb_no: string;
    status: string;
    date_ordered: string;
};

// Tambahkan parameter onCancel pada fungsi columns
export const columns = (
    onCancel: (id: number, status: string) => void,
    onEdit: (row: SuratJalan) => void
): ColumnDef<SuratJalan>[] => [
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
        {
            id: "actions",
            header: "Aksi",
            cell: ({ row }) => {
                const sj = row.original;
                const { status, m_inout_id: mInOutId } = sj;

                // Tombol hanya muncul jika status sesuai
                if (status !== "HO: DPK_TO_DRIVER") {
                    return null;
                }

                return (
                    <div className="flex items-center gap-2">
                        {/* Tombol Edit - Warna Oranye Profesional */}
                        <Button
                            variant="outline"
                            size="sm"
                            className="h-8 px-3 border-orange-200 text-orange-600 hover:bg-orange-50 hover:text-orange-700 transition-colors"
                            onClick={() => onEdit(sj)}
                        >
                            <Edit2 className="w-3.5 h-3.5 mr-1.5" />
                            <span>Edit</span>
                        </Button>

                        {/* Tombol Batal - Destructive Modern */}
                        <Button
                            variant="destructive"
                            size="sm"
                            className="h-8 px-3 transition-opacity hover:opacity-90"
                            onClick={() => onCancel(mInOutId, status)}
                        >
                            <XCircle className="w-3.5 h-3.5 mr-1.5" />
                            <span>Batal</span>
                        </Button>
                    </div>
                );
            }
        }
    ];
