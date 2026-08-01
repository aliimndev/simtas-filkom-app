export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen">
      <aside className="w-64 border-r bg-background">
        <div className="flex h-14 items-center border-b px-6 font-semibold">
          SIMTAS FILKOM
        </div>
        <nav className="space-y-1 p-4">{/* Nav items will be added here */}</nav>
      </aside>
      <main className="flex-1">
        <header className="flex h-14 items-center border-b px-6">
          {/* Header content */}
        </header>
        <div className="p-6">{children}</div>
      </main>
    </div>
  )
}