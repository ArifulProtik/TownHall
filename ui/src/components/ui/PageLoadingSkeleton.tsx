export function PageLoadingSkeleton() {
  return (
    <div className="flex h-full w-full min-h-[50vh] items-center justify-center p-6">
      <div className="flex flex-col items-center gap-3">
        <div className="size-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        <span className="text-xs font-medium text-muted-foreground animate-pulse">Loading...</span>
      </div>
    </div>
  );
}

export default PageLoadingSkeleton;
