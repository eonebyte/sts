'use client';

import {
  LayoutDashboard,
  Package,
  Truck,
  CheckCircle2,
  Clock,
  BarChart3
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function DashboardPage() {
  // Data dummy untuk sementara
  const stats = [
    {
      title: "Total Shipment",
      value: "1,284",
      icon: Package,
      color: "text-blue-600",
      bg: "bg-blue-100",
      description: "Total surat jalan bulan ini"
    },
    {
      title: "In Transit",
      value: "42",
      icon: Truck,
      color: "text-amber-600",
      bg: "bg-amber-100",
      description: "Sedang dalam perjalanan"
    },
    {
      title: "Completed",
      value: "1,150",
      icon: CheckCircle2,
      color: "text-emerald-600",
      bg: "bg-emerald-100",
      description: "Sudah terkirim & bundle"
    },
    {
      title: "Pending",
      value: "92",
      icon: Clock,
      color: "text-rose-600",
      bg: "bg-rose-100",
      description: "Menunggu diproses"
    },
  ];

  return (
    <div className="p-4 space-y-6">
      {/* Header */}
      <div className="flex items-center gap-2">
        <div className="p-2 bg-slate-800 rounded-lg">
          <LayoutDashboard className="w-5 h-5 text-white" />
        </div>
        <div>
          <h1 className="text-xl font-bold text-slate-800">Dashboard Summary</h1>
          <p className="text-xs text-slate-500">Ringkasan operasional shipment dan handover.</p>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat, index) => (
          <Card key={index} className="shadow-sm border-none bg-white">
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-xs font-medium text-slate-500 uppercase tracking-wider">
                {stat.title}
              </CardTitle>
              <div className={`p-2 rounded-md ${stat.bg}`}>
                <stat.icon className={`w-4 h-4 ${stat.color}`} />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-slate-800">{stat.value}</div>
              <p className="text-[10px] text-slate-400 mt-1">
                {stat.description}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Placeholder for Charts / Recent Activity */}
      <div className="grid gap-4 md:grid-cols-7">
        <Card className="col-span-4 shadow-sm border-none">
          <CardHeader>
            <CardTitle className="text-sm font-semibold flex items-center gap-2">
              <BarChart3 className="w-4 h-4 text-slate-400" />
              Shipment Trend (Coming Soon)
            </CardTitle>
          </CardHeader>
          <CardContent className="h-[300px] flex items-center justify-center border-2 border-dashed border-slate-100 rounded-lg m-4 mt-0">
            <div className="text-center">
              <p className="text-sm text-slate-400 italic">Grafik sedang dalam pengembangan...</p>
            </div>
          </CardContent>
        </Card>

        <Card className="col-span-3 shadow-sm border-none">
          <CardHeader>
            <CardTitle className="text-sm font-semibold">Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="flex items-center gap-3 pb-3 border-b border-slate-50 last:border-0">
                  <div className="w-2 h-2 rounded-full bg-blue-500" />
                  <div className="flex-1">
                    <p className="text-xs font-medium text-slate-700">SJ-860{i + 50} has been bundled</p>
                    <p className="text-[10px] text-slate-400">2 minutes ago</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}