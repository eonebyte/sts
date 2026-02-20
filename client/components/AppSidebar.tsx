'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
    LayoutDashboard,
    LogOut,
    Users,
    ChevronRight,
    Handshake, // Contoh icon lain untuk Handover
    ClipboardCheck,
    User2,
    History,
    BarChart3,
    Clock,
    BookUser
} from "lucide-react";
import {
    Sidebar,
    SidebarContent,
    SidebarHeader,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupContent,
    SidebarGroupLabel,
    SidebarMenu,
    SidebarMenuItem,
    SidebarMenuButton,
    SidebarMenuSub,
    SidebarMenuSubItem,
    SidebarMenuSubButton,
    useSidebar,
} from "@/components/ui/sidebar";
import {
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { logout, useAuth } from "@/hooks/useAuth";
import Image from 'next/image';

// Menu Items Definition
const menuItems = [
    {
        name: 'Dashboard',
        href: '/',
        icon: LayoutDashboard,
        roles: ['Admin', 'driver', 'dpk', 'Delivery']
    },
    {
        name: 'List Driver',
        href: '/shipment/driver',
        icon: BookUser,
        roles: ['Admin', 'dpk']
    },
    {
        name: 'Receipt',
        icon: ClipboardCheck,
        roles: ['Admin', 'Delivery', 'dpk', 'Marketing', 'APIK Staff Accounting'],
        // Menu ini memiliki anak (Submenu)
        children: [
            { name: "from Delivery", href: "/dpk/fromdelivery", roles: ['Admin', 'dpk'] },
            { name: "from Driver", href: "/dpk/fromdriver", roles: ['Admin', 'dpk'] },
            { name: "from DPK", href: "/delivery/fromdpk", roles: ['Admin', 'Delivery'] },
            { name: "from Delivery", href: "/marketing/fromdelivery", roles: ['Admin', 'Marketing'] },
            { name: "from Marketing", href: "/fat/frommarketing", roles: ['Admin', 'APIK Staff Accounting'] },
        ],
    },
    {
        name: 'Handover',
        icon: Handshake,
        roles: ['Admin', 'Delivery', 'dpk', 'Marketing'],
        // Menu ini memiliki anak (Submenu)
        children: [
            { name: "to DPK", href: "/delivery/todpk", roles: ['Admin', 'Delivery'] },
            { name: "to Driver", href: "/dpk/todriver", roles: ['Admin', 'dpk'] },
            { name: "to Delivery", href: "/dpk/todelivery", roles: ['Admin', 'dpk'] },
            { name: "to Marketing", href: "/delivery/tomarketing", roles: ['Admin', 'Delivery'] },
            { name: "to Fat", href: "/marketing/tofat", roles: ['Admin', 'Marketing'] },
        ],
    },
    {
        name: 'Settings',
        href: '/settings',
        icon: Users, // Contoh item tunggal lainnya
        roles: ['Admin']
    },
    {
        name: 'Outstanding',
        href: '/shipment/outstanding/dpk',
        icon: Clock, // Contoh item tunggal lainnya
        roles: ['dpk']
    },
    {
        name: 'Outstanding',
        href: '/shipment/outstanding/delivery',
        icon: Clock, // Contoh item tunggal lainnya
        roles: ['Delivery']
    },
    {
        name: 'History',
        href: '/shipment/history',
        icon: History, // Contoh item tunggal lainnya
        roles: ['Delivery', 'Admin', 'Marketing']
    },
    {
        name: 'Progress',
        href: '/shipment/progress',
        icon: BarChart3, // Contoh item tunggal lainnya
        roles: ['Delivery', 'dpk', 'Admin', 'Marketing']
    },
];


export function AppSidebar() {
    const pathname = usePathname();
    const { user } = useAuth();
    const userRole = user?.title;

    const hasAccess = (itemRoles?: string[]) => {
        if (!itemRoles || itemRoles.length === 0) return true;
        // Gunakan case-insensitive jika perlu (contoh: dpk vs DPK)
        return itemRoles.some(role => role.toLowerCase() === userRole?.toLowerCase());
    };

    return (
        <Sidebar collapsible="icon" className="border-r-slate-800">
            <SidebarHeader className="bg-slate-900 text-white border-b border-slate-800 py-4">
                <SidebarMenu>
                    <SidebarMenuItem>
                        <SidebarMenuButton size="lg" className="hover:bg-slate-800">
                            <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-white p-1">
                                <Image
                                    src="/sts.png"
                                    alt="Logo"
                                    width={32}
                                    height={32}
                                    className="rounded-sm object-contain"
                                />
                            </div>
                            <div className="grid flex-1 text-left text-sm leading-tight">
                                <span className="truncate font-bold text-lg text-white">STS</span>
                                <span className="truncate text-xs text-slate-400">Tracking System</span>
                            </div>
                        </SidebarMenuButton>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarHeader>

            <SidebarContent className="bg-slate-900 text-white">

                <SidebarGroup>
                    <SidebarGroupLabel className="text-slate-500">Main Menu</SidebarGroupLabel>
                    <SidebarGroupContent>
                        <SidebarMenu>
                            {menuItems
                                .filter(item => hasAccess(item.roles)) // Filter Menu Utama
                                .map((item) => {
                                    if (item.children) {
                                        // 1. FILTER ANAKNYA TERLEBIH DAHULU
                                        const allowedChildren = item.children.filter(child => hasAccess(child.roles));

                                        // 2. JIKA TIDAK ADA ANAK YANG BOLEH DIAKSES, JANGAN TAMPILKAN PARENT-NYA
                                        if (allowedChildren.length === 0) return null;

                                        return (
                                            <Collapsible
                                                key={item.name}
                                                asChild
                                                defaultOpen={true}
                                                className="group/collapsible"
                                            >
                                                <SidebarMenuItem>
                                                    <CollapsibleTrigger asChild>
                                                        <SidebarMenuButton
                                                            tooltip={item.name}
                                                            className="text-slate-400 hover:bg-slate-800 hover:text-white"
                                                        >
                                                            <item.icon />
                                                            <span>{item.name}</span>
                                                            <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                                                        </SidebarMenuButton>
                                                    </CollapsibleTrigger>
                                                    <CollapsibleContent>
                                                        <SidebarMenuSub className="border-slate-800 ml-4">
                                                            {/* 3. MAP DARI HASIL FILTER allowedChildren */}
                                                            {allowedChildren.map((subItem) => (
                                                                <SidebarMenuSubItem key={subItem.name}>
                                                                    <SidebarMenuSubButton
                                                                        asChild
                                                                        isActive={pathname === subItem.href}
                                                                        className="text-slate-300 hover:bg-blue-600 hover:text-white data-[active=true]:!bg-blue-600 data-[active=true]:!text-white"
                                                                    >
                                                                        <Link href={subItem.href}>
                                                                            <span>{subItem.name}</span>
                                                                        </Link>
                                                                    </SidebarMenuSubButton>
                                                                </SidebarMenuSubItem>
                                                            ))}
                                                        </SidebarMenuSub>
                                                    </CollapsibleContent>
                                                </SidebarMenuItem>
                                            </Collapsible>
                                        );
                                    }

                                    // Render menu tunggal (sama seperti sebelumnya)
                                    const isActive = pathname === item.href;
                                    return (
                                        <SidebarMenuItem key={item.name}>
                                            <SidebarMenuButton
                                                asChild
                                                isActive={isActive}
                                                tooltip={item.name}
                                                className={`transition-all duration-200 hover:bg-slate-800 hover:text-white data-[active=true]:!bg-blue-600 data-[active=true]:!text-white ${isActive ? '!bg-blue-600 !text-white' : 'text-slate-400'}`}
                                            >
                                                <Link href={item.href || '#'}>
                                                    <item.icon />
                                                    <span>{item.name}</span>
                                                </Link>
                                            </SidebarMenuButton>
                                        </SidebarMenuItem>
                                    );
                                })}
                        </SidebarMenu>
                    </SidebarGroupContent>
                </SidebarGroup>
            </SidebarContent>

            <SidebarFooter className="bg-slate-900 border-t border-slate-800 p-2 space-y-2">
                <SidebarMenu>
                    {/* INFO USER */}
                    <SidebarMenuItem>
                        <SidebarMenuButton size="lg" className="hover:bg-transparent cursor-default">
                            {/* Avatar/Foto Profile */}
                            <div className="flex aspect-square size-8 items-center justify-center rounded-full bg-slate-800 text-slate-400 border border-slate-700 overflow-hidden">
                                {user?.avatar ? (
                                    <Image
                                        src={user.avatar}
                                        alt="User"
                                        width={32}
                                        height={32}
                                        className="object-cover"
                                    />
                                ) : (
                                    <User2 size={18} />
                                )}
                            </div>

                            {/* Nama & Role */}
                            <div className="grid flex-1 text-left text-sm leading-tight">
                                <span className="truncate font-semibold text-slate-200">
                                    {user?.username || "User Name"}
                                </span>
                                <span className="truncate text-xs text-slate-500 capitalize">
                                    {user?.title || "Role"}
                                </span>
                            </div>
                        </SidebarMenuButton>
                    </SidebarMenuItem>

                    {/* TOMBOL LOGOUT */}
                    <SidebarMenuItem>
                        <SidebarMenuButton
                            onClick={logout}
                            className="text-red-400 hover:text-red-300 hover:bg-red-950/30 transition-colors"
                        >
                            <LogOut className="size-4" />
                            <span className="font-medium">Logout</span>
                        </SidebarMenuButton>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarFooter>
        </Sidebar>
    );
}