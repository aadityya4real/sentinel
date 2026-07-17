import { lazy, Suspense } from 'react';
import { createBrowserRouter, RouterProvider, Navigate } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { queryClient } from '@/lib/queryClient';
import { AppLayout } from '@/layouts/AppLayout';
import { Spinner } from '@/components/ui/Spinner';

const DashboardPage = lazy(() => import('@/pages/DashboardPage'));
const HostsPage = lazy(() => import('@/pages/HostsPage'));
const HostDetailPage = lazy(() => import('@/pages/HostDetailPage'));
const EventsPage = lazy(() => import('@/pages/EventsPage'));
const ReplayPage = lazy(() => import('@/pages/ReplayPage'));
const TimeMachinePage = lazy(() => import('@/pages/TimeMachinePage'));
const AIPage = lazy(() => import('@/pages/AIPage'));
const SettingsPage = lazy(() => import('@/pages/SettingsPage'));

function PageFallback() {
  return (
    <div className="flex h-full items-center justify-center">
      <Spinner />
    </div>
  );
}

const router = createBrowserRouter([
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: 'dashboard', element: <Suspense fallback={<PageFallback />}><DashboardPage /></Suspense> },
      { path: 'hosts', element: <Suspense fallback={<PageFallback />}><HostsPage /></Suspense> },
      { path: 'hosts/:hostname', element: <Suspense fallback={<PageFallback />}><HostDetailPage /></Suspense> },
      { path: 'events', element: <Suspense fallback={<PageFallback />}><EventsPage /></Suspense> },
      { path: 'replay', element: <Suspense fallback={<PageFallback />}><ReplayPage /></Suspense> },
      { path: 'replay/:hostname', element: <Suspense fallback={<PageFallback />}><ReplayPage /></Suspense> },
      { path: 'time-machine', element: <Suspense fallback={<PageFallback />}><TimeMachinePage /></Suspense> },
      { path: 'ai', element: <Suspense fallback={<PageFallback />}><AIPage /></Suspense> },
      { path: 'settings', element: <Suspense fallback={<PageFallback />}><SettingsPage /></Suspense> },
    ],
  },
]);

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  );
}
