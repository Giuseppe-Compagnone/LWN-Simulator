import {
  CreateProfileRequest,
  Profile,
  ProfileArchive,
  UpdateProfileRequest,
} from "@lwn-simulator/contracts";
import { PropsWithChildren } from "react";

export interface ProfileServiceContent {
  profiles: Array<Profile>;
  activeProfile: Profile | null;
  activeProfileID: string;
  profileBaseUrl: string;
  loading: boolean;
  error: Error | null;
  selectProfile: (id: string) => void;
  createProfile: (request: CreateProfileRequest) => Promise<Profile>;
  updateProfile: (request: UpdateProfileRequest) => Promise<Profile>;
  deleteProfile: (id: string) => Promise<void>;
  exportProfile: (id?: string) => Promise<ProfileArchive>;
  importProfile: (archive: ProfileArchive) => Promise<Profile>;
}

export interface ProfileServiceProviderProps extends PropsWithChildren {
  baseUrl: string;
}
