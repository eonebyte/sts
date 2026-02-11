"use client"

import * as React from "react"
import {
    ColumnDef,
    flexRender,
    getCoreRowModel,
    getPaginationRowModel,
    getSortedRowModel,
    getFilteredRowModel,
    SortingState,
    ColumnFiltersState,
    useReactTable,
} from "@tanstack/react-table"

import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
    ChevronLeft, ChevronRight, Loader2, Search,
    ChevronsLeft,
    ChevronsRight
} from "lucide-react"

interface DataTableProps<TData, TValue> {
    columns: ColumnDef<TData, TValue>[]
    data: TData[]
    loading?: boolean
    rowSelection?: any
    setRowSelection?: React.Dispatch<React.SetStateAction<any>>
}

export function DataTable<TData, TValue>({
    columns,
    data = [],
    loading = false,
    rowSelection = {},
    setRowSelection,
}: DataTableProps<TData, TValue>) {
    const [sorting, setSorting] = React.useState<SortingState>([])
    const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([])

    const table = useReactTable({
        data,
        columns,
        getCoreRowModel: getCoreRowModel(),
        getPaginationRowModel: getPaginationRowModel(),
        onSortingChange: setSorting,
        getSortedRowModel: getSortedRowModel(),
        onColumnFiltersChange: setColumnFilters,
        getFilteredRowModel: getFilteredRowModel(),
        onRowSelectionChange: setRowSelection,
        state: {
            sorting,
            columnFilters,
            rowSelection,
        },
    })

    return (
        <div className="space-y-4">
            <div className="rounded-md border bg-white overflow-hidden shadow-sm mb-2">
                <div className="overflow-x-auto"> {/* Container scroll horizontal */}
                    <Table>
                        <TableHeader className="bg-slate-50">
                            {table.getHeaderGroups().map((headerGroup) => (
                                <TableRow key={headerGroup.id} className="hover:bg-transparent">
                                    {headerGroup.headers.map((header) => (
                                        <TableHead key={header.id} className="py-3 px-2 border-r last:border-r-0">
                                            <div className="flex flex-col gap-2">
                                                {/* Header Label */}
                                                <div className="text-[11px] uppercase tracking-wider font-bold text-slate-700">
                                                    {header.isPlaceholder
                                                        ? null
                                                        : flexRender(
                                                            header.column.columnDef.header,
                                                            header.getContext()
                                                        )}
                                                </div>

                                                {/* Filter Input Per Kolom */}
                                                {header.column.getCanFilter() ? (
                                                    <div className="relative">
                                                        <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3 w-3 text-slate-400" />
                                                        <Input
                                                            placeholder="Filter..."
                                                            value={(header.column.getFilterValue() as string) ?? ""}
                                                            onChange={(event) =>
                                                                header.column.setFilterValue(event.target.value)
                                                            }
                                                            className="h-7 pl-7 pr-2 text-[10px] w-full min-w-[80px] bg-white border-slate-200 focus-visible:ring-blue-500"
                                                        />
                                                    </div>
                                                ) : (
                                                    <div className="h-7" /> // Spacer jika kolom tidak bisa difilter
                                                )}
                                            </div>
                                        </TableHead>
                                    ))}
                                </TableRow>
                            ))}
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={columns.length} className="h-32 text-center">
                                        <div className="flex flex-col items-center justify-center gap-2">
                                            <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
                                            <p className="text-sm text-slate-500 font-medium">Memuat data...</p>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            ) : table.getRowModel().rows.length ? (
                                table.getRowModel().rows.map((row) => (
                                    <TableRow
                                        key={row.id}
                                        data-state={row.getIsSelected() && "selected"}
                                        className="hover:bg-slate-50/50 transition-colors border-b"
                                    >
                                        {row.getVisibleCells().map((cell) => (
                                            <TableCell key={cell.id} className="px-2 py-1">
                                                {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                            </TableCell>
                                        ))}
                                    </TableRow>
                                ))
                            ) : (
                                <TableRow>
                                    <TableCell colSpan={columns.length} className="h-24 text-center text-slate-500 italic">
                                        Tidak ada data yang ditemukan.
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </div>
            </div>

            {/* Pagination */}
            <div className="flex items-center justify-between px-2">
                <p className="text-xs text-slate-500 font-medium">
                    Showing <span className="text-slate-900 font-bold">{table.getRowModel().rows.length}</span> rows of{" "}
                    <span className="text-slate-900 font-bold">{table.getFilteredRowModel().rows.length}</span> total data
                </p>
                <div className="flex items-center space-x-2">
                    {/* Tombol First Page */}
                    <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => table.setPageIndex(0)}
                        disabled={!table.getCanPreviousPage() || loading}
                    >
                        <ChevronsLeft className="w-4 h-4" />
                    </Button>

                    {/* Tombol Previous Page */}
                    <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => table.previousPage()}
                        disabled={!table.getCanPreviousPage() || loading}
                    >
                        <ChevronLeft className="w-4 h-4" />
                    </Button>

                    {/* Tombol Next Page */}
                    <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => table.nextPage()}
                        disabled={!table.getCanNextPage() || loading}
                    >
                        <ChevronRight className="w-4 h-4" />
                    </Button>

                    {/* Tombol Last Page */}
                    <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => table.setPageIndex(table.getPageCount() - 1)}
                        disabled={!table.getCanNextPage() || loading}
                    >
                        <ChevronsRight className="w-4 h-4" />
                    </Button>
                </div>
            </div>
        </div>
    )
}