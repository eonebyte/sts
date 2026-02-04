'use client';

import { useState, useRef, useEffect } from 'react';
import Image from 'next/image';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { CheckInForm } from './check-in-form';
import { CheckOutForm } from './check-out-form';
import { ShipmentList } from './shipment-list';
import { AlertCircle, LogOut, User, ChevronDown, Loader2, ArrowBigRight } from 'lucide-react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { logout, useAuth } from '@/hooks/useAuth'; // Pastikan path ini benar
import { apiService } from '@/lib/api-service';

type ViewType = 'menu' | 'checkin' | 'checkout' | 'shipments';

interface CheckedInCustomer {
    customer_id: string;
    customer_name: string;
    tnkb: string;
}

export function DriverDashboard() {
    const { user } = useAuth();
    const [currentView, setCurrentView] = useState<ViewType>('menu');
    const [errorMessage, setErrorMessage] = useState<string | null>(null);
    const [checkedInCustomer, setCheckedInCustomer] = useState<CheckedInCustomer | null>(null);
    const [showProfileMenu, setShowProfileMenu] = useState(false);
    const [isValidating, setIsValidating] = useState(true);
    const profileMenuRef = useRef<HTMLDivElement>(null);

    // 1. Efek untuk Validasi Status dari DB saat Refresh
    useEffect(() => {
        const fetchActiveStatus = async () => {
            if (!user?.user_id) {
                setIsValidating(false);
                return;
            }

            try {
                const driverIdNum = parseInt(user.user_id);
                const shipments = await apiService.getOnCustomerShipments(driverIdNum);

                // Cari apakah ada shipment yang statusnya sedang Check-In
                const activeJob = shipments.find(
                    (s) => s.status === 'HO: DRIVER_CHECKIN'
                );

                if (activeJob) {
                    setCheckedInCustomer({
                        customer_id: activeJob.customer_id,
                        customer_name: activeJob.customer_name,
                        tnkb: activeJob.tnkb_no || '',
                    });
                } else {
                    setCheckedInCustomer(null);
                }
            } catch (error) {
                console.error('Error validating status:', error);
            } finally {
                setIsValidating(false);
            }
        };

        fetchActiveStatus();
    }, [user]);

    // Close profile menu when clicking outside
    useEffect(() => {
        function handleClickOutside(event: MouseEvent) {
            if (profileMenuRef.current && !profileMenuRef.current.contains(event.target as Node)) {
                setShowProfileMenu(false);
            }
        }
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const clearError = () => setErrorMessage(null);
    const handleError = (message: string) => setErrorMessage(message);

    const handleCheckInSuccess = (customer: any) => {
        setCheckedInCustomer({
            customer_id: customer.customer_id,
            customer_name: customer.customer_name,
            tnkb: customer.tnkb_no || customer.tnkb || '',
        });
        setCurrentView('menu');
    };

    const handleCheckOutComplete = () => {
        setCheckedInCustomer(null);
        setCurrentView('menu');
    };

    if (isValidating) {
        return (
            <div className="min-h-screen flex flex-col items-center justify-center bg-slate-50">
                <Loader2 className="w-8 h-8 animate-spin text-blue-600 mb-2" />
                <p className="text-slate-500 text-sm">Memvalidasi status...</p>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gradient-to-b from-slate-50 to-slate-100 pb-8">
            {/* Header */}
            <div className="sticky top-0 z-20 bg-white border-b border-slate-200 shadow-sm">
                <div className="max-w-md mx-auto px-4 py-4">
                    <div className="flex items-center gap-3">
                        <div className="flex-shrink-0">
                            <Image src="/sts.png" alt="Logo" width={48} height={48} className="rounded-lg" priority />
                        </div>
                        <div className="flex-1">
                            <h1 className="text-2xl font-bold text-slate-900">Driver Portal</h1>
                            <p className="text-xs text-slate-600 mt-0.5">Shipment Tracking System</p>
                        </div>

                        <div className="relative" ref={profileMenuRef}>
                            <button onClick={() => setShowProfileMenu(!showProfileMenu)} className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-slate-100 transition text-slate-700">
                                <User className="w-5 h-5" />
                                <ChevronDown className="w-4 h-4" />
                            </button>

                            {showProfileMenu && (
                                <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg border border-slate-200 shadow-lg">
                                    <div className="px-4 py-3 border-b border-slate-200">
                                        <p className="text-sm font-semibold text-slate-900">{user?.username || 'Driver'}</p>
                                        {/* <p className="text-xs text-slate-600">ID: {user?.user_id}</p> */}
                                    </div>
                                    <button onClick={logout} className="w-full flex items-center gap-2 px-4 py-2 text-sm text-red-600 hover:bg-red-50 transition">
                                        <LogOut className="w-4 h-4" /> Logout
                                    </button>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>

            <div className="max-w-md mx-auto px-4 py-6 space-y-4">
                {errorMessage && (
                    <Alert variant="destructive">
                        <AlertCircle className="h-4 w-4" />
                        <AlertDescription className="flex justify-between w-full items-center">
                            {errorMessage}
                            <button onClick={clearError} className="text-xs font-bold underline">Tutup</button>
                        </AlertDescription>
                    </Alert>
                )}

                {currentView === 'menu' && (
                    <div className="space-y-3">
                        <Card className="p-5 bg-white border-none shadow-md overflow-hidden relative">
                            <div className="absolute top-0 left-0 w-1 h-full bg-emerald-500" />
                            <div className="flex items-center gap-4">
                                <div className="flex-shrink-0 bg-emerald-50 p-3 rounded-full">
                                    <span className="text-2xl font-bold">👋</span>
                                </div>
                                <div>
                                    <h2 className="text-sm font-black text-emerald-700 uppercase tracking-wide flex items-center gap-2">
                                        Halo, Pak {user?.username || 'Driver'}!
                                    </h2>
                                    {/* <p className="text-slate-700 font-medium leading-tight mt-1">
                                        Bismillah, utamakan keselamatan, moga urusan lancar, dan jadi rezeki berkah buat keluarga. 🙏 <span className="font-bold">Amin.</span>
                                    </p> */}
                                </div>
                            </div>
                        </Card>
                        <Card className="p-6 bg-white border-none shadow-md">
                            {/* <h2 className="text-2xl font-black text-slate-900 text-center">
                                Welcome, Pak {user?.username || 'Pak Driver'} 👋
                            </h2> */}
                            {/* <h2 className="text-lg font-semibold text-slate-900 mb-6 text-center">
                                Siap Gaspol hari ini ?
                            </h2> */}
                            <div className="space-y-4">
                                <Button
                                    onClick={() => setCurrentView('checkin')}
                                    disabled={!!checkedInCustomer}
                                    className={`w-full py-8 text-lg font-bold transition-all flex items-center justify-center gap-3 ${checkedInCustomer
                                        ? 'bg-slate-100 text-slate-400'
                                        : 'bg-blue-600 hover:bg-blue-700 shadow-lg shadow-blue-100 text-white'
                                        }`}
                                >
                                    {checkedInCustomer ? (
                                        '✓ Sudah Check-In'
                                    ) : (
                                        <>
                                            {/* !w-8 dan !h-8 setara dengan 32px */}
                                            <ArrowBigRight className="!w-6 !h-6 animate-horizontal" />
                                            <span>Check-In</span>
                                        </>
                                    )}
                                </Button>

                                <Button
                                    onClick={() => setCurrentView('checkout')}
                                    disabled={!checkedInCustomer}
                                    className={`w-full py-8 text-lg font-bold transition-all flex items-center justify-center gap-3 ${checkedInCustomer
                                        ? 'bg-green-600 hover:bg-green-700 shadow-lg shadow-green-100 text-white'
                                        : 'bg-slate-100 text-slate-400'
                                        }`}
                                >
                                    {checkedInCustomer && <ArrowBigRight className="!w-6 !h-6 animate-horizontal" />}
                                    <span>
                                        Check-Out {checkedInCustomer ? `(${checkedInCustomer.customer_name})` : ''}
                                    </span>
                                </Button>

                                <div className="pt-4">
                                    <Button
                                        onClick={() => setCurrentView('shipments')}
                                        variant="outline"
                                        className="w-full py-6 text-slate-600 border-slate-200"
                                    >
                                        📋 Daftar Pengiriman Saya
                                    </Button>
                                </div>
                            </div>
                        </Card>

                        {checkedInCustomer && (
                            <Card className="p-4 bg-amber-50 border-amber-200 border animate-pulse">
                                <p className="text-sm text-amber-800 flex items-center gap-2">
                                    <AlertCircle className="w-4 h-4" />
                                    <span>Kamu sedang aktif di <b>{checkedInCustomer.customer_name}</b>. Segera Check-Out jika tugas selesai.</span>
                                </p>
                            </Card>
                        )}
                    </div>
                )}

                {currentView === 'checkin' && (
                    <CheckInForm
                        onBack={() => setCurrentView('menu')}
                        onError={handleError}
                        onSuccess={handleCheckInSuccess}
                    />
                )}

                {currentView === 'checkout' && checkedInCustomer && (
                    <CheckOutForm
                        onBack={() => setCurrentView('menu')}
                        onError={handleError}
                        onSuccess={handleCheckOutComplete}
                        checkedInCustomer={checkedInCustomer}
                    />
                )}

                {currentView === 'shipments' && (
                    <ShipmentList onBack={() => setCurrentView('menu')} />
                )}
            </div>
        </div>
    );
}