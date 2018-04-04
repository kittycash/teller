'use strict'

const { app, Menu, BrowserWindow, dialog } = require('electron');

const userDataPath = app.getPath('userData');

const cxo_dir = userDataPath + "/cxo";

var log = require('electron-log');

var fs = require('fs');
var util = require('util');

const path = require('path');

const childProcess = require('child_process');

const cwd = require('process').cwd();

// This adds refresh and devtools console keybindings
// Page can refresh with cmd+r, ctrl+r, F5
// Devtools can be toggled with cmd+alt+i, ctrl+shift+i, F12
require('electron-debug')({enabled: true, showDevTools: false});
require('electron-context-menu')({});

global.eval = function() { throw new Error('bad!!'); }

const defaultURL = 'http://127.0.0.1:6148/boxes';

//WARNING - Re-enable this after testing!  Need to figure out all the recaptcha urls
// Force everything localhost, in case of a leak
// app.commandLine.appendSwitch('host-rules', 'MAP * 127.0.0.1, EXCLUDE www.google.com, fonts.gstatic.com, www.gstatic.com, fonts.googleapis.com');
app.commandLine.appendSwitch('--enable-viewport-meta', 'true');
app.commandLine.appendSwitch('ssl-version-fallback-min', 'tls1.2');
app.commandLine.appendSwitch('--no-proxy-server');
app.setAsDefaultProtocolClient('kittycash');

// Keep a global reference of the window object, if you don't, the window will
// be closed automatically when the JavaScript object is garbage collected.
let win;

var kittycash = null;

function startKittyCash() {

 
  log.info('Starting KittyCash from electron');

  if (kittycash) {
    log.info('KittyCash already running');
    app.emit('kittycash-ready');
    return
  }

  var reset = () => {
    kittycash = null;
  }

  // Resolve kittycash wallet binary location
  var appPath = app.getPath('exe');
  var exe = (() => {

  if (isDev())
  {
    return 'go';
  }

  switch (process.platform) {
    case 'darwin':
      return path.join(appPath, '../../Resources/app/wallet');
    case 'win32':
      // Use only the relative path on windows due to short path length
      // limits
      return './resources/app/wallet.exe';
    case 'linux':
      return path.join(path.dirname(appPath), './resources/app/wallet');
    default:
      return './resources/app/wallet';
  }
})()

  var gui_dir = (() => {

   if (isDev())
    {
      return '../tabs/dist';
    }
  switch (process.platform) {
    case 'darwin':
      return path.join(appPath, '../../Resources/app/dist');
    case 'win32':
      // Use only the relative path on windows due to short path length
      // limits
      return './resources/app/dist';
    case 'linux':
      return path.join(path.dirname(appPath), './resources/app/dist');
    default:
      return './resources/app/dist';
  }
})()

  var args = [
    '--test=true',
    '--test-gen-pk=03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753',
    '--test-root-pk=03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753',
    '--test-root-nonce=1234',
    '--cxo-address=127.0.0.1:6140',
    '--http-address=127.0.0.1:6148',
    '--messenger-addresses=:8880',
    '--cxo-dir=' + cxo_dir,
    '--gui=true',
    '--gui-dir=' + gui_dir
  ];

  if (isDev())
  {
    args.unshift("../cmd/wallet/wallet.go");
    args.unshift("run");
  }

  kittycash = childProcess.spawn(exe, args);

  kittycash.on('error', (e) => {
    dialog.showErrorBox('Failed to start kittycash', e.toString());
    app.quit();
  });

  //WARNING - for some reason everything is coming out as stderr instead of stdout
  kittycash.stderr.on('data', (data) => {
    // log.info(data.toString());
    console.log(data.toString());
    app.emit('kittycash-ready', { url: defaultURL });
});

//   kittycash.stderr.on('data', (data) => {
//     console.log(data.toString());
// });

  kittycash.on('close', (code) => {
    log.info('KittyCash closed');
    reset();
  });

  kittycash.on('exit', (code) => {
    log.info('KittyCash exited');
    reset();
  });
}


