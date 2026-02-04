"use client"

import { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Eye, Edit2 } from "lucide-react"
import Link from "next/link"
import dayjs from "dayjs"
import 'dayjs/locale/id'

export type SuratJalan = {
    m_inout_id: number
    document_no: string
    movement_date: string
    customer_name: string
    driver_name: string
    tnkb_no: string
}

export const columns = (onEdit: (row: SuratJalan) => void): ColumnDef<SuratJalan>[] => [
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
        header: "Date",
        cell: ({ row }) => {
            const date = row.getValue("movement_date") as string;
            return <div className="whitespace-nowrap">{date ? dayjs(date).locale('id').format('DD MMM YYYY') : "-"}</div>;
        },
    },
    {
        accessorKey: "customer_name",
        header: "Customer",
        cell: ({ row }) => <div className="truncate max-w-[150px]">{row.getValue("customer_name") || "-"}</div>,
    },
    {
        accessorKey: "driver_name",
        header: "Driver",
        cell: ({ row }) => <div>{row.getValue("driver_name") || "-"}</div>,
    },
    {
        accessorKey: "tnkb_no",
        header: "TNKB",
        cell: ({ row }) => <div className="uppercase font-mono text-xs">{row.getValue("tnkb_no") || "-"}</div>,
    },
    {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => {
            const sj = row.original;
            return (
                <div className="flex items-center gap-2">
                    <Button
                        variant="outline"
                        size="sm"
                        className="h-8 border-orange-200 text-orange-600 hover:bg-orange-50"
                        onClick={() => onEdit(sj)}
                    >
                        <Edit2 className="w-3.5 h-3.5 mr-1.5" /> Edit
                    </Button>
                    {/* <Link href={`/surat-jalan/${sj.m_inout_id}`}>
                        <Button variant="outline" size="sm" className="h-8 border-blue-200 text-blue-600 hover:bg-blue-50">
                            <Eye className="w-3.5 h-3.5 mr-1.5" /> Detail
                        </Button>
                    </Link> */}
                </div>
            )
        },
    },
]