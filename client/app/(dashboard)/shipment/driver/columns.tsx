'use client';

import { ColumnDef } from "@tanstack/react-table";
import { MoreHorizontal, Edit } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export type Driver = {
    driver_by: number;
    Name: string; // Sesuaikan dengan accessorKey Anda
};

// Kita buat interface untuk props agar bisa menerima fungsi onEdit dari page
interface ColumnProps {
    onEdit: (driver: Driver) => void;
}

export const getColumns = ({ onEdit }: ColumnProps): ColumnDef<Driver>[] => [
    {
        id: "no",
        header: "NO",
        cell: ({ row }) => <div className="text-center w-8">{row.index + 1}</div>,
    },
    {
        accessorKey: "Name",
        header: "Driver",
        cell: ({ row }) => row.getValue("Name") || <span className="text-slate-400 italic">Belum diset</span>
    },
    {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => {
            const driver = row.original;
            return (
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button variant="ghost" className="h-8 w-8 p-0">
                            <MoreHorizontal className="h-4 w-4" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        <DropdownMenuLabel>Aksi</DropdownMenuLabel>
                        <DropdownMenuItem onClick={() => onEdit(driver)}>
                            <Edit className="mr-2 h-4 w-4" /> Edit Driver
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            );
        },
    },
];