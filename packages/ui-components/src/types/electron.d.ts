export {};

declare global {
  interface Window {
    electron: {
      connectLocal(): Promise<void>;

      connectRemote(url: string): Promise<{
        success: boolean;
        message?: string;
      }>;

      disconnect(): Promise<void>;

      getStorage(key: string): Promise<string | undefined>;

      syncStorage(key: string, value: string): Promise<void>;
    };
  }
}
