import * as React from 'react';
import { Camera, Spinner } from '@phosphor-icons/react';
import { useUploadFileMutation } from '@/features/profile/profileApi';
import { cn } from 'cn';

interface ImageUploadButtonProps {
  onUploaded: (url: string) => void;
  className?: string;
  label?: string;
  iconOnly?: boolean;
}

export function ImageUploadButton({
  onUploaded,
  className,
  label = 'Change Photo',
  iconOnly = false,
}: ImageUploadButtonProps) {
  const [uploadFile, { isLoading }] = useUploadFileMutation();
  const fileInputRef = React.useRef<HTMLInputElement>(null);
  const [error, setError] = React.useState<string | null>(null);

  async function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;

    // Reset input so re-selecting same file triggers change
    e.target.value = '';

    if (!file.type.startsWith('image/')) {
      setError('Please select an image file (PNG, JPG, WebP, GIF)');
      return;
    }

    if (file.size > 8 * 1024 * 1024) {
      setError('Image must be less than 8MB');
      return;
    }

    setError(null);
    const formData = new FormData();
    formData.append('file', file);

    try {
      const res = await uploadFile(formData).unwrap();
      onUploaded(res.url);
    } catch (err: unknown) {
      const apiErr = err as { data?: { error?: string } };
      setError(apiErr?.data?.error || 'Failed to upload image. Please try again.');
    }
  }

  return (
    <div className="relative inline-flex flex-col items-center">
      <input
        ref={fileInputRef}
        type="file"
        accept="image/jpeg,image/png,image/webp,image/gif"
        className="hidden"
        onChange={handleFileChange}
        disabled={isLoading}
      />
      <button
        type="button"
        disabled={isLoading}
        onClick={() => fileInputRef.current?.click()}
        aria-label={label}
        className={cn(
          'flex items-center justify-center gap-1.5 rounded-full bg-black/60 text-white backdrop-blur-md transition-all hover:bg-black/80 hover:scale-105 active:scale-95 disabled:pointer-events-none disabled:opacity-50 cursor-pointer shadow-md',
          iconOnly ? 'size-9' : 'px-3 py-1.5 text-xs font-medium',
          className,
        )}
      >
        {isLoading ? (
          <Spinner className="size-4 animate-spin text-white" />
        ) : (
          <Camera className="size-4" weight="bold" />
        )}
        {!iconOnly && <span>{isLoading ? 'Uploading...' : label}</span>}
      </button>
      {error && (
        <span className="absolute -bottom-6 whitespace-nowrap text-xs font-medium text-destructive">
          {error}
        </span>
      )}
    </div>
  );
}
