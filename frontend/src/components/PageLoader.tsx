export function PageLoader({ label = "Loading" }: { label?: string }) {
  return (
    <div
      className="pointer-events-none fixed inset-0 z-10 flex items-center justify-center"
      role="status"
      aria-label={label}
    >
      <span className="h-5 w-5 animate-spin rounded-full border border-white/10 border-t-white/45 motion-reduce:animate-none" />
      <span className="sr-only">{label}</span>
    </div>
  );
}
