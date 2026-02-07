"use client"

import { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox" // Pastikan sudah install checkbox
import { Eye, XCircle } from "lucide-react"
import Link from "next/link"
import dayjs from "dayjs"
import 'dayjs/locale/id'

export type SuratJalan = {
    m_inout_id: number
    document_no: string
    movement_date: string
    customer_name: string
    spp_no: string
    status: string
}

export const columns = (onCancel: (id: number, status: string) => void): ColumnDef<SuratJalan>[] => [
    {
        id: "select",
        header: ({ table }) => (
            <Checkbox
                checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && "indeterminate")}
                onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
                aria-label="Select all"
            />
        ),
        cell: ({ row }) => (
            <Checkbox
                checked={row.getIsSelected()}
                onCheckedChange={(value) => row.toggleSelected(!!value)}
                aria-label="Select row"
            />
        ),
        enableSorting: false,
        enableHiding: false,
    },
    {
        id: "no",
        header: "No",
        cell: ({ row }) => <div className="w-4">{row.index + 1}</div>,
    },
    {
        accessorKey: "document_no",
        header: "Document No",
        cell: ({ row }) => <div className="font-bold text-slate-700">{row.getValue("document_no")}</div>,
    },
    {
        accessorKey: "movement_date",
        header: "Movement Date",
        cell: ({ row }) => {
            const date = row.getValue("movement_date") as string;
            return (
                <div className="whitespace-nowrap">
                    {date ? dayjs(date).locale('id').format('DD MMM YYYY') : "-"}
                </div>
            );
        },
    },
    {
        accessorKey: "customer_name",
        header: "Customer",
        cell: ({ row }) => (
            <div className="max-w-[150px] truncate font-medium">
                {row.getValue("customer_name") || "-"}
            </div>
        ),
    },
    {
        accessorKey: "spp_no",
        header: "SPP No",
        cell: ({ row }) => <div>{row.getValue("spp_no") || "-"}</div>,
    },
    {
        id: "actions",
        header: "Aksi",
        cell: ({ row }) => {
            const status = row.original.status;
            const mInOutId = row.original.m_inout_id;

            // Tombol hanya muncul jika status sesuai
            if (status !== "HO: DEL_TO_MKT") {
                return null;
            }

            return (
                <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 text-red-600 hover:text-red-700 hover:bg-red-50"
                    onClick={() => onCancel(mInOutId, status)}
                >
                    <XCircle className="w-4 h-4" />
                    Reject
                </Button>
            );
        }
    }
]
