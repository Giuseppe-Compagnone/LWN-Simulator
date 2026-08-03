import { contextBridge, ipcRenderer } from "electron";

contextBridge.exposeInMainWorld("electron", {
  connectLocal() {
    return ipcRenderer.invoke("connect-local");
  },

  connectRemote(url: string) {
    return ipcRenderer.invoke("connect-remote", url);
  },
});
