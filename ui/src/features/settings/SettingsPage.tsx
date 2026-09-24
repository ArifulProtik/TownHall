import * as React from 'react';
import {
  CheckCircle,
  WarningCircle,
  Sun,
  Moon,
  ArrowRight,
  SignOut,
  Spinner,
} from '@phosphor-icons/react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Tabs, TabsList, TabsTrigger, TabsContent, TabsIndicator } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { resolveMediaUrl } from '@/lib/media';
import { useTheme } from '@/components/ThemeToggle';
import {
  useGetMeQuery,
  useCheckUsernameQuery,
  useUpdateUsernameMutation,
  useChangePasswordMutation,
  useLogoutAllMutation,
  type UserResponse,
} from '@/features/auth/authApi';
import { useUpdateProfileMutation } from '@/features/profile/profileApi';
import { ImageUploadButton } from '@/features/profile/components/ImageUploadButton';

function ProfileSettingsForm({ user }: { user: UserResponse }) {
  const [updateProfile, { isLoading: isUpdatingProfile }] = useUpdateProfileMutation();
  const [name, setName] = React.useState(user.name || '');
  const [bio, setBio] = React.useState(user.bio || '');
  const [location, setLocation] = React.useState(user.location || '');
  const [website, setWebsite] = React.useState(user.website || '');
  const [avatarUrl, setAvatarUrl] = React.useState(user.avatar_url || '');
  const [bannerUrl, setBannerUrl] = React.useState(user.banner_url || '');
  const [profileSuccess, setProfileSuccess] = React.useState(false);
  const [profileError, setProfileError] = React.useState<string | null>(null);

  async function handleProfileSubmit(e: React.FormEvent) {
    e.preventDefault();
    setProfileError(null);
    setProfileSuccess(false);

    if (name.trim().length < 2) {
      setProfileError('Name must be at least 2 characters');
      return;
    }
    if (bio.length > 280) {
      setProfileError('Bio must not exceed 280 characters');
      return;
    }

    try {
      await updateProfile({
        name: name.trim(),
        bio: bio.trim(),
        location: location.trim(),
        website: website.trim(),
        avatar_url: avatarUrl.trim(),
        banner_url: bannerUrl.trim(),
      }).unwrap();
      setProfileSuccess(true);
      setTimeout(() => setProfileSuccess(false), 3000);
    } catch (err: unknown) {
      const apiErr = err as { data?: { error?: string } };
      setProfileError(apiErr?.data?.error || 'Failed to save profile changes.');
    }
  }

  const initials = user.name ? user.name.slice(0, 2).toUpperCase() : 'TH';

  return (
    <Card>
      <CardHeader>
        <CardTitle>Profile Details</CardTitle>
        <CardDescription>
          This information will be displayed on your public profile.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleProfileSubmit} className="space-y-5">
          {/* Banner and Avatar upload */}
          <div className="space-y-2">
            <Label className="text-xs font-semibold">Cover & Profile Photos</Label>
            <div className="border-border relative h-36 w-full overflow-hidden rounded-xl border bg-gradient-to-br from-[#2f6bff] via-[#6d5cff] to-[#a855f7]">
              {bannerUrl ? (
                <img
                  src={resolveMediaUrl(bannerUrl)}
                  alt="Banner"
                  className="h-full w-full object-cover"
                />
              ) : (
                <>
                  <div className="absolute inset-0 bg-[radial-gradient(420px_circle_at_15%_0%,rgba(255,255,255,0.32),transparent_62%)]" />
                  <div className="absolute inset-0 bg-[radial-gradient(400px_circle_at_88%_108%,rgba(255,186,120,0.38),transparent_65%)]" />
                </>
              )}
              <div className="absolute inset-0 bg-gradient-to-t from-black/50 via-black/5 to-transparent" />
              <div className="absolute top-2.5 right-2.5">
                <ImageUploadButton
                  label="Cover"
                  onUploaded={(url) => setBannerUrl(url)}
                  className="px-2.5 py-1 text-xs"
                />
              </div>
            </div>

            <div className="relative -mt-10 ml-4 flex items-end gap-2.5">
              <div className="relative size-20 shrink-0">
                <div className="border-card bg-muted size-full overflow-hidden rounded-full border-4">
                  <Avatar className="size-full">
                    {avatarUrl && <AvatarImage src={resolveMediaUrl(avatarUrl)} alt={name} />}
                    <AvatarFallback className="text-xl font-bold">{initials}</AvatarFallback>
                  </Avatar>
                </div>
                <div className="absolute -right-1 -bottom-1">
                  <ImageUploadButton
                    iconOnly
                    label="Change Avatar"
                    onUploaded={(url) => setAvatarUrl(url)}
                    className="size-7"
                  />
                </div>
              </div>
              <span className="text-muted-foreground pb-0.5 text-xs">
                Upload with UploadThing cloud storage.
              </span>
            </div>
          </div>

          {profileSuccess && (
            <div className="flex items-center gap-2 rounded-lg bg-emerald-500/10 p-3 text-xs font-medium text-emerald-600 dark:text-emerald-400">
              <CheckCircle className="size-4 shrink-0" weight="fill" />
              <span>Profile successfully updated!</span>
            </div>
          )}

          {profileError && (
            <div className="bg-destructive/10 text-destructive flex items-center gap-2 rounded-lg p-3 text-xs font-medium">
              <WarningCircle className="size-4 shrink-0" weight="fill" />
              <span>{profileError}</span>
            </div>
          )}

          {/* Display Name */}
          <div className="space-y-1.5">
            <Label htmlFor="settings-name" className="text-xs font-semibold">
              Display Name
            </Label>
            <Input
              id="settings-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Your display name"
              required
            />
          </div>

          {/* Bio */}
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <Label htmlFor="settings-bio" className="text-xs font-semibold">
                Bio
              </Label>
              <span
                className={`text-xs ${
                  bio.length > 260
                    ? bio.length > 280
                      ? 'text-destructive font-bold'
                      : 'font-medium text-amber-500'
                    : 'text-muted-foreground'
                }`}
              >
                {bio.length} / 280
              </span>
            </div>
            <Textarea
              id="settings-bio"
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              placeholder="Short bio about yourself"
              rows={3}
              maxLength={280}
            />
          </div>

          {/* Location */}
          <div className="space-y-1.5">
            <Label htmlFor="settings-location" className="text-xs font-semibold">
              Location
            </Label>
            <Input
              id="settings-location"
              value={location}
              onChange={(e) => setLocation(e.target.value)}
              placeholder="City, Country"
            />
          </div>

          {/* Website */}
          <div className="space-y-1.5">
            <Label htmlFor="settings-website" className="text-xs font-semibold">
              Website
            </Label>
            <Input
              id="settings-website"
              type="url"
              value={website}
              onChange={(e) => setWebsite(e.target.value)}
              placeholder="https://yourwebsite.com"
            />
          </div>

          <div className="flex justify-end pt-2">
            <Button type="submit" disabled={isUpdatingProfile}>
              {isUpdatingProfile ? 'Saving...' : 'Save Profile Changes'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function AccountSettingsForm({ user }: { user: UserResponse }) {
  const [usernameInput, setUsernameInput] = React.useState(user.username || '');
  const [debouncedUsername, setDebouncedUsername] = React.useState(user.username || '');
  const [updateUsername, { isLoading: isUpdatingUsername }] = useUpdateUsernameMutation();
  const [usernameSuccess, setUsernameSuccess] = React.useState(false);
  const [usernameError, setUsernameError] = React.useState<string | null>(null);

  const isUsernameChanged =
    usernameInput.trim().toLowerCase() !== (user.username || '').toLowerCase();
  const shouldCheckUsername =
    isUsernameChanged && debouncedUsername.length >= 3 && debouncedUsername.length <= 30;

  const { data: checkData, isFetching: isCheckingUsername } = useCheckUsernameQuery(
    debouncedUsername,
    { skip: !shouldCheckUsername },
  );

  React.useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedUsername(usernameInput.trim().toLowerCase());
    }, 400);
    return () => clearTimeout(handler);
  }, [usernameInput]);

  async function handleUsernameSubmit(e: React.FormEvent) {
    e.preventDefault();
    setUsernameError(null);
    setUsernameSuccess(false);

    const cleanUsername = usernameInput.trim().toLowerCase();
    if (cleanUsername.length < 3 || cleanUsername.length > 30) {
      setUsernameError('Username must be between 3 and 30 characters');
      return;
    }

    try {
      await updateUsername({ username: cleanUsername }).unwrap();
      setUsernameSuccess(true);
      setTimeout(() => setUsernameSuccess(false), 3000);
    } catch (err: unknown) {
      const apiErr = err as { data?: { error?: string } };
      setUsernameError(apiErr?.data?.error || 'Failed to update username.');
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Account Details</CardTitle>
        <CardDescription>
          Manage your login credentials, email address, and unique username.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Email & Provider Display */}
        <div className="space-y-1.5">
          <Label className="text-xs font-semibold">Email Address</Label>
          <div className="flex items-center gap-2">
            <Input
              value={user.email || ''}
              disabled
              className="bg-muted/50 text-muted-foreground cursor-not-allowed"
            />
            {user.email_verified ? (
              <Badge variant="secondary" className="shrink-0 gap-1">
                <CheckCircle className="text-primary size-3.5" weight="fill" />
                <span>Verified</span>
              </Badge>
            ) : (
              <Badge
                variant="outline"
                className="shrink-0 gap-1 border-amber-500/30 text-amber-500"
              >
                <span>Unverified</span>
              </Badge>
            )}
          </div>
        </div>

        <div className="bg-border my-2 h-px w-full" />

        {/* Username Change Form */}
        <form onSubmit={handleUsernameSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <Label htmlFor="settings-username" className="text-xs font-semibold">
                Username / Handle
              </Label>
              {isCheckingUsername && (
                <span className="text-muted-foreground flex items-center gap-1 text-xs">
                  <Spinner className="size-3 animate-spin" />
                  <span>Checking...</span>
                </span>
              )}
              {shouldCheckUsername && !isCheckingUsername && checkData && (
                <span
                  className={`text-xs font-medium ${
                    checkData.available ? 'text-emerald-500' : 'text-destructive'
                  }`}
                >
                  {checkData.available ? '✓ Username available' : '✗ Already taken'}
                </span>
              )}
            </div>
            <div className="relative">
              <span className="text-muted-foreground absolute top-1/2 left-3.5 -translate-y-1/2 text-sm">
                @
              </span>
              <Input
                id="settings-username"
                value={usernameInput}
                onChange={(e) => setUsernameInput(e.target.value)}
                placeholder="handle"
                className="pl-8"
                maxLength={30}
                required
              />
            </div>
            <p className="text-muted-foreground text-xs">
              Must be 3-30 characters using only letters, numbers, and underscores.
            </p>
          </div>

          {usernameSuccess && (
            <div className="flex items-center gap-2 rounded-lg bg-emerald-500/10 p-3 text-xs font-medium text-emerald-600 dark:text-emerald-400">
              <CheckCircle className="size-4 shrink-0" weight="fill" />
              <span>Username updated successfully!</span>
            </div>
          )}

          {usernameError && (
            <div className="bg-destructive/10 text-destructive flex items-center gap-2 rounded-lg p-3 text-xs font-medium">
              <WarningCircle className="size-4 shrink-0" weight="fill" />
              <span>{usernameError}</span>
            </div>
          )}

          <div className="flex justify-end">
            <Button
              type="submit"
              disabled={
                isUpdatingUsername ||
                !isUsernameChanged ||
                (shouldCheckUsername && !checkData?.available)
              }
            >
              {isUpdatingUsername ? 'Updating...' : 'Update Username'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function SecuritySettingsForm() {
  const [changePassword, { isLoading: isChangingPassword }] = useChangePasswordMutation();
  const [currentPassword, setCurrentPassword] = React.useState('');
  const [newPassword, setNewPassword] = React.useState('');
  const [confirmPassword, setConfirmPassword] = React.useState('');
  const [passwordSuccess, setPasswordSuccess] = React.useState(false);
  const [passwordError, setPasswordError] = React.useState<string | null>(null);

  const [logoutAll, { isLoading: isLoggingOutAll }] = useLogoutAllMutation();
  const [logoutAllSuccess, setLogoutAllSuccess] = React.useState(false);

  async function handlePasswordSubmit(e: React.FormEvent) {
    e.preventDefault();
    setPasswordError(null);
    setPasswordSuccess(false);

    if (newPassword.length < 8) {
      setPasswordError('New password must be at least 8 characters');
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordError('New passwords do not match');
      return;
    }

    try {
      await changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      }).unwrap();
      setPasswordSuccess(true);
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      setTimeout(() => setPasswordSuccess(false), 4000);
    } catch (err: unknown) {
      const apiErr = err as { data?: { error?: string } };
      setPasswordError(
        apiErr?.data?.error || 'Failed to change password. Check your current password.',
      );
    }
  }

  async function handleLogoutAll() {
    try {
      await logoutAll().unwrap();
      setLogoutAllSuccess(true);
    } catch {
      // fallback
    }
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Change Password</CardTitle>
          <CardDescription>
            Ensure your account is using a long, random password to stay secure.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handlePasswordSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="current-pass" className="text-xs font-semibold">
                Current Password
              </Label>
              <Input
                id="current-pass"
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                placeholder="Enter current password"
                required
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="new-pass" className="text-xs font-semibold">
                New Password
              </Label>
              <Input
                id="new-pass"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder="Minimum 8 characters"
                minLength={8}
                required
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="confirm-pass" className="text-xs font-semibold">
                Confirm New Password
              </Label>
              <Input
                id="confirm-pass"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="Re-enter new password"
                required
              />
            </div>

            {passwordSuccess && (
              <div className="flex items-center gap-2 rounded-lg bg-emerald-500/10 p-3 text-xs font-medium text-emerald-600 dark:text-emerald-400">
                <CheckCircle className="size-4 shrink-0" weight="fill" />
                <span>Password updated successfully!</span>
              </div>
            )}

            {passwordError && (
              <div className="bg-destructive/10 text-destructive flex items-center gap-2 rounded-lg p-3 text-xs font-medium">
                <WarningCircle className="size-4 shrink-0" weight="fill" />
                <span>{passwordError}</span>
              </div>
            )}

            <div className="flex justify-end pt-2">
              <Button type="submit" disabled={isChangingPassword}>
                {isChangingPassword ? 'Updating...' : 'Change Password'}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Session Security</CardTitle>
          <CardDescription>Sign out of all active sessions and other devices.</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-center">
          <div>
            <p className="text-foreground text-sm font-medium">Log out all other sessions</p>
            <p className="text-muted-foreground mt-0.5 text-xs">
              Invalidates all refresh tokens and sessions except your current one.
            </p>
            {logoutAllSuccess && (
              <p className="mt-2 text-xs font-medium text-emerald-500">
                ✓ All other sessions have been logged out.
              </p>
            )}
          </div>
          <Button
            type="button"
            variant="destructive"
            size="sm"
            onClick={handleLogoutAll}
            disabled={isLoggingOutAll}
            className="shrink-0 gap-2"
          >
            <SignOut className="size-4" />
            <span>{isLoggingOutAll ? 'Logging out...' : 'Log out all devices'}</span>
          </Button>
        </CardContent>
      </Card>
    </>
  );
}

function AppearanceSettingsCard() {
  const { dark, toggle: toggleTheme } = useTheme();

  return (
    <Card>
      <CardHeader>
        <CardTitle>Theme Preference</CardTitle>
        <CardDescription>Select your preferred interface color mode for TownHall.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <button
            type="button"
            onClick={() => {
              if (dark) toggleTheme();
            }}
            className={`flex cursor-pointer flex-col items-center justify-center gap-3 rounded-2xl border-2 p-5 transition-all ${
              !dark
                ? 'border-primary bg-primary/5 text-primary shadow-xs'
                : 'border-border bg-card text-muted-foreground hover:border-primary/50'
            }`}
          >
            <div className="flex size-12 items-center justify-center rounded-xl bg-amber-500/10 text-amber-500">
              <Sun className="size-6" weight={!dark ? 'fill' : 'regular'} />
            </div>
            <div className="text-center">
              <p className="text-foreground text-sm font-semibold">Light Mode</p>
              <p className="text-muted-foreground text-xs">Crisp clean light background</p>
            </div>
          </button>

          <button
            type="button"
            onClick={() => {
              if (!dark) toggleTheme();
            }}
            className={`flex cursor-pointer flex-col items-center justify-center gap-3 rounded-2xl border-2 p-5 transition-all ${
              dark
                ? 'border-primary bg-primary/5 text-primary shadow-xs'
                : 'border-border bg-card text-muted-foreground hover:border-primary/50'
            }`}
          >
            <div className="flex size-12 items-center justify-center rounded-xl bg-indigo-500/10 text-indigo-400">
              <Moon className="size-6" weight={dark ? 'fill' : 'regular'} />
            </div>
            <div className="text-center">
              <p className="text-foreground text-sm font-semibold">Dark Mode</p>
              <p className="text-muted-foreground text-xs">Eye-friendly deep dark theme</p>
            </div>
          </button>
        </div>

        <div className="border-border bg-muted/40 text-muted-foreground flex items-center justify-between rounded-xl border p-4 text-xs">
          <span>Theme tokens configured via Tailwind CSS v4 and OKLCH color spaces.</span>
          <span className="text-foreground flex items-center gap-1 font-semibold">
            Active: {dark ? 'Dark' : 'Light'} <ArrowRight className="size-3" />
          </span>
        </div>
      </CardContent>
    </Card>
  );
}

export default function SettingsPage() {
  const { data: user, isLoading: isUserLoading } = useGetMeQuery();

  if (isUserLoading || !user) {
    return (
      <div className="mx-auto flex h-64 max-w-2xl items-center justify-center">
        <Spinner className="text-primary size-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="mx-auto w-full max-w-3xl pb-16">
      <Tabs defaultValue="profile" className="w-full space-y-6">
        <TabsList className="[scrollbar-width:none] justify-start gap-1 overflow-x-auto [&::-webkit-scrollbar]:hidden">
          <TabsTrigger
            value="profile"
            className="data-[active]:text-foreground text-base data-[active]:font-bold"
          >
            Profile
          </TabsTrigger>
          <TabsTrigger
            value="account"
            className="data-[active]:text-foreground text-base data-[active]:font-bold"
          >
            Account
          </TabsTrigger>
          <TabsTrigger
            value="security"
            className="data-[active]:text-foreground text-base data-[active]:font-bold"
          >
            Security
          </TabsTrigger>
          <TabsTrigger
            value="appearance"
            className="data-[active]:text-foreground text-base data-[active]:font-bold"
          >
            Appearance
          </TabsTrigger>
          <TabsIndicator className="h-1 rounded-full" />
        </TabsList>

        <TabsContent value="profile" className="space-y-6">
          <ProfileSettingsForm key={user.id} user={user} />
        </TabsContent>

        <TabsContent value="account" className="space-y-6">
          <AccountSettingsForm key={user.id} user={user} />
        </TabsContent>

        <TabsContent value="security" className="space-y-6">
          <SecuritySettingsForm />
        </TabsContent>

        <TabsContent value="appearance" className="space-y-6">
          <AppearanceSettingsCard />
        </TabsContent>
      </Tabs>
    </div>
  );
}
