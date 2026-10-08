"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  CreateProfileRequest,
  Profile,
  ProfileArchive,
  RealtimeWebSocketMessage,
  RealtimeWebSocketMessageType,
  UpdateProfileRequest,
} from "@lwn-simulator/contracts";
import { useWebSocket } from "../../websocket";
import ProfileServiceContext from "./ProfileServiceContext";
import {
  ProfileServiceContent,
  ProfileServiceProviderProps,
} from "./ProfileService.types";
import { ProfileService } from "./ProfileService";

const activeProfileStorageKey = "lwn-simulator.active-profile";
const defaultProfileID = "00000000-0000-4000-8000-000000000001";

const readStoredProfileID = (): string => {
  if (typeof window === "undefined") return defaultProfileID;
  return window.localStorage.getItem(activeProfileStorageKey) ?? defaultProfileID;
};

const ProfileServiceProvider = (props: ProfileServiceProviderProps) => {
  const [profiles, setProfiles] = useState<Array<Profile>>([]);
  const [activeProfileID, setActiveProfileID] = useState(readStoredProfileID);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const profilesRef = useRef<Array<Profile>>([]);
  const activeProfileIDRef = useRef(activeProfileID);
  const service = useMemo(
    () => new ProfileService(props.baseUrl),
    [props.baseUrl],
  );
  const { subscribe } = useWebSocket();

  const commitProfiles = useCallback((nextProfiles: Array<Profile>) => {
    profilesRef.current = nextProfiles;
    setProfiles(nextProfiles);
  }, []);

  const persistActiveProfile = useCallback((id: string) => {
    activeProfileIDRef.current = id;
    setActiveProfileID(id);
    if (typeof window !== "undefined") {
      window.localStorage.setItem(activeProfileStorageKey, id);
    }
  }, []);

  const selectProfile = useCallback(
    (id: string) => {
      if (!profilesRef.current.some((profile) => profile.id === id)) return;
      persistActiveProfile(id);
    },
    [persistActiveProfile],
  );

  const getProfiles = useCallback(async () => {
    setLoading(true);
    try {
      const nextProfiles = await service.getProfiles();
      commitProfiles(nextProfiles);
      const storedID = activeProfileIDRef.current;
      const nextActive = nextProfiles.some((profile) => profile.id === storedID)
        ? storedID
        : nextProfiles.find((profile) => profile.id === defaultProfileID)?.id ??
          nextProfiles[0]?.id ??
          "";
      if (nextActive) persistActiveProfile(nextActive);
      setError(null);
    } catch (loadError) {
      setError(
        loadError instanceof Error
          ? loadError
          : new Error("Failed to load profiles"),
      );
    } finally {
      setLoading(false);
    }
  }, [commitProfiles, persistActiveProfile, service]);

  useEffect(() => {
    void Promise.resolve().then(getProfiles);
  }, [getProfiles]);

  const handleRealtimeMessage = useCallback(
    (rawMessage: string) => {
      try {
        const message = JSON.parse(rawMessage) as RealtimeWebSocketMessage;
        if (
          (message.type === RealtimeWebSocketMessageType.ProfileCreated ||
            message.type === RealtimeWebSocketMessageType.ProfileUpdated) &&
          message.profile
        ) {
          const nextProfile = message.profile as Profile;
          const nextProfiles = profilesRef.current.some(
            (profile) => profile.id === nextProfile.id,
          )
            ? profilesRef.current.map((profile) =>
                profile.id === nextProfile.id ? nextProfile : profile,
              )
            : [...profilesRef.current, nextProfile];
          commitProfiles(nextProfiles);
        }
        if (
          message.type === RealtimeWebSocketMessageType.ProfileDeleted &&
          message.resourceID
        ) {
          const nextProfiles = profilesRef.current.filter(
            (profile) => profile.id !== message.resourceID,
          );
          commitProfiles(nextProfiles);
          if (
            activeProfileIDRef.current === message.resourceID &&
            nextProfiles[0]
          ) {
            persistActiveProfile(nextProfiles[0].id);
          }
        }
      } catch {
        // Other realtime consumers may share this websocket stream.
      }
    },
    [commitProfiles, persistActiveProfile],
  );

  useEffect(
    () => subscribe(handleRealtimeMessage),
    [handleRealtimeMessage, subscribe],
  );

  const createProfile = useCallback(
    async (request: CreateProfileRequest) => {
      try {
        const profile = await service.createProfile(request);
        const nextProfiles = [...profilesRef.current, profile];
        commitProfiles(nextProfiles);
        persistActiveProfile(profile.id);
        setError(null);
        return profile;
      } catch (requestError) {
        setError(requestError instanceof Error ? requestError : new Error("Failed to create profile"));
        throw requestError;
      }
    },
    [commitProfiles, persistActiveProfile, service],
  );

  const updateProfile = useCallback(
    async (request: UpdateProfileRequest) => {
      try {
        const profile = await service.updateProfile(request);
        commitProfiles(
          profilesRef.current.map((current) =>
            current.id === profile.id ? profile : current,
          ),
        );
        setError(null);
        return profile;
      } catch (requestError) {
        setError(requestError instanceof Error ? requestError : new Error("Failed to update profile"));
        throw requestError;
      }
    },
    [commitProfiles, service],
  );

  const deleteProfile = useCallback(
    async (id: string) => {
      await service.deleteProfile(id);
      const nextProfiles = profilesRef.current.filter(
        (profile) => profile.id !== id,
      );
      commitProfiles(nextProfiles);
      if (activeProfileIDRef.current === id && nextProfiles[0]) {
        persistActiveProfile(nextProfiles[0].id);
      }
      setError(null);
    },
    [commitProfiles, persistActiveProfile, service],
  );

  const exportProfile = useCallback(
    (id = activeProfileIDRef.current): Promise<ProfileArchive> =>
      service.exportProfile(id),
    [service],
  );

  const importProfile = useCallback(
    async (archive: ProfileArchive) => {
      const profile = await service.importProfile(archive);
      const nextProfiles = [...profilesRef.current, profile];
      commitProfiles(nextProfiles);
      persistActiveProfile(profile.id);
      setError(null);
      return profile;
    },
    [commitProfiles, persistActiveProfile, service],
  );

  const activeProfile =
    profiles.find((profile) => profile.id === activeProfileID) ?? null;
  const profileBaseUrl = `${props.baseUrl.replace(/\/+$/, "")}/profiles/${encodeURIComponent(activeProfileID)}`;
  const value = useMemo(
    (): ProfileServiceContent => ({
      profiles,
      activeProfile,
      activeProfileID,
      profileBaseUrl,
      loading,
      error,
      selectProfile,
      createProfile,
      updateProfile,
      deleteProfile,
      exportProfile,
      importProfile,
    }),
    [
      profiles,
      activeProfile,
      activeProfileID,
      profileBaseUrl,
      loading,
      error,
      selectProfile,
      createProfile,
      updateProfile,
      deleteProfile,
      exportProfile,
      importProfile,
    ],
  );

  return (
    <ProfileServiceContext.Provider value={value}>
      {props.children}
    </ProfileServiceContext.Provider>
  );
};

export default ProfileServiceProvider;
