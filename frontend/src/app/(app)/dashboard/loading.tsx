export default function DashboardLoading() {
  return (
    <div className="shell-grid flex flex-1 flex-col overflow-auto">
      <div className="mx-auto w-full max-w-5xl space-y-6 px-5 py-8 sm:px-8">
        <div className="skeleton h-10 w-40" />
        <div className="skeleton h-48 w-full rounded-[2rem]" />
        <div className="skeleton h-64 w-full rounded-[2rem]" />
      </div>
    </div>
  )
}
