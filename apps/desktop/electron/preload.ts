import { contextBridge, ipcRenderer } from "electron";

async function connectLocal() {
  return ipcRenderer.invoke("connect-local");
}

async function connectRemote(url: string) {
  if (!/^https?:\/\//i.test(url)) {
    url = `http://${url}`;
  }

  return ipcRenderer.invoke("connect-remote", new URL(url).toString());
}

contextBridge.exposeInMainWorld("electron", {
  connectLocal,
  connectRemote,
});
