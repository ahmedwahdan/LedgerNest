export default function ExpensesLoading() {
  return (
    <div className="shell-grid flex flex-1 flex-col overflow-auto">
      <div className="mx-auto w-full max-w-4xl space-y-6 px-5 py-8 sm:px-8">
        <div className="flex items-center justify-between">
          <div className="skeleton h-10 w-32" />
          <div className="skeleton h-9 w-28 rounded-full" />
        </div>
        <div className="flex gap-3">
          {[1, 2, 3].map((i) => <div key={i} className="skeleton h-9 w-28 rounded-xl" />)}
        </div>
        <div className="rounded-[2rem] overflow-hidden space-y-px">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="skeleton h-16 rounded-none first:rounded-t-[2rem] last:rounded-b-[2rem]" />
          ))}
        </div>
      </div>
    </div>
  )
}
