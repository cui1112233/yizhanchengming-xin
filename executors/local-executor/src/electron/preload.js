const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('yiZhanExecutor', {
  getState: () => ipcRenderer.invoke('executor:get-state'),
  pair: input => ipcRenderer.invoke('executor:pair', input),
  addAccount: () => ipcRenderer.invoke('executor:add-account'),
  openAccount: id => ipcRenderer.invoke('executor:open-account', id),
  markAccountAvailable: id => ipcRenderer.invoke('executor:mark-account-available', id),
  setAutomationEnabled: enabled => ipcRenderer.invoke('executor:set-automation-enabled', Boolean(enabled)),
  onState: callback => {
    const listener = (_event, state) => callback(state);
    ipcRenderer.on('executor:state', listener);
    return () => ipcRenderer.removeListener('executor:state', listener);
  },
  updater: {
    getState: () => ipcRenderer.invoke('updater:get-state'),
    check: () => ipcRenderer.invoke('updater:check'),
    setChannel: channel => ipcRenderer.invoke('updater:set-channel', channel),
    install: () => ipcRenderer.invoke('updater:install'),
    onState: callback => {
      const listener = (_event, state) => callback(state);
      ipcRenderer.on('updater:state', listener);
      return () => ipcRenderer.removeListener('updater:state', listener);
    }
  }
});
