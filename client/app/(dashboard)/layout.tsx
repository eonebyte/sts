import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { AppSidebar } from "@/components/AppSidebar";

export default function DashboardLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <SidebarProvider>
            {/* 1. Sidebar Kiri */}
            <AppSidebar />

            {/* 2. Area Konten Utama */}
            <main className="w-full min-h-screen bg-slate-50 flex flex-col">

                {/* Trigger Sidebar (Desktop) */}
                <div className="p-4 pt-1 pb-0">
                    <SidebarTrigger className="text-slate-500 hover:text-slate-900" />
                </div>

                {/* Konten Halaman (Dashboard/Project) */}
                <div className="flex-1 pt-1 pr-4 pb-4 pl-4 md:pt-1 md:pr-5 md:pb-1 md:pl-5 overflow-y-auto">
                    {children}
                </div>
            </main>
        </SidebarProvider>
    );
}