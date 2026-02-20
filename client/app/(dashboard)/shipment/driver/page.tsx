'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { Loader2 } from "lucide-react";
import { DataTable } from "@/components/ui/data-table";
import { getColumns, Driver } from "./columns";
import { toast } from "sonner";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function ListDriver() {
    const { isAuthorized } = useAuth();
    const [drivers, setDrivers] = useState<Driver[]>([]);
    const [loading, setLoading] = useState(true);

    // State untuk Edit
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [selectedDriver, setSelectedDriver] = useState<Driver | null>(null);
    const [newName, setNewName] = useState("");
    const [newPassword, setNewPassword] = useState("");
    const [isUpdating, setIsUpdating] = useState(false);

    const fetchDrivers = async () => {
        const authToken = localStorage.getItem('token');
        setLoading(true);
        try {
            const res = await fetch(`${API_BASE_URL}/shipments/drivers`, {
                headers: { Authorization: `Bearer ${authToken}` },
            });
            const data = await res.json();
            setDrivers(Array.isArray(data.data) ? data.data : data.data?.data || []);
        } catch (error) {
            toast.error("Gagal mengambil data");
        } finally {
            setLoading(false);
        }
    };

    const handleEditClick = (driver: Driver) => {
        setSelectedDriver(driver);
        setNewName(driver.Name); // Isi field name otomatis
        setNewPassword("");      // Kosongkan password untuk keamanan
        setIsDialogOpen(true);
    };

    const handleUpdate = async () => {
        if (!selectedDriver) return;


        if (!newName && !newPassword) {
            toast.error("Setidaknya nama atau password harus diubah");
            return;
        }


        setIsUpdating(true);
        const authToken = localStorage.getItem('token');

        try {
            const payload = {
                driver_id: Number(selectedDriver.driver_by), // Pastikan ini ID numerik sesuai int64 di Go
                driver_name: newName || "",
                password: newPassword || "" // Kirim string kosong jika tidak diisi
            };

            const res = await fetch(`${API_BASE_URL}/shipments/drivers`, {
                method: 'PUT', // atau PATCH sesuai API Anda
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${authToken}`
                },
                body: JSON.stringify(payload),
            });


            if (res.ok) {
                toast.success("Driver berhasil diperbarui");
                setIsDialogOpen(false);
                fetchDrivers(); // Refresh data
            } else {
                throw new Error();
            }
        } catch (error) {
            toast.error("Gagal memperbarui data");
        } finally {
            setIsUpdating(false);
        }
    };

    useEffect(() => {
        if (isAuthorized) fetchDrivers();
    }, [isAuthorized]);

    if (!isAuthorized) return (
        <div className="h-screen flex items-center justify-center">
            <Loader2 className="animate-spin text-blue-600 w-8 h-8" />
        </div>
    );

    return (
        <div className="p-4 space-y-4">
            <div className="flex flex-col gap-1">
                <h1 className="text-2xl font-bold tracking-tight">List Driver</h1>
                <p className="text-muted-foreground text-sm">Daftar driver yang ada di system.</p>
            </div>

            <DataTable
                columns={getColumns({ onEdit: handleEditClick })}
                data={drivers}
                loading={loading}
            />

            {/* Modal Edit */}
            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Edit Driver</DialogTitle>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label htmlFor="name">Nama Driver</Label>
                            <Input
                                id="name"
                                value={newName}
                                onChange={(e) => setNewName(e.target.value)}
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="password">Password Baru (Opsional)</Label>
                            <Input
                                id="password"
                                type="password"
                                placeholder="Isi jika ingin ganti password"
                                value={newPassword}
                                onChange={(e) => setNewPassword(e.target.value)}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setIsDialogOpen(false)}>Batal</Button>
                        <Button onClick={handleUpdate} disabled={isUpdating}>
                            {isUpdating ? "Menyimpan..." : "Simpan Perubahan"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}