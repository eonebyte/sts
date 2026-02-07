"use client"

import * as React from "react"
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
import { ChevronDown, ChevronRight, Loader2, Search } from "lucide-react"
import { Checkbox } from "@/components/ui/checkbox"
import dayjs from "dayjs"
import 'dayjs/locale/id'
import { XCircle } from "lucide-react"

interface GroupedData {
    spp_no: string
    customer_name: string
    items: any[]
}

interface DataTableGroupProps {
    data: any[]
    loading?: boolean
    rowSelection?: any
    setRowSelection?: React.Dispatch<React.SetStateAction<any>>
}

export function DataTableGroup({
    data = [],
    loading = false,
    rowSelection = {},
    setRowSelection,
}: DataTableGroupProps) {
    const [expandedGroups, setExpandedGroups] = React.useState<Set<string>>(new Set())
    const [filters, setFilters] = React.useState({
        spp_no: "",
        customer_name: "",
    })
    const [childFilters, setChildFilters] = React.useState<{ [key: string]: { document_no: string, movement_date: string } }>({})

    // Get child filter for a specific group
    const getChildFilter = (sppNo: string) => {
        return childFilters[sppNo] || { document_no: "", movement_date: "" }
    }

    // Set child filter for a specific group
    const setChildFilter = (sppNo: string, field: string, value: string) => {
        setChildFilters(prev => ({
            ...prev,
            [sppNo]: {
                ...getChildFilter(sppNo),
                [field]: value
            }
        }))
    }

    // Filter child items based on child filters
    const getFilteredItems = (items: any[], sppNo: string) => {
        const filter = getChildFilter(sppNo)
        return items.filter(item => {
            const docMatch = item.document_no.toLowerCase().includes(filter.document_no.toLowerCase())
            const dateMatch = filter.movement_date === "" ||
                (item.movement_date && dayjs(item.movement_date).locale('id').format('DD MMM YYYY').toLowerCase().includes(filter.movement_date.toLowerCase()))
            return docMatch && dateMatch
        })
    }

    // Group data by SPP
    const groupedData = React.useMemo(() => {
        const groups: { [key: string]: GroupedData } = {}

        data.forEach(item => {
            const key = item.spp_no || "NO_SPP"
            if (!groups[key]) {
                groups[key] = {
                    spp_no: item.spp_no || "-",
                    customer_name: item.customer_name || "-",
                    items: []
                }
            }
            groups[key].items.push(item)
        })

        return Object.values(groups)
    }, [data])

    // Filter grouped data
    const filteredGroupedData = React.useMemo(() => {
        return groupedData.filter(group => {
            const sppMatch = group.spp_no.toLowerCase().includes(filters.spp_no.toLowerCase())
            const customerMatch = group.customer_name.toLowerCase().includes(filters.customer_name.toLowerCase())
            return sppMatch && customerMatch
        })
    }, [groupedData, filters])

    const toggleGroup = (sppNo: string) => {
        setExpandedGroups(prev => {
            const newSet = new Set(prev)
            if (newSet.has(sppNo)) {
                newSet.delete(sppNo)
            } else {
                newSet.add(sppNo)
            }
            return newSet
        })
    }

    const toggleAllInGroup = (items: any[], checked: boolean) => {
        if (!setRowSelection) return

        setRowSelection((prev: any) => {
            const newSelection = { ...prev }
            items.forEach(item => {
                const index = data.findIndex(d => d.m_inout_id === item.m_inout_id)
                if (index !== -1) {
                    if (checked) {
                        newSelection[index] = true
                    } else {
                        delete newSelection[index]
                    }
                }
            })
            return newSelection
        })
    }

    const isGroupSelected = (items: any[]) => {
        return items.every(item => {
            const index = data.findIndex(d => d.m_inout_id === item.m_inout_id)
            return rowSelection[index]
        })
    }

    const isGroupPartiallySelected = (items: any[]) => {
        const selectedCount = items.filter(item => {
            const index = data.findIndex(d => d.m_inout_id === item.m_inout_id)
            return rowSelection[index]
        }).length
        return selectedCount > 0 && selectedCount < items.length
    }

    const toggleRow = (item: any) => {
        if (!setRowSelection) return

        const index = data.findIndex(d => d.m_inout_id === item.m_inout_id)
        if (index !== -1) {
            setRowSelection((prev: any) => {
                const newSelection = { ...prev }
                if (newSelection[index]) {
                    delete newSelection[index]
                } else {
                    newSelection[index] = true
                }
                return newSelection
            })
        }
    }

    const isRowSelected = (item: any) => {
        const index = data.findIndex(d => d.m_inout_id === item.m_inout_id)
        return rowSelection[index] || false
    }

    if (loading) {
        return (
            <div className="rounded-md border bg-white overflow-hidden shadow-sm">
                <div className="h-32 flex items-center justify-center">
                    <div className="flex flex-col items-center justify-center gap-2">
                        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
                        <p className="text-sm text-slate-500 font-medium">Memuat data...</p>
                    </div>
                </div>
            </div>
        )
    }

    if (groupedData.length === 0) {
        return (
            <div className="rounded-md border bg-white overflow-hidden shadow-sm">
                <div className="h-24 flex items-center justify-center text-slate-500 italic">
                    Tidak ada data yang ditemukan.
                </div>
            </div>
        )
    }

    if (filteredGroupedData.length === 0) {
        return (
            <div className="rounded-md border bg-white overflow-hidden shadow-sm">
                <div className="overflow-x-auto">
                    <Table>
                        <TableHeader className="bg-slate-50">
                            <TableRow className="hover:bg-transparent">
                                <TableHead className="w-12 py-3 px-2">
                                    <Checkbox disabled />
                                </TableHead>
                                <TableHead className="w-16 py-3 px-2">
                                    <div className="flex flex-col gap-2">
                                        <div className="text-[11px] uppercase tracking-wider font-bold text-slate-700">
                                            No
                                        </div>
                                        <div className="h-7" />
                                    </div>
                                </TableHead>
                                <TableHead className="py-3 px-2">
                                    <div className="flex flex-col gap-2">
                                        <div className="text-[11px] uppercase tracking-wider font-bold text-slate-700">
                                            SPP No
                                        </div>
                                        <div className="relative">
                                            <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3 w-3 text-slate-400" />
                                            <Input
                                                placeholder="Filter SPP..."
                                                value={filters.spp_no}
                                                onChange={(e) => setFilters(prev => ({ ...prev, spp_no: e.target.value }))}
                                                className="h-7 pl-7 pr-2 text-[10px] w-full min-w-[120px] bg-white border-slate-200 focus-visible:ring-blue-500"
                                            />
                                        </div>
                                    </div>
                                </TableHead>
                                <TableHead className="py-3 px-2">
                                    <div className="flex flex-col gap-2">
                                        <div className="text-[11px] uppercase tracking-wider font-bold text-slate-700">
                                            Customer
                                        </div>
                                        <div className="relative">
                                            <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3 w-3 text-slate-400" />
                                            <Input
                                                placeholder="Filter Customer..."
                                                value={filters.customer_name}
                                                onChange={(e) => setFilters(prev => ({ ...prev, customer_name: e.target.value }))}
                                                className="h-7 pl-7 pr-2 text-[10px] w-full min-w-[120px] bg-white border-slate-200 focus-visible:ring-blue-500"
                                            />
                                        </div>
                                    </div>
                                </TableHead>
                                <TableHead className="w-24 py-3 px-2">
                                    <div className="flex flex-col gap-2">
                                        <div className="text-[11px] uppercase tracking-wider font-bold text-slate-700">
                                            Items
                                        </div>
                                        <div className="h-7" />
                                    </div>
                                </TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            <TableRow>
                                <TableCell colSpan={5} className="h-24 text-center text-slate-500 italic">
                                    Tidak ada data yang sesuai dengan filter.
                                </TableCell>
                            </TableRow>
                        </TableBody>
                    </Table>
                </div>
            </div>
        )
    }

    return (
        <div className="space-y-4">
            <div className="rounded-md border bg-white overflow-hidden shadow-sm">
                <div className="overflow-x-auto">
                    <Table>
                        <TableHeader className="bg-slate-50">
                            <TableRow className="hover:bg-transparent">
                                <TableHead className="w-12 py-3 px-2">
                                    <Checkbox disabled />
                                </TableHead>
                                <TableHead className="w-16 py-3 px-2">
                                    <div className="flex flex-col gap-2">
                                        <div className="text-[11px] uppercase tracking-wider font-bold text-slate-700">
                                            No
                                        </div>
                                        <div className="h-7" /> {/* Spacer */}
                                    </div>
                                </TableHead>
                                <TableHead className="py-3 px-2">
                                    <div className="flex flex-col gap-2">
                                        <div className="text-[11px] uppercase tracking-wider font-bold text-slate-700">
                                            SPP No
                                        </div>
                                        <div className="relative">
                                            <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3 w-3 text-slate-400" />
                                            <Input
                                                placeholder="Filter SPP..."
                                                value={filters.spp_no}
                                                onChange={(e) => setFilters(prev => ({ ...prev, spp_no: e.target.value }))}
                                                className="h-7 pl-7 pr-2 text-[10px] w-full min-w-[120px] bg-white border-slate-200 focus-visible:ring-blue-500"
                                            />
                                        </div>
                                    </div>
                                </TableHead>
                                <TableHead className="py-3 px-2">
                                    <div className="flex flex-col gap-2">
                                        <div className="text-[11px] uppercase tracking-wider font-bold text-slate-700">
                                            Customer
                                        </div>
                                        <div className="relative">
                                            <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3 w-3 text-slate-400" />
                                            <Input
                                                placeholder="Filter Customer..."
                                                value={filters.customer_name}
                                                onChange={(e) => setFilters(prev => ({ ...prev, customer_name: e.target.value }))}
                                                className="h-7 pl-7 pr-2 text-[10px] w-full min-w-[120px] bg-white border-slate-200 focus-visible:ring-blue-500"
                                            />
                                        </div>
                                    </div>
                                </TableHead>
                                <TableHead className="w-24 py-3 px-2">
                                    <div className="flex flex-col gap-2">
                                        <div className="text-[11px] uppercase tracking-wider font-bold text-slate-700">
                                            Items
                                        </div>
                                        <div className="h-7" /> {/* Spacer */}
                                    </div>
                                </TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {filteredGroupedData.map((group, groupIndex) => (
                                <React.Fragment key={group.spp_no}>
                                    {/* Parent Row */}
                                    <TableRow className="hover:bg-slate-50/50 transition-colors border-b bg-slate-50/30">
                                        <TableCell className="px-2 py-2">
                                            <Checkbox
                                                checked={isGroupSelected(group.items)}
                                                onCheckedChange={(checked) => toggleAllInGroup(group.items, !!checked)}
                                                aria-label="Select group"
                                                ref={(el) => {
                                                    if (el) {
                                                        el.indeterminate = isGroupPartiallySelected(group.items)
                                                    }
                                                }}
                                            />
                                        </TableCell>
                                        <TableCell className="px-2 py-2">
                                            <div className="w-4 font-medium">{groupIndex + 1}</div>
                                        </TableCell>
                                        <TableCell className="px-2 py-2">
                                            <div className="font-bold text-blue-600">{group.spp_no}</div>
                                        </TableCell>
                                        <TableCell className="px-2 py-2">
                                            <div className="font-medium text-slate-700">{group.customer_name}</div>
                                        </TableCell>
                                        <TableCell className="px-2 py-2">
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                className="h-7 px-2 text-xs"
                                                onClick={() => toggleGroup(group.spp_no)}
                                            >
                                                {expandedGroups.has(group.spp_no) ? (
                                                    <ChevronDown className="w-4 h-4 mr-1" />
                                                ) : (
                                                    <ChevronRight className="w-4 h-4 mr-1" />
                                                )}
                                                {group.items.length} Items
                                            </Button>
                                        </TableCell>
                                    </TableRow>

                                    {/* Child Rows */}
                                    {expandedGroups.has(group.spp_no) && (
                                        <>
                                            {/* Child Header */}
                                            <TableRow className="bg-slate-100 border-b">
                                                <TableHead className="py-2 px-2 pl-12"></TableHead>
                                                <TableHead className="py-2 px-2">
                                                    <div className="flex flex-col gap-2">
                                                        <div className="text-[10px] uppercase tracking-wider font-bold text-slate-600">
                                                            No
                                                        </div>
                                                        <div className="h-7" /> {/* Spacer */}
                                                    </div>
                                                </TableHead>
                                                <TableHead className="py-2 px-2">
                                                    <div className="flex flex-col gap-2">
                                                        <div className="text-[10px] uppercase tracking-wider font-bold text-slate-600">
                                                            Document No
                                                        </div>
                                                        <div className="relative">
                                                            <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3 w-3 text-slate-400" />
                                                            <Input
                                                                placeholder="Filter..."
                                                                value={getChildFilter(group.spp_no).document_no}
                                                                onChange={(e) => setChildFilter(group.spp_no, "document_no", e.target.value)}
                                                                className="h-7 pl-7 pr-2 text-[10px] w-full min-w-[100px] bg-white border-slate-200 focus-visible:ring-blue-500"
                                                            />
                                                        </div>
                                                    </div>
                                                </TableHead>
                                                <TableHead className="py-2 px-2">
                                                    <div className="flex flex-col gap-2">
                                                        <div className="text-[10px] uppercase tracking-wider font-bold text-slate-600">
                                                            Movement Date
                                                        </div>
                                                        <div className="relative">
                                                            <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3 w-3 text-slate-400" />
                                                            <Input
                                                                placeholder="Filter..."
                                                                value={getChildFilter(group.spp_no).movement_date}
                                                                onChange={(e) => setChildFilter(group.spp_no, "movement_date", e.target.value)}
                                                                className="h-7 pl-7 pr-2 text-[10px] w-full min-w-[100px] bg-white border-slate-200 focus-visible:ring-blue-500"
                                                            />
                                                        </div>
                                                    </div>
                                                </TableHead>
                                                <TableHead className="py-2 px-2">
                                                    <div className="flex flex-col gap-2">
                                                        <div className="text-[10px] uppercase tracking-wider font-bold text-slate-600">
                                                            Aksi
                                                        </div>
                                                        <div className="h-7" /> {/* Spacer */}
                                                    </div>
                                                </TableHead>
                                            </TableRow>

                                            {/* Child Data Rows */}
                                            {getFilteredItems(group.items, group.spp_no).length > 0 ? (
                                                getFilteredItems(group.items, group.spp_no).map((item, itemIndex) => (
                                                    <TableRow
                                                        key={item.m_inout_id}
                                                        className="hover:bg-blue-50/30 transition-colors border-b"
                                                        data-state={isRowSelected(item) && "selected"}
                                                    >
                                                        <TableCell className="px-2 py-1 pl-12">
                                                            <Checkbox
                                                                checked={isRowSelected(item)}
                                                                onCheckedChange={() => toggleRow(item)}
                                                                aria-label="Select row"
                                                            />
                                                        </TableCell>
                                                        <TableCell className="px-2 py-1">
                                                            <div className="text-xs">{itemIndex + 1}</div>
                                                        </TableCell>
                                                        <TableCell className="px-2 py-1">
                                                            <div className="text-xs font-bold text-slate-700">
                                                                {item.document_no}
                                                            </div>
                                                        </TableCell>
                                                        <TableCell className="px-2 py-1">
                                                            <div className="text-xs whitespace-nowrap">
                                                                {item.movement_date
                                                                    ? dayjs(item.movement_date).locale('id').format('DD MMM YYYY')
                                                                    : "-"
                                                                }
                                                            </div>
                                                        </TableCell>
                                                        <TableCell className="px-2 py-1">
                                                            {item.status === "HO: DEL_TO_MKT" && (
                                                                <Button
                                                                    variant="ghost"
                                                                    size="sm"
                                                                    className="h-7 text-xs text-red-600 hover:text-red-700 hover:bg-red-50"
                                                                    onClick={() => onCancel(item.m_inout_id, item.status)}
                                                                >
                                                                    <XCircle className="w-3 h-3 mr-1" />
                                                                    Reject
                                                                </Button>
                                                            )}
                                                        </TableCell>
                                                    </TableRow>
                                                ))
                                            ) : (
                                                <TableRow>
                                                    <TableCell colSpan={5} className="px-2 py-4 text-center text-xs text-slate-500 italic">
                                                        Tidak ada data yang sesuai dengan filter
                                                    </TableCell>
                                                </TableRow>
                                            )}
                                        </>
                                    )}
                                </React.Fragment>
                            ))}
                        </TableBody>
                    </Table>
                </div>
            </div>
        </div>
    )
}