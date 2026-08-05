import { Menu, MenuItemConstructorOptions, shell } from "electron";
import { getMainWindow } from "./window-manager";

export const createMenu = () => {
  const menuTemplate: MenuItemConstructorOptions[] = [
    {
      label: "File",
      submenu: [
        {
          id: "disconnect",
          label: "Disconnect",
          enabled: false,
          click: async () => {
            const window = getMainWindow();

            if (window) {
              await window.webContents.executeJavaScript(
                "window.electron.disconnect()",
              );
            }
          },
        },
        {
          label: "Quit",
          role: "quit",
        },
      ],
    },
    {
      label: "View",
      submenu: [
        {
          label: "DevTools",
          role: "toggleDevTools",
        },
      ],
    },
    {
      label: "Help",
      submenu: [
        {
          label: "Documentation",
          click: async () => {
            await shell.openExternal(
              "https://giuseppe-compagnone.github.io/LWN-Simulator/docs",
            );
          },
        },
        {
          label: "Comunity Discussions",
          click: async () => {
            await shell.openExternal(
              "https://github.com/Giuseppe-Compagnone/LWN-Simulator/discussions",
            );
          },
        },
        {
          label: "Search Issues",
          click: async () => {
            await shell.openExternal(
              "https://github.com/Giuseppe-Compagnone/LWN-Simulator/issues",
            );
          },
        },
      ],
    },
  ];

  const menu = Menu.buildFromTemplate(menuTemplate);
  Menu.setApplicationMenu(menu);

  return menu.getMenuItemById("disconnect");
};