function createWindow(url) {

  if (!url) {
    url = defaultURL;
  }

  // To fix appImage doesn't show icon in dock issue.
  var appPath = app.getPath('exe');
  var iconPath = (() => {
    switch (process.platform) {
      case 'linux':
        return path.join(path.dirname(appPath), './resources/icon512x512.png');
    }
  })()

  // Create the browser window.
  win = new BrowserWindow({
    width: 1200,
    height: 900,
    title: 'KittyCash',
    icon: iconPath,
    nodeIntegration: false,
    webPreferences: {
      webgl: false,
      webaudio: false,
    },
  });

  let webContents = win.webContents;

  webContents.on('did-finish-load', () => {
    webContents.setZoomFactor(1);
    webContents.setVisualZoomLevelLimits(1, 1);
    webContents.setLayoutZoomLevelLimits(0, 0);
  });

  // patch out eval
  win.eval = global.eval;

  const ses = win.webContents.session
  ses.clearCache(function () {
    log.info('Cleared the caching of the kittycash wallet.');
  });

  //Don't clear out the localstorage, because we use it!
  // ses.clearStorageData([],function(){
  //   log.info('Cleared the stored cached data');
  // });

  win.loadURL(url);

  // Open the DevTools.
  // win.webContents.openDevTools();

  // Emitted when the window is closed.
  win.on('closed', () => {
    // Dereference the window object, usually you would store windows
    // in an array if your app supports multi windows, this is the time
    // when you should delete the corresponding element.
    win = null;
});

  // create application's main menu
  var template = [{
    label: "KittyCash",
    submenu: [
      { label: "About KittyCash", selector: "orderFrontStandardAboutPanel:", click: function() 
        { 
          var appVersion = require('electron').app.getVersion();
          dialog.showMessageBox({ type: 'info', title: 'About KittyCash', message: 'KittyCash version: ' + appVersion });
        }
      },
      { label: "Quit", accelerator: "Command+Q", click: function() { app.quit(); } }
    ]
  }, {
    label: "Edit",
    submenu: [
      { label: "Undo", accelerator: "CmdOrCtrl+Z", selector: "undo:" },
      { label: "Redo", accelerator: "Shift+CmdOrCtrl+Z", selector: "redo:" },
      { type: "separator" },
      { label: "Cut", accelerator: "CmdOrCtrl+X", selector: "cut:" },
      { label: "Copy", accelerator: "CmdOrCtrl+C", selector: "copy:" },
      { label: "Paste", accelerator: "CmdOrCtrl+V", selector: "paste:" },
      { label: "Select All", accelerator: "CmdOrCtrl+A", selector: "selectAll:" }
    ]
  }];

  Menu.setApplicationMenu(Menu.buildFromTemplate(template));
}

// Enforce single instance
const alreadyRunning = app.makeSingleInstance((commandLine, workingDirectory) => {
      // Someone tried to run a second instance, we should focus our window.
      if (win) {
        if (win.isMinimized()) {
          win.restore();
        }
        win.focus();
      } else {
        createWindow(defaultURL);
}
});

if (alreadyRunning) {
  app.quit();
  return;
}

// This method will be called when Electron has finished
// initialization and is ready to create browser windows.
// Some APIs can only be used after this event occurs.
app.on('ready', startKittyCash);

var kittyCashLoaded = false;
app.on('kittycash-ready', (e) => {
  if (!kittyCashLoaded)
  {
    kittyCashLoaded = true;
    setTimeout(function() {
      createWindow(e.url);
    }, 1000);
  }

});

// Quit when all windows are closed.
app.on('window-all-closed', () => {
  // On OS X it is common for applications and their menu bar
  // to stay active until the user quits explicitly with Cmd + Q
  if (process.platform !== 'darwin') {
  app.quit();
}
});

app.on('activate', () => {
  // On OS X it's common to re-create a window in the app when the
  // dock icon is clicked and there are no other windows open.
  if (win === null) {
  createWindow();
}
});

app.on('will-quit', () => {
  if (kittycash) {
    kittycash.kill('SIGINT');
  }
});


function isDev() {
  return process.mainModule.filename.indexOf('app.asar') === -1;
}

