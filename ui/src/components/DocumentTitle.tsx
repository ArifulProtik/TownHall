import { useDocumentTitle } from '@/lib/documentTitle';

interface DocumentTitleProps {
  title?: string;
  fallback?: string;
}

/**
 * Declarative page title. Render it anywhere in a page:
 *
 *   <DocumentTitle title={profile ? `${profile.name} | TownHall` : `@${handle} | TownHall`} />
 *
 * Pass a value that is available on first render (e.g. derived from route
 * params) and upgrade it when data loads — that avoids flashing the fallback.
 */
export function DocumentTitle({ title, fallback }: DocumentTitleProps) {
  useDocumentTitle(title, fallback);
  return null;
}
