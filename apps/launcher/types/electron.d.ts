export {};

declare global {
  interface Window {
    electron: {
      connectLocal(): Promise<void>;
      connectRemote(url: string): Promise<void>;
    };
  }
}
