import { Link } from "react-router-dom";

export function MusicTile({
  to,
  imageUrl,
  title,
  subtitle,
  round,
}: {
  to: string;
  imageUrl?: string;
  title: string;
  subtitle?: string;
  round?: boolean;
}) {
  return (
    <Link
      to={to}
      className="group w-36 shrink-0 transition-transform duration-200 ease-out hover:scale-105 focus:scale-105 focus:outline-none sm:w-44"
    >
      <div
        className={`aspect-square w-full overflow-hidden bg-neutral-800 shadow-lg ${round ? "rounded-full" : "rounded-md"}`}
      >
        {imageUrl ? (
          <img src={imageUrl} alt={title} loading="lazy" className="h-full w-full object-cover" />
        ) : (
          <div className="flex h-full w-full items-center justify-center px-2 text-center text-sm text-neutral-500">
            {title}
          </div>
        )}
      </div>
      <p className="mt-2 truncate text-sm font-medium text-neutral-100">{title}</p>
      {subtitle && <p className="truncate text-xs text-neutral-400">{subtitle}</p>}
    </Link>
  );
}
