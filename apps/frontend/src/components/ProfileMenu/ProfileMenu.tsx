"use client";

import {
  Button,
  ButtonType,
  NotificationHandler,
} from "@lwn-simulator/ui-components";
import { useProfileService } from "@lwn-simulator/sdk";
import { useRef, useState } from "react";

const readFile = (file: File): Promise<string> =>
  new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result));
    reader.onerror = () => reject(new Error("Unable to read the profile file"));
    reader.readAsText(file);
  });

const ProfileMenu = () => {
  const profileService = useProfileService();
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const notification = NotificationHandler.instance;

  const runAction = async (action: () => Promise<void>, success: string) => {
    setBusy(true);
    try {
      await action();
      notification.success(success);
    } catch (error) {
      notification.error(
        error instanceof Error ? error.message : "The profile action failed",
      );
    } finally {
      setBusy(false);
    }
  };

  const createProfile = () => {
    const name = window.prompt("Profile name");
    if (!name?.trim()) return;
    void runAction(
      () => profileService.createProfile({ name: name.trim() }).then(() => undefined),
      "Profile created",
    );
  };

  const renameProfile = () => {
    if (!profileService.activeProfile) return;
    const name = window.prompt("New profile name", profileService.activeProfile.name);
    if (!name?.trim()) return;
    void runAction(
      () =>
        profileService
          .updateProfile({ id: profileService.activeProfileID, name: name.trim() })
          .then(() => undefined),
      "Profile renamed",
    );
  };

  const deleteProfile = () => {
    if (!profileService.activeProfile) return;
    if (
      !window.confirm(
        `Delete profile “${profileService.activeProfile.name}”? This also removes its devices, gateways and logs.`,
      )
    ) {
      return;
    }
    void runAction(
      () => profileService.deleteProfile(profileService.activeProfileID),
      "Profile deleted",
    );
  };

  const exportProfile = () => {
    void runAction(
      async () => {
        const archive = await profileService.exportProfile();
        const blob = new Blob([JSON.stringify(archive, null, 2)], {
          type: "application/json",
        });
        const url = URL.createObjectURL(blob);
        const anchor = document.createElement("a");
        anchor.href = url;
        anchor.download = `${profileService.activeProfile?.name ?? "lwn-profile"}.lwn-profile.json`;
        anchor.click();
        URL.revokeObjectURL(url);
      },
      "Profile exported",
    );
  };

  const importProfile = () => {
    fileInputRef.current?.click();
  };

  const handleFile = async (file: File | undefined) => {
    if (!file) return;
    await runAction(
      async () => {
        const archive = JSON.parse(await readFile(file));
        await profileService.importProfile(archive);
      },
      "Profile imported",
    );
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  return (
    <div className="profile-menu">
      <button
        className="profile-menu-trigger"
        type="button"
        aria-expanded={open}
        aria-label="Select profile"
        onClick={() => setOpen((value) => !value)}
        disabled={busy || profileService.loading}
      >
        <span className="material-symbols-outlined icon">account_tree</span>
        <span className="profile-menu-name">
          {profileService.activeProfile?.name ?? "Profiles"}
        </span>
        <span className="material-symbols-outlined expand-icon">
          expand_more
        </span>
      </button>
      {open && (
        <div className="profile-menu-panel">
          <label htmlFor="active-profile">Active profile</label>
          <select
            id="active-profile"
            value={profileService.activeProfileID}
            onChange={(event) => {
              profileService.selectProfile(event.target.value);
              setOpen(false);
            }}
          >
            {profileService.profiles.map((profile) => (
              <option key={profile.id} value={profile.id}>
                {profile.name}
              </option>
            ))}
          </select>
          <div className="profile-menu-actions">
            <Button
              value="New profile"
              type={ButtonType.Outlined}
              onClick={createProfile}
              disabled={busy}
            />
            <Button
              value="Rename"
              type={ButtonType.Outlined}
              onClick={renameProfile}
              disabled={busy || !profileService.activeProfile}
            />
            <Button
              value="Delete"
              type={ButtonType.Outlined}
              onClick={deleteProfile}
              disabled={busy || profileService.profiles.length <= 1}
            />
            <Button
              value="Export"
              type={ButtonType.Outlined}
              onClick={exportProfile}
              disabled={busy || !profileService.activeProfile}
            />
            <Button
              value="Import"
              type={ButtonType.Outlined}
              onClick={importProfile}
              disabled={busy}
            />
          </div>
          <input
            ref={fileInputRef}
            className="profile-menu-file"
            type="file"
            accept="application/json,.json"
            onChange={(event) => void handleFile(event.target.files?.[0])}
          />
        </div>
      )}
    </div>
  );
};

export default ProfileMenu;
