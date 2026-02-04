'use client';

import { useState, useEffect, useMemo } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Spinner } from '@/components/ui/spinner';
import { apiService, SuratJalanItem } from '@/lib/api-service';
import { ArrowLeft, Package, Search, X } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';

interface ShipmentListProps {
    onBack: () => void;
}
export function ShipmentList({ onBack }: ShipmentListProps) {
    const { user } = useAuth();
    const [shipments, setShipments] = useState<SuratJalanItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [filterDate, setFilterDate] = useState('');
    const [driverId] = useState('1'); // Mock driver ID

    useEffect(() => {
        const loadShipments = async () => {
            if (!user?.user_id) return;
            setLoading(true);
            try {
                const driverIdNum = parseInt(user.user_id);
                const data = await apiService.getActiveShipments(driverIdNum);
                setShipments(data);
            } catch (error) {
                console.error('Failed to load shipments:', error);
            } finally {
                setLoading(false);
            }
        };

        loadShipments();
    }, [user]);

    const getStatusColor = (status?: string) => {
        switch (status) {
            case 'Pending':
                return 'bg-yellow-100 text-yellow-800';
            case 'In Transit':
                return 'bg-blue-100 text-blue-800';
            case 'Completed':
                return 'bg-green-100 text-green-800';
            default:
                return 'bg-slate-100 text-slate-800';
        }
    };

    // Filter shipments based on search query and date
    const filteredShipments = useMemo(() => {
        let result = shipments;

        // Filter by search query
        if (searchQuery.trim()) {
            const query = searchQuery.toLowerCase();
            result = result.filter(
                (shipment) =>
                    shipment.customer_name.toLowerCase().includes(query) ||
                    shipment.document_no.toLowerCase().includes(query) ||
                    shipment.tnkb?.toLowerCase().includes(query)
            );
        }

        // Filter by date (if provided, match shipments from that date)
        if (filterDate) {
            result = result.filter((shipment) => {
                // Mock: assuming shipments have a date property or we can use current date
                // In real scenario, shipment would have a created_at or similar date field
                const shipmentDate = new Date().toISOString().split('T')[0]; // Mock current date
                return shipmentDate === filterDate;
            });
        }

        return result;
    }, [shipments, searchQuery, filterDate]);

    return (
        <div className="space-y-4">
            {/* Back Button */}
            <Button
                variant="ghost"
                onClick={onBack}
                className="w-full justify-start text-slate-600 hover:text-slate-900 hover:bg-slate-100"
            >
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back to Menu
            </Button>

            {/* Header */}
            <Card className="p-4 bg-white border border-slate-200">
                <h2 className="text-lg font-semibold text-slate-900">Pengiriman sedang aktif</h2>
                <p className="text-sm text-slate-600 mt-1">
                    {loading ? 'Loading...' : `${filteredShipments.length} of ${shipments.length} shipment(s)`}
                </p>
            </Card>

            {/* Filters */}
            {!loading && shipments.length > 0 && (
                <Card className="p-4 bg-white border border-slate-200 space-y-3">
                    <div>
                        <label className="text-xs font-semibold text-slate-700 block mb-2">
                            Search
                        </label>
                        <div className="relative">
                            <Search className="absolute left-3 top-3 w-4 h-4 text-slate-400" />
                            <Input
                                placeholder="Search by name, shipment, or vehicle..."
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
                    </div>

                    <div>
                        <label className="text-xs font-semibold text-slate-700 block mb-2">
                            Filter by Date
                        </label>
                        <div className="flex gap-2">
                            <Input
                                type="date"
                                value={filterDate}
                                onChange={(e) => setFilterDate(e.target.value)}
                            />
                            {filterDate && (
                                <Button
                                    onClick={() => setFilterDate('')}
                                    variant="outline"
                                    className="px-3"
                                >
                                    <X className="w-4 h-4" />
                                </Button>
                            )}
                        </div>
                    </div>
                </Card>
            )}

            {/* Loading State */}
            {loading && (
                <div className="flex items-center justify-center py-12">
                    <div className="text-center">
                        <Spinner className="mx-auto mb-4" />
                        <p className="text-slate-600 text-sm">Loading shipments...</p>
                    </div>
                </div>
            )}

            {/* Empty State */}
            {!loading && shipments.length === 0 && (
                <Card className="p-8 bg-slate-50 text-center border border-slate-200">
                    <Package className="w-12 h-12 mx-auto text-slate-400 mb-3" />
                    <p className="text-slate-600 font-medium">No active shipments</p>
                    <p className="text-sm text-slate-500 mt-1">
                        All shipments have been completed or there are no assignments yet.
                    </p>
                </Card>
            )}

            {/* Shipment List */}
            {!loading && shipments.length > 0 && filteredShipments.length > 0 && (
                <div className="space-y-3">
                    {filteredShipments.map((shipment, index) => (
                        <Card key={shipment.m_inout_id} className="p-4 bg-white hover:shadow-md transition">
                            <div className="flex items-start justify-between gap-3">
                                <div className="flex-1">
                                    <div className="flex items-center gap-2">
                                        <span className="inline-flex items-center justify-center w-6 h-6 bg-blue-100 text-blue-700 rounded-full text-xs font-semibold">
                                            {index + 1}
                                        </span>
                                        <h3 className="font-semibold text-slate-900">{shipment.document_no}</h3>
                                    </div>

                                    <p className="text-sm text-slate-600 mt-2">{shipment.customer_name}</p>

                                    <div className="mt-3 space-y-1">
                                        {shipment.tnkb && (
                                            <div className="text-xs text-slate-600">
                                                <span className="font-semibold">TNKB:</span> {shipment.tnkb}
                                            </div>
                                        )}
                                        {shipment.status && (
                                            <div>
                                                <Badge className={`text-xs ${getStatusColor(shipment.status)}`}>
                                                    {shipment.status}
                                                </Badge>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </Card>
                    ))}

                    {/* Summary */}
                    <Card className="p-4 bg-blue-50 border border-blue-200 mt-4">
                        <p className="text-sm text-blue-900">
                            <span className="font-semibold">📦 Total:</span> {shipments.length} shipment(s)
                            assigned to you
                        </p>
                        <p className="text-xs text-blue-700 mt-2">
                            Complete your check-in to start delivery
                        </p>
                    </Card>
                </div>
            )}

            {/* No results state */}
            {!loading && shipments.length > 0 && filteredShipments.length === 0 && (
                <Card className="p-8 bg-slate-50 text-center border border-slate-200">
                    <Package className="w-12 h-12 mx-auto text-slate-400 mb-3" />
                    <p className="text-slate-600 font-medium">No shipments match your filters</p>
                    <p className="text-sm text-slate-500 mt-1">
                        Try adjusting your search or date filter
                    </p>
                </Card>
            )}
        </div>
    );
}
