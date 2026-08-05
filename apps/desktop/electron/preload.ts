import { contextBridge, ipcRenderer } from "electron";

contextBridge.exposeInMainWorld("electron", {
  connectLocal() {
    return ipcRenderer.invoke("connect-local");
  },

  connectRemote(url: string) {
    if (!/^https?:\/\//i.test(url)) {
      url = `http://${url}`;
    }

    const parsed = new URL(url);

    return ipcRenderer.invoke("connect-remote", parsed.toString());
  },
});
