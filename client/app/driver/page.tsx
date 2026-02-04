'use client';

import { useEffect, useState } from 'react';
import { DriverDashboard } from '@/components/driver/driver-dashboard';
import { Spinner } from '@/components/ui/spinner';
import { useAuth } from '@/hooks/useAuth';

export default function Home() {
    const isAuthorized = useAuth();
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setIsLoading(false);
    }, []);

    if (!isAuthorized || isLoading) {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <Spinner />
            </div>
        );
    }

    return <DriverDashboard />;
}
