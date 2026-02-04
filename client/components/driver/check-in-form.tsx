'use client';

import { useState, useEffect, useMemo } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Spinner } from '@/components/ui/spinner';
import { apiService, SuratJalanItem } from '@/lib/api-service';
import { AlertCircle, CheckCircle2, ArrowLeft, Search, X } from 'lucide-react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { useAuth } from '@/hooks/useAuth';

interface GroupedCustomer {
    customer_id: number;
    customer_name: string;
    tnkb_no: string;
    tnkb_id: number;
    m_inout_ids: number[];
    document_nos: string[];
}

interface CheckedInCustomer {
    customer_id: number;
    customer_name: string;
    tnkb_no: string;
    tnkb_id: number;
}

interface CheckInFormProps {
    onBack: () => void;
    onError: (message: string) => void;
    onSuccess: (customer: CheckedInCustomer) => void;
}

type FormStep = 'select-customer' | 'confirm' | 'success';

export function CheckInForm({ onBack, onError, onSuccess }: CheckInFormProps) {
    const { user } = useAuth();
    const [step, setStep] = useState<FormStep>('select-customer');
    const [customers, setCustomers] = useState<SuratJalanItem[]>([]);
    const [searchQuery, setSearchQuery] = useState('');

    // PERBAIKAN 1: Gunakan GroupedCustomer untuk state selected
    const [selectedCustomer, setSelectedCustomer] = useState<GroupedCustomer | null>(null);

    const [loading, setLoading] = useState(false);
    const [submitting, setSubmitting] = useState(false);

    useEffect(() => {
        const loadCustomers = async () => {
            if (!user?.user_id) return;
            setLoading(true);
            try {
                const driverIdNum = parseInt(user.user_id);
                const data = await apiService.getActiveShipments(driverIdNum);
                setCustomers(data);
            } catch (error) {
                onError('Failed to load customers');
                console.error(error);
            } finally {
                setLoading(false);
            }
        };
        loadCustomers();
    }, [user, onError]);

    const groupedCustomers = useMemo(() => {
        const groups: { [key: string]: GroupedCustomer } = {};

        customers.forEach((item) => {
            if (!groups[item.customer_name]) {
                groups[item.customer_name] = {
                    customer_id: parseInt(item.customer_id),
                    customer_name: item.customer_name,
                    tnkb_no: item.tnkb_no || '',
                    tnkb_id: parseInt(item.tnkb_id) || 0,
                    m_inout_ids: [],
                    document_nos: [],
                };
            }
            groups[item.customer_name].m_inout_ids.push(item.m_inout_id);
            groups[item.customer_name].document_nos.push(item.document_no);
        });

        const result = Object.values(groups);

        if (!searchQuery.trim()) return result;
        const query = searchQuery.toLowerCase();
        return result.filter(c =>
            c.customer_name.toLowerCase().includes(query) ||
            c.document_nos.some(doc => doc.toLowerCase().includes(query))
        );
    }, [customers, searchQuery]);

    // PERBAIKAN 2: Update parameter type
    const handleSelectCustomer = (customer: GroupedCustomer) => {
        setSelectedCustomer(customer);
        setStep('confirm');
    };

    const handleSubmitCheckIn = async () => {
        if (!selectedCustomer || !user?.user_id) {
            onError('Missing customer or driver information');
            return;
        }

        setSubmitting(true);
        try {
            const driverIdNum = parseInt(user.user_id);
            const response = await apiService.processHandover({
                // PERBAIKAN 3: Kirim m_inout_ids (array) bukan ID tunggal
                m_inout_ids: selectedCustomer.m_inout_ids,
                status: 'HO: DRIVER_CHECKIN',
                user_id: driverIdNum,
                driver_by: driverIdNum,
                // notes: `Driver Check-In ${selectedCustomer.m_inout_ids.length} SJ via App`,
                notes: `Driver Check-In`,
                customer_id: selectedCustomer.customer_id,
                tnkb_id: selectedCustomer.tnkb_id,
            });

            if (response.success) {
                onSuccess({
                    // Menggunakan ID customer, bukan ID transaksi
                    customer_id: selectedCustomer.customer_id,
                    customer_name: selectedCustomer.customer_name,
                    tnkb_no: selectedCustomer.tnkb_no || '',
                    tnkb_id: selectedCustomer.tnkb_id,
                });
                setStep('success');
            } else {
                onError(response.message || 'Check-in failed');
            }
        } catch (error) {
            onError(error instanceof Error ? error.message : 'Check-in failed');
            console.error(error);
        } finally {
            setSubmitting(false);
        }
    };

    if (!user) {
        return (
            <div className="flex flex-col items-center justify-center py-20 space-y-4">
                <Spinner className="w-8 h-8" />
                <p className="text-slate-500 text-sm">Authenticating...</p>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            {step !== 'success' && (
                <Button
                    variant="ghost"
                    onClick={onBack}
                    className="w-full justify-start text-slate-600 hover:text-slate-900 hover:bg-slate-100"
                >
                    <ArrowLeft className="w-4 h-4 mr-2" />
                    Kembali
                </Button>
            )}

            {step === 'select-customer' && (
                <Card className="p-6 bg-white border-slate-200 shadow-sm">
                    <h2 className="text-lg font-bold text-slate-900 mb-4">Pilih Customer</h2>

                    {loading ? (
                        <div className="flex flex-col items-center justify-center py-12 space-y-3">
                            <Spinner />
                            <p className="text-sm text-slate-500">Fetching shipments...</p>
                        </div>
                    ) : customers.length === 0 ? (
                        <Alert className="bg-amber-50 border-amber-200">
                            <AlertCircle className="h-4 w-4 text-amber-600" />
                            <AlertDescription className="text-amber-800">
                                Data pengiriman kosong. Belum ada Surat Jalan yang ditugaskan ke ID Anda hari ini.
                            </AlertDescription>
                        </Alert>
                    ) : (
                        <>
                            <div className="mb-4 relative">
                                <Search className="absolute left-3 top-3 w-4 h-4 text-slate-400" />
                                <Input
                                    placeholder="Search name or document..."
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    className="pl-10 pr-10"
                                />
                                {searchQuery && (
                                    <button
                                        onClick={() => setSearchQuery('')}
                                        className="absolute right-3 top-3 text-slate-400 hover:text-slate-600"
                                    >
                                        <X className="w-4 h-4" />
                                    </button>
                                )}
                            </div>

                            <div className="max-h-[400px] overflow-y-auto space-y-2 pr-1">
                                {groupedCustomers.map((customer) => (
                                    <button
                                        key={customer.customer_name}
                                        onClick={() => handleSelectCustomer(customer)}
                                        className="w-full text-left p-4 bg-slate-50 rounded-lg border border-slate-200 hover:border-blue-400 hover:bg-blue-50 transition-all active:scale-[0.98]"
                                    >
                                        <div className="flex justify-between items-start">
                                            <div>
                                                <p className="font-bold text-slate-900">{customer.customer_name}</p>
                                                <p className="text-xs text-slate-500 line-clamp-1">
                                                    {customer.document_nos.join(', ')}
                                                </p>
                                            </div>
                                            {customer.m_inout_ids.length > 1 && (
                                                <span className="bg-blue-100 text-blue-700 text-[10px] px-2 py-1 rounded-full font-bold">
                                                    {customer.m_inout_ids.length} SJ
                                                </span>
                                            )}
                                        </div>
                                        {customer.tnkb_no && (
                                            <div className="mt-2 inline-block px-2 py-0.5 bg-slate-200 text-slate-700 text-[10px] font-bold rounded">
                                                {customer.tnkb_no}
                                            </div>
                                        )}
                                    </button>
                                ))}
                            </div>
                        </>
                    )}
                </Card>
            )}

            {step === 'confirm' && selectedCustomer && (
                <Card className="p-6 bg-white border-slate-200">
                    <h2 className="text-lg font-bold text-slate-900 mb-4">Confirm Check In</h2>
                    <div className="space-y-4">
                        <div className="bg-blue-50 border border-blue-100 rounded-xl p-5">
                            <div className="mb-4">
                                <label className="text-[10px] font-bold text-blue-500 uppercase tracking-wider">Customer</label>
                                <p className="text-xl font-extrabold text-blue-900 leading-tight">
                                    {selectedCustomer.customer_name}
                                </p>
                            </div>
                            <div className="grid grid-cols-1 gap-4 pt-4 border-t border-blue-100">
                                <div>
                                    <label className="text-[10px] font-bold text-blue-500 uppercase tracking-wider">
                                        Surat Jalan ({selectedCustomer.m_inout_ids.length})
                                    </label>
                                    <p className="text-sm font-semibold text-blue-800 break-words">
                                        {selectedCustomer.document_nos.join(', ')}
                                    </p>
                                </div>
                                {selectedCustomer.tnkb_no && (
                                    <div>
                                        <label className="text-[10px] font-bold text-blue-500 uppercase tracking-wider">Vehicle</label>
                                        <p className="text-sm font-semibold text-blue-800">{selectedCustomer.tnkb_no}</p>
                                    </div>
                                )}
                            </div>
                        </div>

                        <div className="flex flex-col gap-2">
                            <Button
                                onClick={handleSubmitCheckIn}
                                disabled={submitting}
                                className="w-full bg-green-600 hover:bg-green-700 text-white font-bold h-14"
                            >
                                {submitting ? <Spinner className="mr-2" /> : 'Confirm & Check In'}
                            </Button>
                            <Button
                                onClick={() => setStep('select-customer')}
                                variant="outline"
                                className="w-full h-12"
                            >
                                Change Customer
                            </Button>
                        </div>
                    </div>
                </Card>
            )}

            {step === 'success' && selectedCustomer && (
                <Card className="p-8 bg-white border-green-100 text-center">
                    <div className="flex justify-center mb-6">
                        <div className="bg-green-100 p-4 rounded-full">
                            <CheckCircle2 className="w-12 h-12 text-green-600" />
                        </div>
                    </div>
                    <h3 className="text-2xl font-black text-slate-900 mb-2">Success!</h3>
                    <p className="text-slate-600 mb-8">
                        Checked in at <span className="font-bold">{selectedCustomer.customer_name}</span>.
                        Safe work!
                    </p>
                    <Button onClick={onBack} className="w-full bg-slate-900 text-white h-12">
                       Kembali ke Menu
                    </Button>
                </Card>
            )}
        </div>
    );
}