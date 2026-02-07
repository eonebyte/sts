'use client';

import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export default function DashboardSkeleton() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-white to-slate-50 p-6">
      <div className="max-w-7xl mx-auto space-y-6">

        {/* Header Skeleton */}
        <div className="flex items-start justify-between">
          <div className="space-y-3">
            <div className="flex items-center gap-3">
              <Skeleton className="w-12 h-12 rounded-lg" />
              <Skeleton className="h-9 w-48" />
            </div>
            <Skeleton className="h-4 w-96" />
          </div>
          <div className="text-right">
            <Skeleton className="h-4 w-40 ml-auto" />
          </div>
        </div>

        {/* Shipment Status Flow Skeleton */}
        <div>
          <div className="flex items-center gap-2 mb-3">
            <Skeleton className="w-5 h-5 rounded" />
            <Skeleton className="h-6 w-48" />
          </div>

          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            {[1, 2, 3, 4, 5, 6].map((idx) => (
              <div key={idx} className="flex items-center gap-2">
                <Card className="flex-1 p-0 min-w-0 bg-white border-slate-200 shadow-sm">
                  <CardContent className="p-2.5">
                    <div className="flex items-center justify-between gap-3">
                      <div className="flex items-center gap-2 min-w-0 flex-1">
                        <Skeleton className="w-9 h-9 rounded-md flex-shrink-0" />
                        <div className="space-y-2 flex-1 min-w-0">
                          <Skeleton className="h-4 w-24" />
                          <Skeleton className="h-3 w-20" />
                        </div>
                      </div>
                      <Skeleton className="h-8 w-8" />
                    </div>
                  </CardContent>
                </Card>
                {(idx + 1) % 4 !== 0 && idx < 6 && (
                  <div className="hidden lg:block">
                    <Skeleton className="w-3.5 h-3.5 -mr-2" />
                  </div>
                )}
                {(idx + 1) % 2 !== 0 && idx < 6 && (
                  <div className="block lg:hidden">
                    <Skeleton className="w-3.5 h-3.5 -mr-2" />
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Charts Section Skeleton */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">

          {/* Bar Chart Skeleton */}
          <Card className="lg:col-span-2 shadow-sm border-slate-200 overflow-hidden bg-white/50 backdrop-blur-sm">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-7">
              <div className="space-y-2 flex-1">
                <Skeleton className="h-5 w-40" />
                <Skeleton className="h-4 w-72" />
              </div>
              <Skeleton className="h-4 w-4" />
            </CardHeader>
            <CardContent className="h-[300px] pl-0">
              <div className="h-full flex items-end justify-between gap-4 px-8">
                {[1, 2, 3, 4, 5].map((i) => (
                  <div key={i} className="flex flex-col items-center gap-2 flex-1">
                    <Skeleton
                      className="w-full rounded-t"
                      // eslint-disable-next-line react-hooks/purity
                      style={{ height: `${Math.random() * 150 + 100}px` }}
                    />
                    <Skeleton className="h-3 w-16" />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* Pie Chart Skeleton */}
          <Card className="shadow-sm border-slate-200 bg-white/50 backdrop-blur-sm">
            <CardHeader className="space-y-2 pb-7">
              <Skeleton className="h-5 w-36" />
              <Skeleton className="h-4 w-48" />
            </CardHeader>
            <CardContent className="h-[300px] flex flex-col justify-center items-center">
              <Skeleton className="w-48 h-48 rounded-full mb-4" />
              <div className="grid grid-cols-2 gap-x-2 gap-y-2 w-full">
                {[1, 2, 3, 4, 5, 6].map((i) => (
                  <div key={i} className="flex items-center gap-1.5">
                    <Skeleton className="w-2 h-2 rounded-full" />
                    <Skeleton className="h-3 w-20" />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Detail Rit & Customer Skeleton */}
        <div>
          <div className="flex items-center gap-2 mb-4">
            <Skeleton className="w-5 h-5 rounded" />
            <Skeleton className="h-6 w-48" />
          </div>

          <Card className="bg-white border-slate-200 shadow-sm">
            <CardContent className="p-0">
              <div className="divide-y divide-slate-200">
                {[1, 2, 3, 4, 5].map((idx) => (
                  <div key={idx} className="px-6 py-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-4 flex-1">
                        <Skeleton className="w-10 h-10 rounded-lg" />
                        <div className="space-y-2 flex-1">
                          <Skeleton className="h-4 w-24" />
                          <Skeleton className="h-3 w-32" />
                        </div>
                        <div className="flex items-center gap-8 mr-4">
                          <div className="text-center space-y-1">
                            <Skeleton className="h-6 w-8 mx-auto" />
                            <Skeleton className="h-3 w-8" />
                          </div>
                          <div className="text-center space-y-1">
                            <Skeleton className="h-6 w-8 mx-auto" />
                            <Skeleton className="h-3 w-8" />
                          </div>
                        </div>
                      </div>
                      <Skeleton className="w-5 h-5" />
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>

      </div>
    </div>
  );
}