'use client';

import {
  Package, Truck, CheckCircle2, Clock, BarChart3, ArrowRight,
  MapPin, Calendar, TrendingUp, ChevronDown, Users
} from "lucide-react";
import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
// Import Recharts components
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip,
  ResponsiveContainer, PieChart, Pie, Cell, Legend
} from 'recharts';

export default function DashboardPage() {
  const [expandedVehicle, setExpandedVehicle] = useState<string | null>(null);
  const [showDetailView, setShowDetailView] = useState(false);

  // --- DATA EXISTING ---
  const vehiclesWithCustomers = [
    { vehicle: "Truck-001", totalTrips: 4, drivers: [{ driver: "Budi Santoso", trips: 2, customers: ["PT Maju Sejahtera", "CV Bersama Jaya"] }, { driver: "Rudi Hermawan", trips: 2, customers: ["Toko Sinar Abadi", "PT Makmur Sentosa"] }] },
    { vehicle: "Truck-002", totalTrips: 3, drivers: [{ driver: "Ahmad Wijaya", trips: 3, customers: ["PT Cahaya Nusantara", "CV Sukses Bersama", "Toko Rajawali"] }] },
    { vehicle: "Truck-003", totalTrips: 10, drivers: [{ driver: "Eka Prasetya", trips: 4, customers: ["PT Global Logistik", "CV Terpercaya", "PT Maju Bersama", "Toko Sentosa"] }, { driver: "Bambang Suryanto", trips: 3, customers: ["CV Jaya Mandiri", "PT Ekspor Jaya", "Toko Baru"] }, { driver: "Dwi Handoko", trips: 3, customers: ["PT Mitra Global", "CV Sukses Niaga", "Toko Makmur"] }] },
    { vehicle: "Truck-004", totalTrips: 5, drivers: [{ driver: "Hendra Kusuma", trips: 5, customers: ["PT Sentosa Jaya", "CV Mitra Setia", "Toko Rajasa", "PT Ekspor Maju", "CV Dinamis"] }] },
    { vehicle: "Truck-005", totalTrips: 1, drivers: [{ driver: "Siti Nurhaliza", trips: 1, customers: ["PT Inovasi Mandiri"] }] }
  ];

  const shipmentStatus = [
    { title: "SJ in Delivery", value: "12", subtitle: "Belum handover", icon: Package, color: "bg-blue-500", lightColor: "bg-blue-100", textColor: "text-blue-600", hex: "#3b82f6" },
    { title: "At DPK", value: "28", subtitle: "Belum ke driver", icon: MapPin, color: "bg-purple-500", lightColor: "bg-purple-100", textColor: "text-purple-600", hex: "#a855f7" },
    { title: "Driver", value: "45", subtitle: "In Transit", icon: Truck, color: "bg-amber-500", lightColor: "bg-amber-100", textColor: "text-amber-600", hex: "#f59e0b" },
    { title: "At Customer", value: "18", subtitle: "Driver checkout", icon: CheckCircle2, color: "bg-green-500", lightColor: "bg-green-100", textColor: "text-green-600", hex: "#22c55e" },
    { title: "To Marketing", value: "8", subtitle: "Belum ke marketing", icon: TrendingUp, color: "bg-orange-500", lightColor: "bg-orange-100", textColor: "text-orange-600", hex: "#f97316" },
    { title: "To FAT", value: "3", subtitle: "Belum ke FAT", icon: Clock, color: "bg-red-500", lightColor: "bg-red-100", textColor: "text-red-600", hex: "#ef4444" }
  ];

  // --- CHART LOGIC ---
  const barData = vehiclesWithCustomers.map(v => ({ name: v.vehicle, rit: v.totalTrips }));
  const pieData = shipmentStatus.map(s => ({ name: s.title, value: parseInt(s.value), color: s.hex }));

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-white to-slate-50 p-6">
      <div className="max-w-7xl mx-auto space-y-6">

        {/* Header (Original) */}
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-3xl font-bold text-slate-900 flex items-center gap-3">
              <div className="p-2 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg">
                <Truck className="w-6 h-6 text-white" />
              </div>
              STS Dashboard
            </h1>
            <p className="text-slate-600 mt-1">Shipment Tracking System - Ringkasan Operasional Real-time</p>
          </div>
          <div className="text-right">
            <p className="text-slate-600 text-sm">Updated: {new Date().toLocaleString('id-ID')}</p>
          </div>
        </div>

        {/* Shipment Status Flow (Original) */}
        <div>
          <h2 className="text-lg font-semibold text-slate-900 mb-3 flex items-center gap-2">
            <Package className="w-5 h-5 text-blue-500" />
            Status Shipment Flow
          </h2>

          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            {shipmentStatus.map((status, idx) => (
              <div key={idx} className="flex items-center gap-2">
                <Card className="flex-1 p-0 min-w-0 bg-white border-slate-200 hover:border-blue-400 transition-colors shadow-sm">
                  <CardContent className="p-2.5">
                    <div className="flex items-center justify-between gap-3">
                      <div className="flex items-center gap-2 min-w-0">
                        <div className={`p-1.5 rounded-md ${status.lightColor || 'bg-slate-100'} flex-shrink-0`}>
                          <status.icon className={`w-5 h-5 ${status.textColor || 'text-slate-600'}`} />
                        </div>
                        <div className="min-w-0">
                          <p className="text-[14px] font-bold text-slate-700 leading-none truncate">{status.title}</p>
                          <p className="text-[12px] text-slate-400 truncate mt-1 leading-none">{status.subtitle}</p>
                        </div>
                      </div>
                      <div className="text-2xl font-black text-slate-900 leading-none">{status.value}</div>
                    </div>
                  </CardContent>
                </Card>
                <div className="hidden lg:flex items-center justify-center">
                  {(idx + 1) % 4 !== 0 && <ArrowRight className="w-3.5 h-3.5 text-slate-300 -mr-2" />}
                </div>
                <div className="flex lg:hidden items-center justify-center">
                  {(idx + 1) % 2 !== 0 && <ArrowRight className="w-3.5 h-3.5 text-slate-300 -mr-2" />}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* --- SECTION GRAFIK MODERN (SHADCN STYLE) --- */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">

          {/* Bar Chart - Productivity */}
          <Card className="lg:col-span-2 shadow-sm border-slate-200 overflow-hidden bg-white/50 backdrop-blur-sm">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-7">
              <div className="space-y-1">
                <CardTitle className="text-base font-semibold tracking-tight">
                  Kinerja Kendaraan
                </CardTitle>
                <p className="text-[13px] text-slate-500">Total ritase berdasarkan unit armada bulan ini</p>
              </div>
              <BarChart3 className="h-4 w-4 text-slate-400" />
            </CardHeader>
            <CardContent className="h-[300px] pl-0"> {/* pl-0 agar label Y dekat dengan border */}
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={barData} margin={{ top: 0, right: 30, left: 10, bottom: 0 }}>
                  <defs>
                    <linearGradient id="barPrimary" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="#0f172a" stopOpacity={0.9} />
                      <stop offset="100%" stopColor="#334155" stopOpacity={1} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
                  <XAxis
                    dataKey="name"
                    axisLine={false}
                    tickLine={false}
                    fontSize={11}
                    tick={{ fill: '#94a3b8' }}
                    dy={10}
                  />
                  <YAxis
                    axisLine={false}
                    tickLine={false}
                    fontSize={11}
                    tick={{ fill: '#94a3b8' }}
                  />
                  <Tooltip
                    cursor={{ fill: '#f1f5f9', radius: 4 }}
                    content={({ active, payload }) => {
                      if (active && payload && payload.length) {
                        return (
                          <div className="rounded-lg border bg-white p-2 shadow-md outline-none animate-in fade-in zoom-in-95">
                            <div className="grid grid-cols-2 gap-2">
                              <div className="flex flex-col">
                                <span className="text-[10px] uppercase text-slate-500 font-bold">Armada</span>
                                <span className="font-bold text-slate-900 text-xs">{payload[0].payload.name}</span>
                              </div>
                              <div className="flex flex-col border-l pl-2">
                                <span className="text-[10px] uppercase text-slate-500 font-bold">Total Rit</span>
                                <span className="font-bold text-blue-600 text-xs">{payload[0].value}</span>
                              </div>
                            </div>
                          </div>
                        );
                      }
                      return null;
                    }}
                  />
                  <Bar dataKey="rit" fill="url(#barPrimary)" radius={[4, 4, 0, 0]} barSize={32} />
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          {/* Pie Chart - Distribution */}
          <Card className="shadow-sm border-slate-200 bg-white/50 backdrop-blur-sm">
            <CardHeader className="space-y-1 pb-7">
              <CardTitle className="text-base font-semibold tracking-tight">
                Distribusi Status
              </CardTitle>
              <p className="text-[13px] text-slate-500">Persentase beban kerja saat ini</p>
            </CardHeader>
            <CardContent className="h-[300px] flex flex-col justify-center relative">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={pieData}
                    innerRadius={75}
                    outerRadius={95}
                    paddingAngle={4}
                    dataKey="value"
                    strokeWidth={0}
                  >
                    {pieData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip
                    content={({ active, payload }) => {
                      if (active && payload && payload.length) {
                        return (
                          <div className="rounded-md border bg-white px-2 py-1.5 shadow-sm text-xs font-medium text-slate-900">
                            {payload[0].name}: <span className="font-bold ml-1">{payload[0].value}</span>
                          </div>
                        );
                      }
                      return null;
                    }}
                  />
                </PieChart>
              </ResponsiveContainer>
              {/* Info Center of Pie */}
              <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none mt-4">
                <span className="text-3xl font-black text-slate-900">
                  {shipmentStatus.reduce((acc, curr) => acc + parseInt(curr.value), 0)}
                </span>
                <span className="text-[10px] uppercase font-bold text-slate-400">Total SJ</span>
              </div>
              {/* Legend Manual agar lebih Shadcn */}
              <div className="grid grid-cols-2 gap-x-2 gap-y-1 mt-4">
                {pieData.map((item, i) => (
                  <div key={i} className="flex items-center gap-1.5">
                    <div className="w-2 h-2 rounded-full" style={{ backgroundColor: item.color }} />
                    <span className="text-[10px] text-slate-600 font-medium truncate uppercase">{item.name}</span>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* All Vehicles Detail View (Original) */}
        <div>
          <h2 className="text-lg font-semibold text-slate-900 mb-4 flex items-center gap-2">
            <Truck className="w-5 h-5 text-blue-500" />
            Detail Rit & Customer
          </h2>
          <Card className="bg-white border-slate-200 shadow-sm">
            <CardContent className="p-0">
              <div className="divide-y divide-slate-200">
                {vehiclesWithCustomers.map((vehicle, idx) => (
                  <div key={idx}>
                    <button
                      onClick={() => setExpandedVehicle(expandedVehicle === vehicle.vehicle ? null : vehicle.vehicle)}
                      className="w-full px-6 py-4 hover:bg-slate-50 transition-colors flex items-center justify-between"
                    >
                      <div className="flex items-center gap-4 flex-1 text-left">
                        <div className="p-2 bg-blue-100 rounded-lg">
                          <Truck className="w-5 h-5 text-blue-500" />
                        </div>
                        <div className="flex-1">
                          <div className="text-slate-900 font-bold text-sm">{vehicle.vehicle}</div>
                          <div className="text-slate-500 text-xs">{vehicle.drivers.length} Driver Aktif</div>
                        </div>
                        <div className="flex items-center gap-8 mr-4">
                          <div className="text-center">
                            <div className="text-blue-600 font-bold text-lg">{vehicle.totalTrips}</div>
                            <div className="text-slate-500 text-[10px] uppercase font-bold">Rit</div>
                          </div>
                          <div className="text-center">
                            <div className="text-purple-600 font-bold text-lg">
                              {vehicle.drivers.reduce((acc, d) => acc + d.customers.length, 0)}
                            </div>
                            <div className="text-slate-500 text-[10px] uppercase font-bold">Cust</div>
                          </div>
                        </div>
                      </div>
                      <ChevronDown className={`w-5 h-5 text-slate-400 transition-transform ${expandedVehicle === vehicle.vehicle ? 'rotate-180' : ''}`} />
                    </button>

                    {expandedVehicle === vehicle.vehicle && (
                      <div className="bg-slate-50 p-4 border-t border-slate-200">
                        <div className="space-y-3">
                          {vehicle.drivers.map((driverData, didx) => (
                            <div key={didx} className="bg-white rounded-xl border border-slate-200 overflow-hidden shadow-sm">
                              <div className="grid grid-cols-1 md:grid-cols-3 divide-y md:divide-y-0 md:divide-x divide-slate-100">
                                <div className="p-4 bg-white">
                                  <div className="flex items-center gap-2 mb-2">
                                    <span className="text-[10px] bg-blue-100 text-blue-700 font-bold px-2 py-0.5 rounded-full uppercase">Driver {didx + 1}</span>
                                  </div>
                                  <div className="text-slate-900 font-bold text-sm mb-1">{driverData.driver}</div>
                                  <div className="flex gap-3 text-xs">
                                    <span className="text-slate-500">Rit: <b className="text-blue-600">{driverData.trips}</b></span>
                                    <span className="text-slate-500">Customer: <b className="text-purple-600">{driverData.customers.length}</b></span>
                                  </div>
                                </div>
                                <div className="p-4 md:col-span-2 bg-slate-50/30">
                                  <div className="text-slate-500 text-[10px] font-bold uppercase tracking-wider mb-2 flex items-center gap-1">
                                    <MapPin className="w-3 h-3 text-amber-500" /> Customer
                                  </div>
                                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                                    {driverData.customers.map((customer, cidx) => (
                                      <div key={cidx} className="flex items-center gap-2 text-xs text-slate-700 p-2 rounded-lg bg-white border border-slate-100 shadow-sm">
                                        <span className="bg-slate-100 text-slate-500 w-5 h-5 flex items-center justify-center rounded text-[10px] font-bold">{cidx + 1}</span>
                                        <span className="font-medium truncate">{customer}</span>
                                      </div>
                                    ))}
                                  </div>
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
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