const storage = {
  get(key: string) {
    if (window.electron) {
      return window.electron.getStorage(key);
    }

    return localStorage.getItem(key);
  },

  set(key: string, value: string) {
    localStorage.setItem(key, value);

    window.electron?.syncStorage(key, value);
  },
};

export default storage;
