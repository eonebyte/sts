'use client';

import { useState, useEffect } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Spinner } from '@/components/ui/spinner';
import { apiService, SuratJalanItem } from '@/lib/api-service';
import { AlertCircle, CheckCircle2, ArrowLeft, Loader2, Info } from 'lucide-react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { useAuth } from '@/hooks/useAuth';

interface CheckedInCustomer {
    customer_id: string;
    customer_name: string;
    tnkb: string;
}

interface CheckOutFormProps {
    onBack: () => void;
    onError: (message: string) => void;
    onSuccess: () => void;
    checkedInCustomer: CheckedInCustomer;
}

type FormStep = 'select-shipments' | 'confirm' | 'success';

export function CheckOutForm({
    onBack,
    onError,
    onSuccess,
    checkedInCustomer,
}: CheckOutFormProps) {
    const { user } = useAuth();
    const [step, setStep] = useState<FormStep>('select-shipments');
    const [shipments, setShipments] = useState<SuratJalanItem[]>([]);
    const [selectedShipments, setSelectedShipments] = useState<Set<number>>(new Set());
    const [loading, setLoading] = useState(false);
    const [submitting, setSubmitting] = useState(false);

    useEffect(() => {
        const loadShipments = async () => {
            if (!user?.user_id) return;
            setLoading(true);
            try {
                const driverIdNum = parseInt(user.user_id);
                const customerIdNum = parseInt(checkedInCustomer.customer_id)
                const data = await apiService.getOnCustomerShipments(customerIdNum, driverIdNum);

                const customerShipments = data.filter(
                    (s) => s.customer_name === checkedInCustomer.customer_name
                );

                setShipments(customerShipments);

                // Default: Centang semua
                if (customerShipments.length > 0) {
                    setSelectedShipments(new Set(customerShipments.map(s => s.m_inout_id)));
                }
            } catch (error) {
                onError('Gagal memuat daftar pengiriman');
            } finally {
                setLoading(false);
            }
        };

        loadShipments();
    }, [checkedInCustomer, user]);

    const handleShipmentToggle = (id: number) => {
        const newSelected = new Set(selectedShipments);
        if (newSelected.has(id)) {
            newSelected.delete(id);
        } else {
            newSelected.add(id);
        }
        setSelectedShipments(newSelected);
    };

    const handleSelectAll = () => {
        if (selectedShipments.size === shipments.length) {
            setSelectedShipments(new Set());
        } else {
            setSelectedShipments(new Set(shipments.map((s) => s.m_inout_id)));
        }
    };

    const handleSubmitCheckOut = async () => {
        if (!user?.user_id) return;

        setSubmitting(true);
        try {
            const driverIdNum = parseInt(user.user_id);

            // Cari shipment untuk mendapatkan TNKB jika ada yang dipilih
            // Jika tidak ada, kita harus pastikan TNKB tetap terkirim (dari data awal)
            const sampleShipment = shipments.length > 0 ? shipments[0] : null;

            const payload = {
                m_inout_ids: Array.from(selectedShipments),
                status: 'HO: DRIVER_CHECKOUT',
                user_id: driverIdNum,
                driver_by: driverIdNum,

                customer_id: parseInt(checkedInCustomer.customer_id),

                notes: selectedShipments.size === 0
                    ? `Driver Check-Out (SJ ditunda di customer)`
                    : `Driver Check-Out`,

                // Ambil TNKB_ID dari sample shipment atau 0 jika benar-benar kosong
                tnkb_id: sampleShipment?.tnkb_id ? parseInt(sampleShipment.tnkb_id) : 0,
            }

            const response = await apiService.processHandover(payload);

            if (response.success) {
                setStep('success');
            } else {
                onError(response.message || 'Check-out gagal');
            }
        } catch (error) {
            console.error(error);
            onError('Terjadi kesalahan koneksi');
        } finally {
            setSubmitting(false);
        }
    };

    const selectedCount = selectedShipments.size;
    const allSelected = selectedCount === shipments.length && shipments.length > 0;

    return (
        <div className="space-y-4">
            {step !== 'success' && (
                <Button variant="ghost" onClick={onBack} className="p-0 hover:bg-transparent text-slate-500">
                    <ArrowLeft className="w-4 h-4 mr-2" /> Kembali
                </Button>
            )}

            {step === 'select-shipments' && (
                <Card className="p-6">
                    <h2 className="text-lg font-bold mb-1">Pilih Surat Jalan yang akan dibawa pulang</h2>
                    <p className="text-1xl text-slate-500 mb-6">Customer : {checkedInCustomer.customer_name}</p>

                    {loading ? (
                        <div className="flex flex-col items-center py-10">
                            <Loader2 className="w-8 h-8 animate-spin text-blue-600 mb-2" />
                            <p className="text-sm text-slate-400">Memuat data...</p>
                        </div>
                    ) : (
                        <div className="space-y-4">
                            {shipments.length > 0 ? (
                                <>
                                    <div className="flex items-center space-x-2 pb-2 border-b">
                                        <Checkbox id="all" checked={allSelected} onCheckedChange={handleSelectAll} />
                                        <Label htmlFor="all" className="text-sm font-bold">Pilih Semua ({shipments.length})</Label>
                                    </div>

                                    <div className="max-h-[300px] overflow-y-auto space-y-3 pr-2">
                                        {shipments.map((s) => (
                                            <div key={s.m_inout_id} className="flex items-center space-x-3 p-3 bg-slate-50 rounded-lg border border-transparent hover:border-blue-200">
                                                <Checkbox
                                                    id={s.m_inout_id.toString()}
                                                    checked={selectedShipments.has(s.m_inout_id)}
                                                    onCheckedChange={() => handleShipmentToggle(s.m_inout_id)}
                                                />
                                                <div className="grid gap-0.5">
                                                    <Label htmlFor={s.m_inout_id.toString()} className="font-semibold">{s.document_no}</Label>
                                                    {/* <p className="text-[10px] text-slate-500 uppercase">{s.status}</p> */}
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                </>
                            ) : (
                                <Alert className="bg-blue-50 border-blue-200">
                                    <Info className="h-4 w-4 text-blue-600" />
                                    <AlertDescription className="text-blue-800">
                                        Tidak ada Surat Jalan aktif. Anda dapat langsung melakukan Check-Out.
                                    </AlertDescription>
                                </Alert>
                            )}

                            <Button
                                className="w-full bg-blue-600 py-6 h-auto flex flex-col gap-1"
                                onClick={() => setStep('confirm')}
                            >
                                <span className="text-base">Lanjut Check-Out</span>
                                <span className="text-[10px] opacity-80">
                                    {selectedCount === 0 ? "Tanpa membawa Surat Jalan" : `Membawa ${selectedCount} Surat Jalan`}
                                </span>
                            </Button>
                        </div>
                    )}
                </Card>
            )}

            {step === 'confirm' && (
                <Card className="p-6">
                    <h2 className="text-lg font-bold mb-4">Konfirmasi Keluar</h2>
                    <div className="space-y-4 bg-slate-50 p-4 rounded-xl mb-6">
                        <div>
                            <p className="text-[10px] text-slate-500 font-bold uppercase">Customer</p>
                            <p className="font-semibold">{checkedInCustomer.customer_name}</p>
                        </div>
                        <div>
                            <p className="text-[10px] text-slate-500 font-bold uppercase">Surat Jalan Dibawa</p>
                            <p className={`font-semibold ${selectedCount === 0 ? 'text-amber-600' : 'text-blue-600'}`}>
                                {selectedCount === 0 ? 'SJ ditinggal/tunda' : `${selectedCount} Dokumen`}
                            </p>
                        </div>
                    </div>

                    {selectedCount === 0 && (
                        <Alert className="mb-6 bg-amber-50 border-amber-100">
                            <AlertCircle className="h-4 w-4 text-amber-600" />
                            <AlertDescription className="text-[11px] text-amber-800">
                                Perhatian: Anda keluar tanpa memilih Surat Jalan. SJ akan tetap berada di status "On Customer" untuk driver lain.
                            </AlertDescription>
                        </Alert>
                    )}

                    <Button
                        className="w-full bg-green-600 hover:bg-green-700 py-6 font-bold h-auto"
                        onClick={handleSubmitCheckOut}
                        disabled={submitting}
                    >
                        {submitting ? <Loader2 className="animate-spin mr-2" /> : 'Konfirmasi Check-Out'}
                    </Button>
                    <Button variant="ghost" className="w-full mt-2" onClick={() => setStep('select-shipments')}>
                        Batal
                    </Button>
                </Card>
            )}

            {step === 'success' && (
                <Card className="p-8 text-center">
                    <CheckCircle2 className="w-16 h-16 text-green-500 mx-auto mb-4" />
                    <h2 className="text-2xl font-bold">Berhasil!</h2>
                    <p className="text-slate-500 mt-2 mb-6 text-sm">
                        {selectedCount === 0
                            ? "Check-out berhasil. Surat jalan telah ditunda di lokasi customer."
                            : "Check-out selesai. Pastikan dokumen surat jalan sudah terbawa."}
                    </p>
                    <Button className="w-full py-6 h-auto font-bold" onClick={() => {
                        onSuccess();
                        onBack();
                    }}>
                        Kembali ke Menu
                    </Button>
                </Card>
            )}
        </div>
    );
}