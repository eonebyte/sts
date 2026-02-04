"use client"

import { ColumnDef } from "@tanstack/react-table"
import { CheckCircle2, Circle, Truck, Eye } from "lucide-react"
import { Button } from "@/components/ui/button"
import Link from "next/link"

export type ShipmentProgress = {
    documentno: string
    customer: string
    movementdate: string
    driver: string
    tnkb: string
    delivery: number
    ondpk: number
    ondriver: number
    oncustomer: number
    outcustomer: number
    comebackdpk: number
    comebackdel: number
    comebackmkt: number
    comebackfat: number
    finishfat: number
}

const StatusCell = (val: number, isCurrent: boolean) => {
    if (val === 1) return <CheckCircle2 className="w-5 h-5 text-green-500 mx-auto" />;
    if (isCurrent) return (
        <div className="relative flex justify-center">
            <Truck className="w-5 h-5 text-blue-500 animate-pulse" />
            <span className="absolute top-0 right-1/4 h-2 w-2 rounded-full bg-blue-500 animate-ping" />
        </div>
    );
    return <Circle className="w-4 h-4 text-slate-200 mx-auto" />;
};

export const columns: ColumnDef<ShipmentProgress>[] = [
    {
        accessorKey: "documentno",
        header: "Doc No",
        cell: ({ row }) => <span className="font-bold text-xs">{row.getValue("documentno")}</span>,
    },
    {
        accessorKey: "customer",
        header: "Customer",
        cell: ({ row }) => <span className="text-xs truncate block max-w-[100px]">{row.getValue("customer")}</span>,
    },
    {
        accessorKey: "delivery",
        header: "Delivery",
        enableColumnFilter: true, // Diaktifkan
        cell: ({ row }) => StatusCell(row.original.delivery, false),
    },
    {
        accessorKey: "ondpk",
        header: "On DPK",
        enableColumnFilter: true, // Diaktifkan
        cell: ({ row }) => {
            const d = row.original;
            const isCurrent = d.ondpk === 0 && d.delivery === 1;
            return StatusCell(d.ondpk, isCurrent);
        },
    },
    {
        accessorKey: "ondriver",
        header: "On Driver",
        enableColumnFilter: true, // Diaktifkan
        cell: ({ row }) => {
            const d = row.original;
            const isCurrent = d.ondriver === 0 && d.ondpk === 1;
            return StatusCell(d.ondriver, isCurrent);
        },
    },
    {
        accessorKey: "oncustomer",
        header: "At Cust",
        enableColumnFilter: true, // Diaktifkan
        cell: ({ row }) => {
            const d = row.original;
            const isCurrent = d.oncustomer === 0 && d.ondriver === 1;
            return StatusCell(d.oncustomer, isCurrent);
        },
    },
    {
        accessorKey: "outcustomer",
        header: "Out Cust",
        enableColumnFilter: true, // Diaktifkan
        cell: ({ row }) => {
            const d = row.original;
            const isCurrent = d.outcustomer === 0 && d.oncustomer === 1;
            return StatusCell(d.outcustomer, isCurrent);
        },
    },
    {
        accessorKey: "comebackfat",
        header: "On MKT",
        enableColumnFilter: true, // Diaktifkan
        cell: ({ row }) => StatusCell(row.original.comebackfat, false),
    },
    {
        accessorKey: "finishfat",
        header: "Finish FAT",
        enableColumnFilter: true, // Diaktifkan
        cell: ({ row }) => StatusCell(row.original.finishfat, false),
    },
    {
        accessorKey: "driver",
        header: "Driver",
        cell: ({ row }) => <span className="text-[11px]">{row.getValue("driver") || "-"}</span>,
    },
    {
        accessorKey: "tnkb",
        header: "TNKB",
        cell: ({ row }) => <span className="text-[11px] whitespace-nowrap">{row.getValue("tnkb") || "-"}</span>,
    },
    {
        id: "actions",
        header: "Action",
        enableColumnFilter: false,
        cell: ({ row }) => (
            <Link href={`/shipment/${row.original.documentno}`}>
                <Button variant="ghost" size="icon" className="h-8 w-8 text-blue-600 hover:bg-blue-50">
                    <Eye className="w-4 h-4" />
                </Button>
            </Link>
        ),
    },
]