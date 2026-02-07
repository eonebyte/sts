"use client"

import * as React from "react"
import { format } from "date-fns"
import { id } from "date-fns/locale"
import { Calendar as CalendarIcon, X } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover"

interface DateRangeFilterProps {
    dateRange: { from: Date | undefined; to: Date | undefined }
    setDateRange: (range: { from: Date | undefined; to: Date | undefined }) => void
    handleResetFilter: () => void
}

export function DateRangeFilter({
    dateRange,
    setDateRange,
    handleResetFilter,
}: DateRangeFilterProps) {

    return (
        <div className="flex items-center gap-2">
            {/* Date Range Picker */}
            <Popover>
                <PopoverTrigger asChild>
                    <Button
                        variant="outline"
                        className={cn(
                            "justify-start text-left font-normal min-w-[280px]",
                            !dateRange.from && !dateRange.to && "text-muted-foreground"
                        )}
                    >
                        <CalendarIcon className="mr-2 h-4 w-4" />
                        {dateRange.from ? (
                            dateRange.to ? (
                                <>
                                    {format(dateRange.from, "dd MMM yyyy", { locale: id })} -{" "}
                                    {format(dateRange.to, "dd MMM yyyy", { locale: id })}
                                </>
                            ) : (
                                format(dateRange.from, "dd MMM yyyy", { locale: id })
                            )
                        ) : (
                            <span>Pilih periode</span>
                        )}
                    </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="end">
                    <Calendar
                        mode="range"
                        defaultMonth={dateRange.from}
                        captionLayout="dropdown"
                        selected={{
                            from: dateRange.from,
                            to: dateRange.to,
                        }}
                        onSelect={(range) => {
                            setDateRange({
                                from: range?.from,
                                to: range?.to,
                            });
                        }}
                        numberOfMonths={2}
                        locale={id}
                    />
                </PopoverContent>
            </Popover>

            {/* Reset Filter Button */}
            {(dateRange.from || dateRange.to) && (
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleResetFilter}
                    className="h-9 px-2"
                >
                    <X className="h-4 w-4" />
                </Button>
            )}
        </div>
    )
}