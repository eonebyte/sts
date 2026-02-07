"use client"

import { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox" // Pastikan sudah install checkbox
import { Eye } from "lucide-react"
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

export const columns: ColumnDef<SuratJalan>[] = [
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
        accessorKey: "driver_name",
        header: "Driver",
        cell: ({ row }) => <div>{row.getValue("driver_name") || "-"}</div>,
    },
    {
        accessorKey: "tnkb_no",
        header: "TNKBNo",
        cell: ({ row }) => <div className="uppercase font-mono text-xs">{row.getValue("tnkb_no") || "-"}</div>,
    },
    // {
    //     id: "actions",
    //     header: "Action",
    //     cell: ({ row }) => {
    //         const sj = row.original;
    //         return (
    //             <div className="text-left">
    //                 <Link href={`/surat-jalan/${sj.m_inout_id}`}>
    //                     <Button variant="outline" size="sm" className="h-8 border-blue-200 text-blue-600 hover:bg-blue-50">
    //                         <Eye className="w-3.5 h-3.5 mr-1.5" /> Detail
    //                     </Button>
    //                 </Link>
    //             </div>
    //         )
    //     },
    // },
]